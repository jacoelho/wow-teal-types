-- Stock Teal 0.24.8 checks ordinary types. This pass enforces the secrecy rules
-- it does not: operations on Secret values need same-value native proofs, `as`
-- is the only unwrap, and Secret positions must not reach plain positions.
--
-- One run checks every selected source against one declaration load. Each
-- source gets its own Teal environment over that load, so its globals and
-- authored modules stay its own, as in separate `tl check` runs, and a failure
-- in one source does not stop the checks of the others.

-- `tl run` has already loaded the compiler. A `require("tl")` here would make
-- `tl run` type-check all of tl.lua before this script starts.
local tl = package.loaded["tl"]
if type(tl) ~= "table" or type(tl.version) ~= "function" then
   io.stderr:write("check_secrets.lua: run with `tl run`\n")
   os.exit(2)
end
local environment, native_environment = arg[1], arg[2]
local sources = {}
for index = 3, #arg do sources[#sources + 1] = arg[index] end
if not environment or not native_environment or #sources == 0 then
   io.stderr:write("usage: tl [include flags] run check_secrets.lua -- ENV NATIVE_ENV SOURCE...\n")
   os.exit(2)
end
-- A failure of the run itself fails every selected source closed.
local function fail_run(message)
   for _, source in ipairs(sources) do
      io.stderr:write(source .. ":1:1: secret check failed closed: " .. tostring(message) .. "\n")
   end
   os.exit(1)
end
local TEAL_VERSION = "0.24.8"
if tl.version() ~= TEAL_VERSION then
   fail_run("unsupported Teal version " .. tostring(tl.version()) .. "; expected " .. TEAL_VERSION)
end
if native_environment ~= "wow_env" and native_environment ~= "wow_secure_env" then
   fail_run("unsupported native environment")
end
-- A failure of one source: it is reported failed closed and the run continues.
-- The marker passes through the pcalls that wrap analysis phases.
local FAILED_CLOSED = {}
local function fail_closed(message)
   if type(message) == "table" and message[FAILED_CLOSED] then error(message, 0) end
   error({ [FAILED_CLOSED] = tostring(message) }, 0)
end

local function canonical(path)
   if path:sub(1, 1) ~= "/" then path = (os.getenv("PWD") or ".") .. "/" .. path end
   local pieces = {}
   for piece in path:gmatch("[^/]+") do
      if piece == ".." then table.remove(pieces)
      elseif piece ~= "." then pieces[#pieces + 1] = piece end
   end
   return "/" .. table.concat(pieces, "/")
end
local script = canonical(arg[0])
local root = script:match("^(.*)/tools/check_secrets.lua$")
if not root then fail_run("cannot locate catalogue root from checker path") end

-- Call policy: native secret-accepting positions, keyed by the declaring
-- function type's location in the generated declarations.
local loaded_policy, call_policy = pcall(dofile, root .. "/types/wow/call_policy.lua")
if not loaded_policy or type(call_policy) ~= "table" or call_policy.schema ~= 1
   or call_policy.compiler ~= TEAL_VERSION or type(call_policy.declarations) ~= "table" then
   fail_run("missing or incompatible call policy")
end
local guard_symbols = { canaccessvalue = "value", canaccessallvalues = "values", canaccesstable = "table", issecretvalue = "secret" }
local function valid_position(value)
   return type(value) == "number" and value % 1 == 0 and value >= 1 and value <= 256
end
local function validate_positions(list)
   if type(list) ~= "table" then error("invalid secret-parameter policy") end
   local seen = {}
   for key, value in pairs(list) do
      if type(key) ~= "number" or key % 1 ~= 0 or key < 1 or key > #list or not valid_position(value) or seen[value] then
         error("invalid secret-parameter policy")
      end
      seen[value] = true
   end
end
local policy_ok, policy_error = pcall(function()
   local count = 0
   for path, declarations in pairs(call_policy.declarations) do
      if path ~= "types/wow/wow.d.tl" and path ~= "types/wow/wow_env.d.tl" and path ~= "types/wow/wow_secure_env.d.tl" then error("invalid call policy source") end
      if type(declarations) ~= "table" then error("invalid call policy declarations") end
      for location, policy in pairs(declarations) do
         count = count + 1
         if count > 100000 or type(location) ~= "string" or not location:match("^[1-9]%d*:[1-9]%d*$")
            or type(policy) ~= "table" or type(policy.symbol) ~= "string" or #policy.symbol > 256 then error("invalid call policy identity") end
         validate_positions(policy.secret_parameters)
         -- Contents secrecy is carried by SecretTable types, not by policy.
         if policy.secret_contents_parameters ~= nil or policy.secret_contents_vararg_from ~= nil then error("unsupported secret-contents policy") end
         if policy.secret_vararg_from ~= nil and not valid_position(policy.secret_vararg_from) then error("invalid call policy vararg") end
         if policy.guard and guard_symbols[policy.symbol] ~= policy.guard then error("invalid native guard policy") end
      end
   end
end)
if not policy_ok then fail_run(policy_error) end

local syntax_fields = { "e1", "e2", "vars", "exps", "body", "args", "name", "fn_owner", "if_blocks", "exp", "var", "from", "to", "step", "key", "value", "macrodef" }
local function syntax_children(node, consume)
   for _, child in ipairs(node) do if type(child) == "table" then consume(child) end end
   for _, field in ipairs(syntax_fields) do
      local child = node[field]
      if type(child) == "table" then consume(child) end
   end
end
-- Node-to-type mapping. Teal's collector receives (y, x, type) only, and several
-- nodes share a position, so the stored node is read from the storing frame:
-- the post-visit hook's `n`, a call's `e1`, and declaration variables. A node
-- that receives two different types is ambiguous and fails closed on lookup.
local AMBIGUOUS = {}
local function record(map, node, typ)
   local previous = map[node]
   if previous == nil then map[node] = typ
   elseif previous ~= typ then map[node] = AMBIGUOUS end
end
local function frame_locals()
   local names, values = {}, {}
   for index = 1, 64 do
      local name, value = debug.getlocal(3, index)
      if not name then break end
      names[index], values[name] = name, value
   end
   return names, values
end
local function is_node_at(value, y, x)
   return type(value) == "table" and value.kind ~= nil and value.y == y and value.x == x
end

-- Marker identity must belong to the declarations actually loaded by Teal.
-- Returns the first declaration module loaded from elsewhere, or required by
-- this check and not loaded.
local function declaration_mismatch(env, require_loaded)
   for _, name in ipairs({ "wow", "native_values", "wow_env", "wow_secure_env" }) do
      local actual = env.module_filenames[name]
      local required = require_loaded and (name == "native_values" or name == native_environment)
      if (required and not actual) or (env.modules[name] and not actual)
         or (actual and canonical(actual) ~= root .. "/types/wow/" .. name .. ".d.tl") then
         return name
      end
   end
end

local opts = { defaults = { feat_lax = "off", feat_arity = "on", gen_compat = "off", gen_target = "5.1" } }
if environment ~= "-" then opts.predefined_modules = { environment } end
local declarations, env_error = tl.new_env(opts)
if not declarations then fail_run(env_error) end

-- Prints Teal's diagnostics for the files env loaded from position `first`
-- on, and reports whether any is an error.
local function report_stock(env, first)
   local failed = false
   for index = first, #env.loaded_order do
      local loaded_name = env.loaded_order[index]
      local loaded = env.loaded[loaded_name]
      for _, category in ipairs({ "syntax_errors", "type_errors", "warnings" }) do
         for _, diagnostic in ipairs(loaded[category] or {}) do
            io.stderr:write(string.format("%s:%d:%d: %s%s\n", diagnostic.filename or loaded_name,
               diagnostic.y or 1, diagnostic.x or 1, category == "warnings" and "warning: " or "", diagnostic.msg))
            if category ~= "warnings" then failed = true end
         end
      end
   end
   return failed
end
-- The shared load is checked before its diagnostics: a shadowed declaration
-- module reports its identity, not the type errors it causes.
local shared_mismatch = declaration_mismatch(declarations, false)
if shared_mismatch then
   for _, source in ipairs(sources) do
      io.stderr:write(source .. ":1:1: declaration identity mismatch: " .. shared_mismatch .. "\n")
   end
   os.exit(1)
end
if report_stock(declarations, 1) then fail_run("declarations do not type-check") end
local shared_loaded = #declarations.loaded_order

local function copy(t)
   local result = {}
   for key, value in pairs(t) do result[key] = value end
   return result
end
-- A source's environment over the shared declaration load. Teal adds the
-- source's globals, modules and loaded files to these tables and marks global
-- variables as it reads and writes them, so each source gets its own copies.
local function source_env()
   local globals = {}
   for name, variable in pairs(declarations.globals) do globals[name] = copy(variable) end
   return {
      modules = copy(declarations.modules),
      module_filenames = copy(declarations.module_filenames),
      loaded = copy(declarations.loaded),
      loaded_order = copy(declarations.loaded_order),
      globals = globals,
      defaults = declarations.defaults,
   }
end

local original_parse = tl.parse
local function ignore() end
-- The generated declarations are shared by every source, so their overload
-- sets are checked once per run.
local generated_overloads_checked = false

-- Checks one source and prints its diagnostics; returns whether it passed.
local function check_source(filename)
   local original_is, capture_error = {}, nil
   tl.parse = function(input, source_name, parse_lang)
      local ast, errors, required = original_parse(input, source_name, parse_lang)
      if source_name == filename then
         -- The compiler replaces `is Record` nodes with expanded where bodies.
         -- Keep their original operands while retaining the same type objects.
         local seen, count = {}, 0
         local function capture(node, depth)
            if seen[node] then return end
            seen[node], count = true, count + 1
            if depth > 256 or count > 2000000 then error("source syntax capture exceeds AST bounds") end
            if node.kind == "op" and node.op.op == "is" then
               local snapshot = {}; for key, value in pairs(node) do snapshot[key] = value end
               original_is[node] = snapshot
            end
            syntax_children(node, function(child) capture(child, depth + 1) end)
         end
         local ok, err = pcall(capture, ast, 0)
         if not ok then capture_error = err end
      end
      return ast, errors, required
   end
   local expression_types, callee_types, declaration_types = {}, {}, {}
   local env = source_env()
   -- The checker owns the reporter: Teal's TypeReporter numbers every loaded
   -- declaration type per file, and this pass reads only `store_type` positions.
   -- `store_type` is called by Teal directly, so frame_locals sees the storing frame.
   env.report_types = true
   env.reporter = {
      set_ref = ignore,
      add_field = ignore,
      store_result = ignore,
      get_collector = function(_, source_name)
         return {
            reserve_symbol_list_slot = ignore,
            add_to_symbol_list = ignore,
            begin_symbol_list_scope = ignore,
            rollback_symbol_list_scope = ignore,
            end_symbol_list_scope = ignore,
            store_type = function(y, x, typ)
               if source_name ~= filename or not typ then return end
               local names, values = frame_locals()
               if names[1] == "self" and names[2] == "n" and names[3] == "children" and names[4] == "t" then
                  if is_node_at(values.n, y, x) then record(expression_types, values.n, typ) end
               elseif is_node_at(values.e1, y, x) and type(values.node) == "table" and values.node.e1 == values.e1 then
                  record(callee_types, values.e1, typ)
               elseif is_node_at(values.var, y, x) then record(declaration_types, values.var, typ)
               elseif is_node_at(values.v, y, x) then record(declaration_types, values.v, typ)
               end
            end,
         }
      end,
   }
   local result, process_error = tl.process(filename, env)
   tl.parse = original_parse
   if not result then io.stderr:write(filename .. ":1:1: " .. tostring(process_error) .. "\n"); return false end
   if capture_error then fail_closed(capture_error) end
   local mismatch = declaration_mismatch(env, true)
   if mismatch then
      io.stderr:write(filename .. ":1:1: declaration identity mismatch: " .. mismatch .. "\n")
      return false
   end
   if report_stock(env, shared_loaded + 1) then return false end

   -- Opacity identity: the typedecls that Teal loaded for native_values. Teal
   -- resolves every nominal reference to one of these objects, so a same-named
   -- declaration elsewhere never confers opacity. Only Secret, Data and
   -- SecretTable are compared; the rest of the vocabulary is required so that a
   -- renamed or removed value fails closed instead of silently becoming plain.
   local markers = {}
   do
      local module = env.modules.native_values
      local fields = type(module) == "table" and module.typename == "typedecl" and type(module.def) == "table" and module.def.fields
      if type(fields) ~= "table" then fail_closed("opaque marker definition changed: native_values") end
      for _, name in ipairs({ "Secret", "MaybeSecret", "Data", "Atom", "SecretData", "SecretTable" }) do
         local decl = fields[name]
         if type(decl) ~= "table" or decl.typename ~= "typedecl" then fail_closed("opaque marker definition changed: " .. name) end
         markers[name] = decl
      end
   end

   -- Stdlib builtins are identified by declaration location, so a shadowing
   -- local or record field never inherits a builtin exemption.
   local stdlib_names, stdlib_declarations = {}, {}
   do
      local function index(name, typ, depth)
         if type(typ) ~= "table" or depth > 8 then return end
         if typ.typename == "function" and typ.f == "stdlib.d.tl" then
            local location = tostring(typ.y) .. ":" .. tostring(typ.x)
            stdlib_names[location], stdlib_declarations[location] = name, typ
         elseif typ.typename == "poly" then for _, member in ipairs(typ.types) do index(name, member, depth + 1) end
         elseif typ.typename == "generic" then index(name, typ.t, depth + 1)
         elseif typ.typename == "typedecl" then index(name, typ.def, depth + 1) end
      end
      local globals = env.globals or {}
      for _, name in ipairs({ "pcall", "xpcall", "select", "type", "assert", "ipairs", "pairs", "next", "rawget", "rawset", "setmetatable", "tostring", "tonumber", "print", "error" }) do
         local variable = globals[name]
         if not variable or not variable.t then fail_closed("compiler omitted stdlib identity: " .. name) end
         index(name, variable.t, 0)
      end
      local table_type = globals.table and globals.table.t
      table_type = table_type and (table_type.def or table_type)
      if not table_type or not table_type.fields then fail_closed("compiler omitted stdlib identity: table") end
      for _, name in ipairs({ "pack", "unpack", "insert", "remove", "concat", "sort" }) do
         index("table." .. name, table_type.fields[name], 0)
      end
      local coroutine_type = globals.coroutine and globals.coroutine.t
      coroutine_type = coroutine_type and (coroutine_type.def or coroutine_type)
      if not coroutine_type or not coroutine_type.fields then fail_closed("compiler omitted stdlib identity: coroutine") end
      for _, name in ipairs({ "create", "resume", "wrap", "yield" }) do
         index("coroutine." .. name, coroutine_type.fields[name], 0)
      end
   end
   local function stdlib_name(typ)
      if typ and typ.f == "stdlib.d.tl" then return stdlib_names[tostring(typ.y) .. ":" .. tostring(typ.x)] or "stdlib" end
   end
   -- Lua reads a metatable's `__index` table on a missed key and never writes
   -- through it (`__newindex` receives the writes), so a record table placed
   -- there is inert, but only the table itself: a table nested in the
   -- `__index` value is read back typed by the instance's field. Only a fresh
   -- table constructor or an upcast record puts it there; assigning
   -- `mt.__index` reaches a slot another alias of the same table may type
   -- (`local t = { __index = seed }; local mt: metatable<H> = t`). The Teal
   -- prelude's `metatable` record is identified by its declaration location,
   -- which its instantiations keep. The position is the `__index` key of such
   -- a record, not its field type: Teal gives that type object to a local
   -- inferred from `mt.__index`.
   local metatable_declaration
   do
      local decl = env.globals.metatable and env.globals.metatable.t
      local def = decl and decl.def
      def = def and def.typename == "generic" and def.t or def
      if not (def and def.typename == "record" and def.f == "prelude.d.tl" and def.fields and def.fields.__index) then
         fail_closed("compiler omitted stdlib identity: metatable")
      end
      metatable_declaration = def
   end
   local function is_metatable(t)
      local d = metatable_declaration
      return t.typename == "record" and t.f == d.f and t.y == d.y and t.x == d.x
   end
   -- The builtin name shared by every overload of a poly stdlib function.
   local function stdlib_member_name(typ)
      if typ.typename ~= "poly" then return stdlib_name(typ) end
      local name
      for _, member in ipairs(typ.types) do
         local member_name = stdlib_name(member.typename == "generic" and member.t or member)
         if not member_name or (name and name ~= member_name) then return nil end
         name = member_name
      end
      return name
   end
   -- Instantiation binds stdlib type variables to the argument types, so a
   -- transport like select or table.pack is compared against its declaration.
   local function stdlib_declaration(typ)
      return stdlib_declarations[tostring(typ.y) .. ":" .. tostring(typ.x)] or typ
   end

   local visits = 0
   local function visit()
      visits = visits + 1
      if visits > 2000000 then error("secret check exceeds two million AST visits") end
   end

   local function lookup(map, node)
      local typ = map[node]
      if typ == AMBIGUOUS then error(string.format("ambiguous type identity at %d:%d", node.y or 0, node.x or 0)) end
      return typ
   end

   -- Type model over Teal's raw type objects.
   local function is_marker(t, name)
      return t.typename == "nominal" and t.found == markers[name]
   end
   local function is_secret(t) return is_marker(t, "Secret") end
   local function is_data(t) return is_marker(t, "Data") end
   local function is_secret_table(t) return is_marker(t, "SecretTable") end
   -- Teal resolves a nominal's instantiation only where it compared it; types read
   -- from signatures can still carry bare type arguments. Substitute them here.
   -- Instantiations are keyed by declaration and type-argument identity, not by
   -- the nominal, and substitution keeps every subtree no binding reaches: a
   -- generic record naming itself (`next: Box<T>`, `meta: Box<string>`,
   -- `map: function<U>(Box<T>): Box<U>`) then reaches the same instantiation at
   -- every level instead of instantiating without end.
   local instantiations = {}
   local INSTANCE = {}
   local substituted_fields = { "types", "tuple", "elements", "keys", "values", "args", "rets", "typevals", "t", "def", "vtype" }
   local function substitute(t, bindings, memo, depth)
      if type(t) ~= "table" then return t end
      if memo[t] then return memo[t] end
      depth = depth + 1
      if depth > 64 then error("generic instantiation exceeds type depth") end
      if t.typename == "typevar" and bindings[t.typevar] then return bindings[t.typevar] end
      if t.typename == "generic" then
         for _, argument in ipairs(t.typeargs or {}) do
            if bindings[argument.typearg] then
               local inner = {}; for name, value in pairs(bindings) do inner[name] = value end
               inner[argument.typearg] = nil
               bindings = inner
            end
         end
      end
      -- The copy is memoised before its children, so a cycle back to `t` makes
      -- that child differ and the copy is kept.
      local out = setmetatable({}, getmetatable(t))
      memo[t] = out
      local changed = false
      local function child_of(member)
         local substituted = substitute(member, bindings, memo, depth)
         changed = changed or substituted ~= member
         return substituted
      end
      for key, value in pairs(t) do out[key] = value end
      if out.typename == "nominal" then out.resolved = nil end
      for _, field in ipairs(substituted_fields) do
         local child = t[field]
         if type(child) == "table" then
            if child.typename then out[field] = child_of(child)
            else
               local list = {}
               for index, member in ipairs(child) do list[index] = child_of(member) end
               out[field] = list
            end
         end
      end
      for _, field in ipairs({ "fields", "meta_fields" }) do
         if type(t[field]) == "table" then
            local map = {}
            for name, member in pairs(t[field]) do map[name] = child_of(member) end
            out[field] = map
         end
      end
      if changed then return out end
      memo[t] = t
      return t
   end
   local function instantiate(t)
      local decl = t.found
      local generic = decl and decl.typename == "typedecl" and decl.def
      if not generic or generic.typename ~= "generic" or #generic.typeargs ~= #t.typevals then
         error("compiler omitted generic instantiation")
      end
      local cache = instantiations[decl] or {}
      instantiations[decl] = cache
      for _, value in ipairs(t.typevals) do
         local inner = cache[value] or {}
         cache[value] = inner
         cache = inner
      end
      if cache[INSTANCE] then return cache[INSTANCE] end
      local bindings = {}
      for index, argument in ipairs(generic.typeargs) do bindings[argument.typearg] = t.typevals[index] end
      local result = substitute(generic.t, bindings, {}, 0)
      cache[INSTANCE] = result
      return result
   end
   -- `quantified`, when given, collects the type arguments of the generics
   -- passed through.
   local function resolve(t, quantified)
      for _ = 1, 64 do
         if type(t) ~= "table" then error("compiler omitted type identity") end
         local name = t.typename
         if name == "nominal" then
            if is_secret(t) or is_data(t) then return t end
            local next_type = t.resolved
            if not next_type then
               next_type = t.typevals and instantiate(t) or t.found
            end
            if not next_type or next_type == t then return t end
            t = next_type
         elseif name == "typedecl" then t = t.def
         elseif name == "generic" then
            for _, argument in ipairs(quantified and t.typeargs or {}) do quantified[argument.typearg] = true end
            t = t.t
         elseif name == "literal_table_item" then t = t.vtype
         else return t end
      end
      error("cannot resolve bounded type identity")
   end
   -- A generic record's own table (`Box` read as a value, not a `Box<T>`):
   -- one table that every instantiation reads through `__index`. Teal types it
   -- as the generic's body with its type parameters unbound, so it admits
   -- writes of any type to their fields. Returns that body when `t` resolves
   -- through a generic to a record or interface; an instance never does, its
   -- nominal resolves to an instantiated copy.
   local function record_table(t)
      local quantified = {}
      local body = resolve(t, quantified)
      if next(quantified) and (body.typename == "record" or body.typename == "interface") then return body end
      return nil
   end
   local function record_table_name(body) return tostring(body.declname or "of a generic record") end
   local function payload(t)
      local typevals = t.typevals
      if not typevals or #typevals ~= 1 then error("compiler omitted Secret payload") end
      return typevals[1]
   end

   -- Top-level opacity: a Secret value, or a union that may hold one.
   local function opaque(t)
      if t == nil then return false end
      t = resolve(t)
      if is_secret(t) then return true end
      if t.typename == "union" then
         for _, member in ipairs(t.types) do if opaque(member) then return true end end
      end
      return false
   end
   -- A function-typed record member, optionally `| nil`.
   local function callable_member(t)
      t = resolve(t)
      if t.typename == "union" then
         local only
         for _, member in ipairs(t.types) do
            local resolved = resolve(member)
            if resolved.typename ~= "nil" then
               if only then return false end
               only = resolved
            end
         end
         t = only or t
      end
      return t.typename == "function" or t.typename == "poly"
   end
   -- Opacity anywhere inside the type, including containers and signatures.
   -- `skip(member)` excludes a record member from the walk.
   local function deep(t, seen, depth, skip)
      if t == nil then return false end
      seen, depth = seen or {}, (depth or 0) + 1
      if depth > 64 then error("opaque contents exceed type depth") end
      t = resolve(t)
      if seen[t] then return false end
      seen[t] = true
      if is_secret(t) then return true end
      if is_data(t) then return false end
      local name = t.typename
      if name == "array" then return deep(t.elements, seen, depth, skip) end
      if name == "map" then return deep(t.keys, seen, depth, skip) or deep(t.values, seen, depth, skip) end
      if name == "tuple" then for _, member in ipairs(t.tuple) do if deep(member, seen, depth, skip) then return true end end end
      if name == "union" or name == "tupletable" or name == "poly" then
         for _, member in ipairs(t.types) do if deep(member, seen, depth, skip) then return true end end
      end
      if name == "record" or name == "interface" then
         for _, field in pairs(t.fields or {}) do
            if not (skip and skip(field)) and deep(field, seen, depth, skip) then return true end
         end
         if t.elements and deep(t.elements, seen, depth, skip) then return true end
      end
      if name == "function" then
         return deep(t.args, seen, depth, skip) or deep(t.rets, seen, depth, skip)
      end
      return false
   end
   -- A handle reaching a sink carries no secret value, and its method contracts
   -- reattach wherever it is used at a declared type.
   local function function_member(member) return callable_member(member) end
   -- Assigned once the generated declarations are identified.
   local generated_handle_member
   -- An authored type variable may be instantiated with a Secret
   -- (`Box<MaybeSecret<number>>` is legal), so a value of type-variable type,
   -- or a union holding one, is tested, compared, operated on and cast only
   -- under the proofs a Secret needs.
   local function variable_typed(t)
      t = resolve(t)
      if t.typename == "typevar" or t.typename == "typearg" then return true end
      if t.typename ~= "union" then return false end
      for _, member in ipairs(t.types) do
         if variable_typed(member) then return true end
      end
      return false
   end
   -- Teal 0.24.8 gives a generic referenced by name fresh type variables
   -- (`U@12`, fresh_typeargs): a callee's own, which Teal binds at the call
   -- from the arguments the checker compares there. Any other is a type
   -- parameter of the code around the call, which the call does not bind.
   local function instance_variable(name)
      return name:find("@", 1, true) ~= nil
   end
   -- A type parameter of the authored code being checked: not quantified by a
   -- generic inside the type (`function<A>(A): A`), nor a call's instance.
   local function parameter_variable(t, bound)
      local name = t.typename == "typevar" and t.typevar or t.typename == "typearg" and t.typearg
      return name and not (bound and bound[name]) and not instance_variable(name) or false
   end
   -- Such a parameter anywhere inside the type, traversed as deep does. It may
   -- be instantiated with a Secret, so a value holding one reaches `any`, Data
   -- and plain positions only as a Secret would.
   -- A record table's parameters are free: the table reaches typed positions
   -- only as a value holding them. `inert` binds them where the table is held
   -- inert (see the record table rule in leak): nothing writes its fields
   -- there, so they hold nothing. Nested type declarations are not values; a
   -- table read through one is checked where it is read.
   local QUANTIFIED = {}
   local function holds_variable(t, values, inert, bound, seen, depth)
      if t == nil then return false end
      bound, seen, depth = bound or { [QUANTIFIED] = "" }, seen or {}, (depth or 0) + 1
      if depth > 64 then error("type variable search exceeds type depth") end
      local quantified = {}
      t = resolve(t, quantified)
      if not inert and (t.typename == "record" or t.typename == "interface") then quantified = {} end
      if next(quantified) then
         local inner, names = {}, {}
         for name in pairs(bound) do inner[name] = true end
         for name in pairs(quantified) do inner[name] = true end
         for name in pairs(inner) do if type(name) == "string" then names[#names + 1] = name end end
         table.sort(names)
         inner[QUANTIFIED] = table.concat(names, ",")
         bound = inner
      end
      local marks = seen[t] or {}
      seen[t] = marks
      if marks[bound[QUANTIFIED]] then return false end
      marks[bound[QUANTIFIED]] = true
      local name = t.typename
      if name == "typevar" or name == "typearg" then return parameter_variable(t, bound) end
      if is_secret(t) or is_data(t) then return false end
      local function any_of(list)
         for _, member in ipairs(list or {}) do
            if holds_variable(member, values, inert, bound, seen, depth) then return true end
         end
         return false
      end
      if name == "array" then return holds_variable(t.elements, values, inert, bound, seen, depth) end
      if name == "map" then return holds_variable(t.keys, values, inert, bound, seen, depth) or holds_variable(t.values, values, inert, bound, seen, depth) end
      if name == "tuple" then return any_of(t.tuple) end
      if name == "union" or name == "tupletable" or name == "poly" then return any_of(t.types) end
      if name == "record" or name == "interface" then
         for _, field in pairs(t.fields or {}) do
            if field.typename ~= "typedecl" and not (values and callable_member(field))
               and holds_variable(field, values, inert, bound, seen, depth) then return true end
         end
         return t.elements ~= nil and holds_variable(t.elements, values, inert, bound, seen, depth)
      end
      if name == "function" then
         return holds_variable(t.args, values, inert, bound, seen, depth) or holds_variable(t.rets, values, inert, bound, seen, depth)
      end
      return false
   end

   local function coarse(t)
      if is_secret(t) then return "secret" end
      if is_data(t) then return "data" end
      local name = t.typename
      if name == "integer" or name == "number" then return "number" end
      if name == "enum" or name == "string" then return "string" end
      if name == "array" or name == "map" or name == "tupletable" or name == "record" or name == "interface" or name == "emptytable" then
         return t.is_userdata and "userdata" or "table"
      end
      if name == "poly" then return "function" end
      return name
   end

   -- Structural equality for `as` targets and overload parameters; nominal types
   -- compare by declaration. `renamed` maps the type arguments of enclosing
   -- generic functions on the left to those on the right, so generic functions
   -- compare up to the names of their type arguments. Other type variables are
   -- equal only to themselves.
   local same_type
   local function bind_typeargs(left, right, renamed, depth)
      if #left.typeargs ~= #right.typeargs then return nil end
      local inner = {}
      for name, other in pairs(renamed or {}) do inner[name] = other end
      for index, argument in ipairs(left.typeargs) do
         local other = right.typeargs[index]
         if (argument.constraint == nil) ~= (other.constraint == nil)
            or (argument.constraint and not same_type(argument.constraint, other.constraint, depth, renamed)) then return nil end
         inner[argument.typearg] = other.typearg
      end
      return inner
   end
   local function same_tuple(left, right, depth, renamed)
      if #left.tuple ~= #right.tuple or not left.is_va ~= not right.is_va then return false end
      for index = 1, #left.tuple do
         if not same_type(left.tuple[index], right.tuple[index], depth, renamed) then return false end
      end
      return true
   end
   same_type = function(left, right, depth, renamed)
      depth = (depth or 0) + 1
      if depth > 64 then error("type equality exceeds type depth") end
      if left == right then return true end
      if left.typename == "generic" and right.typename == "generic" then
         local inner = bind_typeargs(left, right, renamed, depth)
         return inner ~= nil and same_type(left.t, right.t, depth, inner)
      end
      if left.typename == "nominal" and right.typename == "nominal" then
         if is_secret(left) ~= is_secret(right) or is_data(left) ~= is_data(right) then return false end
         if left.found and left.found == right.found then
            local lv, rv = left.typevals or {}, right.typevals or {}
            if #lv ~= #rv then return false end
            for index = 1, #lv do if not same_type(lv[index], rv[index], depth, renamed) then return false end end
            return true
         end
      end
      if is_secret(left) or is_secret(right) or is_data(left) or is_data(right) then
         if left.typename ~= "nominal" or right.typename ~= "nominal" then
            return (is_secret(left) or is_data(left)) == (is_secret(right) or is_data(right)) and resolve(left) == resolve(right)
         end
         return false
      end
      local a, b = resolve(left), resolve(right)
      if a == b then return true end
      if a.typename ~= b.typename then return false end
      local name = a.typename
      if name == "number" or name == "integer" or name == "string" or name == "boolean" or name == "nil" or name == "any" then return true end
      if name == "typevar" then return renamed ~= nil and renamed[a.typevar] == b.typevar end
      if name == "typearg" then return renamed ~= nil and renamed[a.typearg] == b.typearg end
      -- Teal types a declared method's receiver as `self`: the called object.
      if name == "self" then return true end
      if name == "array" then return same_type(a.elements, b.elements, depth, renamed) end
      if name == "map" then return same_type(a.keys, b.keys, depth, renamed) and same_type(a.values, b.values, depth, renamed) end
      if name == "union" then
         if #a.types ~= #b.types then return false end
         for _, member in ipairs(a.types) do
            local found = false
            for _, other in ipairs(b.types) do if same_type(member, other, depth, renamed) then found = true; break end end
            if not found then return false end
         end
         return true
      end
      if name == "function" then
         return a.min_arity == b.min_arity and same_tuple(a.args, b.args, depth, renamed) and same_tuple(a.rets, b.rets, depth, renamed)
      end
      return false
   end

   -- Parameter positions that hold one type in every overload of a set. An
   -- argument there suits every overload of the set or none, so it never
   -- decides which overload a call selects. A receiver is the usual case.
   local function uniform_positions(overloads)
      local uniform = {}
      local first = resolve(overloads[1])
      if first.typename ~= "function" then return uniform end
      for index, parameter in ipairs(first.args.tuple) do
         uniform[index] = true
         for other_index = 2, #overloads do
            local other = resolve(overloads[other_index])
            local member = other.typename == "function" and other.args.tuple[index]
            if not member or not same_type(parameter, member) then uniform[index] = false; break end
         end
      end
      return uniform
   end
   -- Whether two overloads accept the same argument lists, so Teal's
   -- first-match overload choice picks them for the same calls. A position
   -- uniform within both sets is not compared: an argument there suits all of
   -- either side's overloads or none of them.
   local function same_parameters(left, right, uniform_left, uniform_right)
      local renamed
      if left.typename == "generic" and right.typename == "generic" then
         renamed = bind_typeargs(left, right, nil, 0)
         if not renamed then return false end
      end
      local a, b = resolve(left), resolve(right)
      if a.typename ~= "function" or b.typename ~= "function" then return false end
      if #a.args.tuple ~= #b.args.tuple or not a.args.is_va ~= not b.args.is_va or a.min_arity ~= b.min_arity then return false end
      for index = 1, #a.args.tuple do
         if not (uniform_left[index] and uniform_right[index]) and not same_type(a.args.tuple[index], b.args.tuple[index], 0, renamed) then return false end
      end
      return true
   end

   -- Returns a reason when opacity in `actual` can reach a non-opaque position of
   -- `expected`. `stdlib` marks a trusted stdlib signature, whose type variables
   -- transport values; authored type variables, `any` and Data are sinks.
   local leak
   local function sink(t, ctx)
      if t.typename == "any" or is_data(t) then return true end
      if t.typename == "typevar" or t.typename == "typearg" then return not ctx.trusted end
      return false
   end
   -- `ctx.transport` names the side ("actual" or "expected") holding a trusted
   -- callee's own signature; its type variables are transports whose bindings
   -- check_bindings owns. Reversing a comparison reverses that side.
   -- `ctx.inert` marks a position that only reads a record table (an argument
   -- of `print` or `tostring`, a metatable's `__index`), and `ctx.written` an
   -- argument the callee writes into (Mixin's target); both describe the
   -- forward position, so a reversal drops them.
   local function reversed(ctx)
      local transport = ctx.transport and (ctx.transport == "actual" and "expected" or "actual")
      return { trusted = ctx.trusted, fresh = ctx.fresh, cast = ctx.cast, exempt = ctx.exempt, transport = transport }
   end
   local function invariant(actual, expected, ctx, depth, what)
      local reason = leak(actual, expected, ctx, depth) or (not ctx.fresh and leak(expected, actual, reversed(ctx), depth))
      return reason and (what .. ": " .. reason) or nil
   end
   -- An alias such as `MaybeSecret<T> | nil` nests a union inside a union; match
   -- against the flattened, resolved members.
   local function union_members(t, out, depth)
      out, depth = out or {}, (depth or 0) + 1
      if depth > 64 then error("union nesting exceeds type depth") end
      for _, member in ipairs(t.types) do
         local resolved = resolve(member)
         if resolved.typename == "union" then union_members(resolved, out, depth)
         else out[#out + 1] = resolved end
      end
      return out
   end
   -- The function types a callback is compared against: the expected function,
   -- or the function members of an expected union.
   local function callback_targets(expected)
      if expected.typename == "function" then return { expected } end
      local out = {}
      if expected.typename == "union" then
         for _, member in ipairs(union_members(expected)) do
            if member.typename == "function" then out[#out + 1] = member end
         end
      end
      return out
   end
   local function member_of(tuple, index)
      local member = tuple.tuple[index]
      if member == nil and tuple.is_va then member = tuple.tuple[#tuple.tuple] end
      return member
   end
   -- A function returns every result it has, whatever its type declares, and a
   -- call in last position passes them all on, so a result position the
   -- receiving type lacks reaches an unknown position: it is judged as `any`.
   local ANY = { typename = "any" }
   local function dropped_leak(actual, ctx, depth)
      return leak(actual, ANY, ctx, depth) and "Secret result dropped by a narrower function type" or nil
   end
   -- Field names in declaration order, so the reported field does not depend on
   -- table iteration order.
   local function field_names(t)
      if t.field_order then return t.field_order end
      local names = {}
      for name in pairs(t.fields or {}) do names[#names + 1] = name end
      table.sort(names)
      return names
   end
   local compare
   -- Recursive types make a comparison recur inside itself (`interface Unit`
   -- whose method takes `other: Unit`). A comparison already in progress is
   -- assumed to hold: a leak has a shortest witness, which never repeats one.
   local comparing = {}
   local function inert_ctx(ctx)
      return { trusted = ctx.trusted, fresh = ctx.fresh, cast = ctx.cast, exempt = ctx.exempt, transport = ctx.transport, inert = true }
   end
   local function record_table_reason(body)
      return "record table " .. record_table_name(body) .. " reaches a position that can write it"
   end
   leak = function(actual, expected, ctx, depth)
      visit()
      depth = (depth or 0) + 1
      if depth > 64 then error("flow comparison exceeds type depth") end
      if actual == nil or expected == nil then return nil end
      local a, e = resolve(actual), resolve(expected)
      -- Every ctx field that changes the verdict is in the key: an inert
      -- comparison in progress must not assume a non-inert one on the same pair.
      local key = string.format("%s|%s|%s|%s|%s|%s|%s", tostring(ctx.trusted), tostring(ctx.fresh), tostring(ctx.cast), tostring(ctx.transport), tostring(ctx.exempt), tostring(ctx.inert), tostring(ctx.written))
      local by_actual = comparing[a] or {}
      comparing[a] = by_actual
      local keys = by_actual[e] or {}
      by_actual[e] = keys
      if keys[key] then return nil end
      keys[key] = true
      local reason
      -- Any position but its own type can write a record table's parameter
      -- fields under an instantiation the other instances do not share: an
      -- instance type directly, `any` and Data through rawset or a downcast.
      -- It may reach only its own type, an inert position, and a trusted type
      -- variable check_bindings follows to its destinations, unless the callee
      -- writes into that argument.
      local table_body = record_table(actual)
      if table_body and not ctx.inert
         and not ((e.typename == "typevar" or e.typename == "typearg") and ctx.trusted and ctx.transport == "expected" and not ctx.written)
         and record_table(expected) ~= table_body then
         reason = record_table_reason(table_body)
      else
         reason = compare(a, e, ctx, depth, actual)
      end
      keys[key] = nil
      if next(keys) == nil then by_actual[e] = nil end
      if next(by_actual) == nil then comparing[a] = nil end
      return reason
   end
   compare = function(a, e, ctx, depth, actual)
      if e.typename == "boolean_context" then return nil end
      -- A single position receives only a tuple's first value.
      if a.typename == "tuple" and e.typename ~= "tuple" then return leak(member_of(a, 1), e, ctx, depth) end
      if e.typename == "tuple" or a.typename == "tuple" then
         local actual_tuple = a.typename == "tuple" and a or { tuple = { a } }
         local expected_tuple = e.typename == "tuple" and e or { tuple = { e } }
         local count = math.max(#actual_tuple.tuple, #expected_tuple.tuple)
         for index = 1, count do
            local actual_member, expected_member = member_of(actual_tuple, index), member_of(expected_tuple, index)
            local reason = actual_member and expected_member and leak(actual_member, expected_member, ctx, depth)
            if reason then return "value " .. index .. ": " .. reason end
         end
         return nil
      end
      -- A type variable may be instantiated with a Secret without a checked call
      -- (`Box<MaybeSecret<T>>`), so it cannot reach a sink, trusted or not.
      -- Nor can it reach a plain position, where a later read types it plain
      -- (`xs as {number}`, a constrained `T` returned as its constraint).
      local variable_actual = (a.typename == "typevar" or a.typename == "typearg") and ctx.transport ~= "actual"
      if variable_actual and (e.typename == "any" or is_data(e)) then
         return "type variable reaches " .. (is_data(e) and "Data" or "any")
      end
      if variable_actual and parameter_variable(a) and not variable_typed(e) and not opaque(e) then
         return "type variable reaches a plain position"
      end
      -- A trusted (stdlib or types/wow) sink checks value positions only: it cannot
      -- recast a handle to a lookalike record with plain methods. Authored sinks and
      -- casts keep the full rule, since `x as L` from `any` is unchecked.
      if sink(e, ctx) then
         local values = ctx.trusted and not ctx.cast
         if deep(a, nil, nil, values and function_member) then return "Secret reaches " .. (is_data(e) and "Data" or e.typename == "any" and "any" or "type variable") end
         if (e.typename == "any" or is_data(e)) and ctx.transport ~= "actual" and holds_variable(actual, values, ctx.inert) then
            return "type variable reaches " .. (is_data(e) and "Data" or "any")
         end
         return nil
      end
      if a.typename == "any" or a.typename == "nil" or a.typename == "typevar" or a.typename == "typearg" then return nil end
      -- A trusted stdlib type variable transports the value unchanged.
      if e.typename == "typevar" or e.typename == "typearg" then return nil end
      -- Callbacks are compared before the opacity shortcut: a callee may hand a
      -- callback a type variable, which is not opaque but may not reach a sink.
      if a.typename == "function" and e.typename == "function" then
         for index = 1, math.max(#e.args.tuple, #a.args.tuple) do
            local expected_parameter, actual_parameter = member_of(e.args, index), member_of(a.args, index)
            local reason = expected_parameter and actual_parameter and leak(expected_parameter, actual_parameter, reversed({ trusted = false, fresh = false, cast = ctx.cast, transport = ctx.transport }), depth)
            if reason then return "callback parameter " .. index .. ": " .. reason end
         end
         -- A variadic actual may return one more result than a fixed expected
         -- type declares; member_of yields the vararg member for that position.
         local spare = a.rets.is_va and not e.rets.is_va and 1 or 0
         for index = 1, math.max(#a.rets.tuple, #e.rets.tuple + spare) do
            local actual_result, expected_result = member_of(a.rets, index), member_of(e.rets, index)
            local result_ctx = { trusted = ctx.trusted, fresh = false, cast = ctx.cast, transport = ctx.transport }
            local reason
            if expected_result then
               reason = actual_result and leak(actual_result, expected_result, result_ctx, depth)
            elseif actual_result then
               reason = dropped_leak(actual_result, result_ctx, depth)
            end
            if reason then return "callback result " .. index .. ": " .. reason end
         end
         return nil
      end
      -- A caller may call through any overload, so the function must fit each.
      if a.typename == "function" and e.typename == "poly" then
         for _, member in ipairs(e.types) do
            local reason = leak(a, member, ctx, depth)
            if reason then return reason end
         end
         return nil
      end
      if a.typename == "function" and e.typename == "union" then
         local first
         for _, target in ipairs(callback_targets(e)) do
            local reason = leak(a, target, ctx, depth)
            if not reason then return nil end
            first = first or reason
         end
         if first then return first end
      end
      -- Teal admits nil in every parameter and an explicit trailing nil where a
      -- call through the target omits it, so argument types and arity never
      -- prove which source overload a call runs: every source overload must
      -- fit every target overload. Overloads with the same parameter lists in
      -- the same order are chosen for the same calls, so they pair by position.
      if a.typename == "poly" and (e.typename == "function" or e.typename == "poly") then
         local targets = e.typename == "poly" and e.types or { e }
         local paired = #a.types == #targets
         local uniform_actual, uniform_expected = paired and uniform_positions(a.types), paired and uniform_positions(targets)
         for index = 1, paired and #targets or 0 do
            if not same_parameters(a.types[index], targets[index], uniform_actual, uniform_expected) then paired = false; break end
         end
         for index, target in ipairs(targets) do
            for _, member in ipairs(paired and { a.types[index] } or a.types) do
               local reason = leak(member, target, ctx, depth)
               if reason then return reason end
            end
         end
         return nil
      end
      if not deep(a) and not deep(e) and (ctx.transport == "actual" or not holds_variable(a)) then return nil end
      if opaque(a) and not opaque(e) then return "Secret reaches a plain position" end
      if a.typename == "union" then
         for _, member in ipairs(a.types) do
            local reason = leak(member, e, ctx, depth)
            if reason then return reason end
         end
         return nil
      end
      if e.typename == "union" then
         local kind, first = coarse(a), nil
         for _, resolved in ipairs(union_members(e)) do
            if coarse(resolved) == kind then
               local reason = leak(a, resolved, ctx, depth)
               if not reason then return nil end
               first = first or reason
            end
         end
         return first or (deep(a) and "Secret reaches a plain position" or nil)
      end
      if is_secret(a) and is_secret(e) then return nil end
      local an, en = a.typename, e.typename
      local element_ctx = { trusted = ctx.trusted, fresh = ctx.fresh, cast = ctx.cast, transport = ctx.transport }
      local expected_values = (en == "array" and e.elements) or (en == "map" and e.values) or ((en == "record" or en == "interface") and e.elements)
      if an == "array" or an == "tupletable" or ((an == "record" or an == "interface") and a.elements) then
         local actual_elements = an == "tupletable" and a.types or { a.elements }
         if en == "tupletable" then
            for index, member in ipairs(e.types) do
               local actual_member = actual_elements[index] or actual_elements[#actual_elements]
               local reason = invariant(actual_member, member, element_ctx, depth, "container element")
               if reason then return reason end
            end
            return nil
         end
         if expected_values then
            for _, member in ipairs(actual_elements) do
               local reason = invariant(member, expected_values, element_ctx, depth, "container element")
               if reason then return reason end
            end
            if en == "map" then
               local reason = invariant({ typename = "integer" }, e.keys, element_ctx, depth, "container key")
               if reason then return reason end
            end
         end
      end
      if an == "map" and en == "map" then
         return invariant(a.keys, e.keys, element_ctx, depth, "container key")
            or invariant(a.values, e.values, element_ctx, depth, "container value")
      end
      if (an == "record" or an == "interface" or an == "emptytable") and (en == "record" or en == "interface") then
         if a == e then return nil end
         for _, name in ipairs(field_names(e)) do
            local expected_field = e.fields[name]
            local actual_field = not (ctx.exempt and ctx.exempt[name]) and a.fields and a.fields[name]
            local field_ctx = name == "__index" and is_metatable(e) and actual_field and record_table(actual_field) and inert_ctx(element_ctx) or element_ctx
            local reason = actual_field and invariant(actual_field, expected_field, field_ctx, depth, "field " .. name)
            if reason then return reason end
         end
         -- An upcast that hides an opaque field lets a later downcast to a
         -- lookalike re-type it as plain. Hidden methods stay allowed (handle
         -- hierarchies), and so do the methods of a generated handle inside a
         -- hidden field; authored function members inside one still count. See
         -- the method-hiding note in teal-catalogue.md.
         for _, name in ipairs(field_names(a)) do
            local actual_field = a.fields[name]
            if actual_field and not (e.fields and e.fields[name]) and not callable_member(actual_field) then
               if deep(actual_field, nil, nil, generated_handle_member) then return "field " .. name .. ": Secret hidden by upcast" end
               if ctx.transport ~= "actual" and holds_variable(actual_field) then return "field " .. name .. ": type variable hidden by upcast" end
            end
         end
         return nil
      end
      if an == "map" and (en == "record" or en == "interface") then
         for _, name in ipairs(field_names(e)) do
            local reason = invariant(a.values, e.fields[name], element_ctx, depth, "field " .. name)
            if reason then return reason end
         end
         return nil
      end
      if (an == "record" or an == "interface") and en == "map" then
         for _, name in ipairs(field_names(a)) do
            local reason = invariant(a.fields[name], e.values, element_ctx, depth, "field " .. name)
            if reason then return reason end
         end
         return nil
      end
      return nil
   end

   -- The arms of a top-level type: the payload of each Secret member and each
   -- plain member, flattened through aliases such as SecretData.
   local function arms(t)
      local resolved = resolve(t)
      local members = resolved.typename == "union" and union_members(resolved) or { resolved }
      local payloads, plain = {}, {}
      for _, member in ipairs(members) do
         if is_secret(member) then payloads[#payloads + 1] = payload(member)
         else plain[#plain + 1] = member end
      end
      return payloads, plain
   end
   -- Whether a value of type t can be a Lua table at runtime. `loose` counts
   -- any and Data, whose shape is unknown.
   local function table_capable(t, loose, depth)
      depth = (depth or 0) + 1
      if depth > 64 then error("table shape exceeds type depth") end
      local resolved = resolve(t)
      if is_secret(resolved) then return table_capable(payload(resolved), loose, depth) end
      if is_data(resolved) or resolved.typename == "any" or resolved.typename == "typevar" or resolved.typename == "typearg" then return loose end
      if resolved.typename == "union" then
         for _, member in ipairs(resolved.types) do if table_capable(member, loose, depth) then return true end end
         return false
      end
      return coarse(resolved) == "table"
   end
   local function type_text(t)
      local ok, text = pcall(tostring, t)
      return ok and text or tostring(t.typename)
   end
   local function scalar_target(t, seen, depth)
      if t == nil then error("compiler omitted conversion type identity") end
      seen, depth = seen or {}, (depth or 0) + 1
      if depth > 64 or seen[t] then error("cannot resolve bounded scalar type identity") end
      seen[t] = true
      local resolved = resolve(t)
      local scalar, marker = false, false
      if is_secret(resolved) then marker = true
      elseif resolved.typename == "union" then
         scalar = true
         for _, member in ipairs(resolved.types) do
            local member_scalar, member_marker = scalar_target(member, seen, depth)
            if not member_scalar and not member_marker then scalar, marker = false, false; break end
            marker = marker or member_marker
         end
      else
         local name = resolved.typename
         scalar = name == "number" or name == "integer" or name == "string" or name == "boolean" or name == "nil" or name == "enum"
      end
      seen[t] = nil
      return scalar, marker
   end

   local transport
   local function node_type(node)
      if transport and node.kind == "op" and node.op.op == "@funcall" then
         local transported = transport(node)
         if transported then return transported end
      end
      local typ = lookup(expression_types, node)
      if typ == nil and node.kind == "argument" then typ = node.argtype end
      return typ
   end
   local function value_type(node)
      local typ = node_type(node)
      if typ and typ.typename == "tuple" then typ = typ.tuple[1] end
      if typ == nil and (node.kind == "variable" or node.kind == "argument" or node.kind == "..." or (node.kind == "op" and node.op.op == "@funcall")) then
         error("compiler omitted expression identity")
      end
      return typ
   end
   local function variable_type(node)
      local typ = lookup(declaration_types, node)
      if typ == nil then typ = node_type(node) end
      if typ == nil then error("compiler omitted declaration identity") end
      return typ
   end

   -- A cast generates no code, so it passes on every result of its operand
   -- whatever its target declares; parentheses truncate them to one.
   local function runtime_results(node)
      if node.kind == "op" and node.op.op == "as" then return runtime_results(node.e1) end
      if node.kind == "paren" then return nil end
      local typ = node_type(node)
      return typ and typ.typename == "tuple" and typ or nil
   end
   local function cast_targets(node)
      local target = node.e2.casttype
      return target.typename == "tuple" and target or { tuple = { target } }
   end
   -- The values a list passes at runtime. A cast in last position expands too,
   -- since it does not stop its operand's results: those past its targets reach
   -- positions its type never declared and are marked `past_cast`.
   local function argument_values(arguments)
      local values = {}
      for index, node in ipairs(arguments or {}) do
         local typ = node_type(node)
         if typ == nil then error("compiler omitted argument identity") end
         if index == #arguments and typ.typename == "tuple" then
            for _, member in ipairs(typ.tuple) do values[#values + 1] = { node = node, type = member, expanded = true, variadic = typ.is_va } end
         else
            if typ.typename == "tuple" then typ = typ.tuple[1] end
            values[#values + 1] = { node = node, type = typ }
         end
      end
      local last = arguments and arguments[#arguments]
      local results = last and last.kind == "op" and last.op.op == "as" and runtime_results(last)
      if results then
         local width = #cast_targets(last).tuple
         local count = results.is_va and math.max(#results.tuple, width + 1) or #results.tuple
         for position = width + 1, count do
            values[#values + 1] = { node = last, type = member_of(results, position), expanded = true, variadic = results.is_va, past_cast = true }
         end
      end
      return values
   end
   -- The value at `index`: a last value with variadic results fills every
   -- later index too.
   local function value_at(values, index)
      local last = values[#values]
      if index > #values and last and last.variadic then return last end
      return values[index]
   end

   -- Teal types a table constructor as the expected type it was checked against,
   -- so constructors are compared item by item against the expected positions.
   local flow_leak
   local function literal_leak(node, expected, ctx, depth)
      depth = (depth or 0) + 1
      if depth > 64 then error("table constructor exceeds nesting depth") end
      local e = resolve(expected)
      if e.typename == "union" then
         local first
         for _, member in ipairs(e.types) do
            local kind = coarse(resolve(member))
            if kind == "table" or kind == "data" or resolve(member).typename == "any" then
               local reason = literal_leak(node, member, ctx, depth)
               if not reason then return nil end
               first = first or reason
            end
         end
         return first
      end
      local record = e.typename == "record" or e.typename == "interface"
      for index, item in ipairs(node) do
         local key = item.key
         local name = key and key.conststr
         local position = key and key.constnum
         if e.typename == "map" and not sink(e, ctx) then
            local key_type = key and node_type(key)
            local reason = key_type and leak(key_type, e.keys, ctx, depth)
            if reason then return "container key: " .. reason end
         end
         local actual = item.value and node_type(item.value)
         -- A call or `...` as the last positional item fills every later index.
         local values = { { node = item.value, type = actual } }
         if actual and index == #node and item.key_parsed == "implicit" then
            values = argument_values({ item.value })
         end
         local last, count = values[#values], #values
         if last and last.variadic and e.typename == "tupletable" and position then
            count = math.max(count, #e.types - position + 1)
         end
         local index_item = name == "__index" and is_metatable(e)
         for offset = 1, count do
            local value = value_at(values, offset)
            local item_ctx = index_item and value.type and record_table(value.type) and inert_ctx(ctx) or ctx
            local at = position and position + offset - 1
            local target
            if sink(e, ctx) then target = e
            elseif record and name then target = e.fields and e.fields[name]
            elseif e.typename == "tupletable" and at then target = e.types[at] or e.types[#e.types]
            elseif e.typename == "array" or (record and at) then target = e.elements
            elseif e.typename == "map" then target = e.values
            elseif value.past_cast then target = ANY end
            local reason = target and value.node and value.type and
               (value.expanded and leak(value.type, target, item_ctx, depth) or flow_leak(value.node, value.type, target, item_ctx, depth))
            if reason then return reason end
         end
      end
      return nil
   end
   flow_leak = function(node, actual, expected, ctx, depth)
      if node.kind == "literal_table" then return literal_leak(node, expected, ctx, depth) end
      if node.kind == "paren" then return flow_leak(node.e1, actual, expected, ctx, depth) end
      return leak(actual, expected, ctx, depth)
   end

   local diagnostics, reported = {}, {}
   -- One diagnostic per position, operation and missing guard: a missing
   -- publicness proof must not hide a missing `type(x)` proof on the same `as`.
   local function report(node, operation, detail, guard)
      local key = tostring(node.y) .. ":" .. tostring(node.x) .. ":" .. operation .. ":" .. (guard or "")
      if not reported[key] then
         reported[key] = true
         diagnostics[#diagnostics + 1] = { y = node.y or 1, x = node.x or 1, message = "secret-value " .. operation .. detail }
      end
   end
   local function reject(node, operation)
      report(node, operation, " requires same-value native accessibility proof")
   end
   local function reject_flow(node, reason)
      report(node, "flow", ": " .. reason .. "; unwrap with a guarded `as` or declare the position MaybeSecret")
   end
   local function copy(map)
      local out = {}; for key, value in pairs(map) do out[key] = value end; return out
   end
   -- Same-value proofs per binding, each holding the epoch it was proved in:
   -- `value` access (canaccessvalue), `indexing` access (canaccesstable),
   -- `public` (issecretvalue false), and the runtime kind from `type(x)`
   -- compared with a literal: `table`, or `nontable` for any other kind.
   -- Access depends on the calling context; publicness and kind belong to the
   -- value, so they survive into nested functions unless the binding is
   -- written there.
   local FACTS = { "value", "indexing", "public", "table", "nontable" }
   local VALUE_FACTS = { public = true, table = true, nontable = true }
   local function clear_facts(state, binding)
      for _, fact in ipairs(FACTS) do state[fact][binding] = nil end
   end
   local function state_copy(state)
      local out = { bindings = copy(state.bindings), owner = state.owner, context = state.context }
      for _, fact in ipairs(FACTS) do out[fact] = copy(state[fact]) end
      return out
   end
   local function join(states, fallback)
      if #states == 0 then return nil end
      local out = state_copy(fallback or states[1])
      for _, facts in ipairs(FACTS) do
         out[facts] = copy(states[1][facts])
         for binding in pairs(out[facts]) do
            for index = 2, #states do if not states[index][facts][binding] then out[facts][binding] = nil; break end end
         end
      end
      return out
   end
   local mutated_names, implemented_functions, call_arguments = {}, {}, {}
   -- Calls made as statements, whose results are discarded.
   local statement_calls = {}
   -- Parameters of literal callbacks that check_arguments compared against the
   -- callee's declared callback parameter; that comparison owns their flow.
   local call_checked_parameters = {}
   -- Record tables at the `__index` key of a table constructor typed as a
   -- `metatable`. flows compares them as inert positions.
   local inert_positions = {}
   local function mark_inert(value)
      while value do
         inert_positions[value] = true
         if value.kind ~= "paren" then return end
         value = value.e1
      end
   end
   -- Follow grammar children only: the checker attaches resolved types and native
   -- macro definitions to ordinary variable nodes during type checking.
   local function check_record_refinement(typ, owner, origin)
      local macro = typ.macroexp
      local argument = macro.args and macro.args[1]
      local formal = typ.args and typ.args.tuple[1]
      if not (typ.is_method and macro.is_method and #macro.args == 1 and #typ.args.tuple == 1
         and argument.tk == "self" and argument.argtype.typename == "self" and argument.argtype.display_type == owner
         and formal.typename == "self" and formal.display_type == owner) then
         error("unsupported secret-check macro expansion")
      end
      -- `where` predicates are macro definitions, so Teal does not report types
      -- for their expression nodes. Resolve only direct discriminator reads from
      -- the owning record's checked field declarations; never guess by spelling.
      local function predicate(node, depth)
         visit()
         if depth > 64 then error("secret check exceeds 64 refinement nesting levels") end
         if node.kind == "paren" then return predicate(node.e1, depth + 1) end
         if node.kind == "string" or node.kind == "integer" or node.kind == "number" or node.kind == "boolean" or node.kind == "nil" then return end
         if node.kind ~= "op" then error("unsupported record refinement expression: " .. tostring(node.kind)) end
         local op = node.op.op
         if op == "." and node.e1.kind == "variable" and node.e1.tk == "self" and node.e2.kind == "identifier" then
            local field = assert(owner.fields[node.e2.tk], "compiler omitted refinement discriminator field")
            if opaque(field) then reject(origin or node, "record refinement field inspection")
            else
               local scalar, marker = scalar_target(field)
               if not scalar or marker then error("record refinement requires a public scalar discriminator") end
            end
         elseif op == "not" then predicate(node.e1, depth + 1)
         elseif op == "and" or op == "or" or op == "==" or op == "~=" or op == "<" or op == ">" or op == "<=" or op == ">=" then
            predicate(node.e1, depth + 1); predicate(node.e2, depth + 1)
         elseif op == "is" then
            predicate(node.e1, depth + 1)
            local scalar, marker = scalar_target(node.e2.casttype)
            if not scalar or marker then reject(origin or node.e2, "record refinement runtime type test") end
         else error("unsupported record refinement operator: " .. tostring(op)) end
      end
      predicate(macro.exp, 0)
   end
   local function authored_type_macros(typ, seen, depth, owner, meta_name)
      if type(typ) ~= "table" or seen[typ] then return end
      visit()
      depth = (depth or 0) + 1
      if depth > 64 then error("secret check exceeds 64 authored type nesting levels") end
      seen[typ] = true
      if typ.macroexp then
         if owner and meta_name == "__is" then check_record_refinement(typ, owner)
         else error("unsupported secret-check macro expansion") end
      end
      for _, field in ipairs({ "def", "types", "tuple", "elements", "keys", "values", "args", "rets", "typeargs", "constraint" }) do
         local child = typ[field]
         if type(child) == "table" then
            authored_type_macros(child, seen, depth)
            for _, member in ipairs(child) do authored_type_macros(member, seen, depth) end
         end
      end
      for _, field in ipairs({ "fields", "meta_fields" }) do
         for name, member in pairs(typ[field] or {}) do
            authored_type_macros(member, seen, depth, field == "meta_fields" and typ or nil, name)
         end
      end
   end
   -- Teal 0.24.8 does not instantiate a generic record's type parameters inside
   -- its nested type declarations: in `Box<Secret>`, a nested
   -- `type Getter = function(): T` keeps a bare `T` that unifies with any type,
   -- so a Secret read through it types as plain. Returns the first watched type
   -- parameter that `t` mentions; a generic inside it shadows its own names.
   local function watched_typevar(t, names, seen, depth)
      if type(t) ~= "table" or seen[t] then return nil end
      depth = depth + 1
      if depth > 64 then error("type parameter search exceeds type depth") end
      seen[t] = true
      if t.typename == "typevar" and names[t.typevar] then return t.typevar end
      if t.typename == "generic" then
         local inner = {}
         for name in pairs(names) do inner[name] = true end
         for _, argument in ipairs(t.typeargs or {}) do inner[argument.typearg] = nil end
         names = inner
      end
      for _, field in ipairs({ "def", "t", "elements", "keys", "values", "args", "rets", "constraint" }) do
         local found = watched_typevar(t[field], names, seen, depth)
         if found then return found end
      end
      for _, field in ipairs({ "typevals", "tuple", "types", "typeargs" }) do
         for _, member in ipairs(t[field] or {}) do
            local found = watched_typevar(member, names, seen, depth)
            if found then return found end
         end
      end
      for _, field in ipairs({ "fields", "meta_fields" }) do
         for _, member in pairs(t[field] or {}) do
            local found = watched_typevar(member, names, seen, depth)
            if found then return found end
         end
      end
      return nil
   end
   local function check_nested_types(decl, origin, depth)
      depth = depth + 1
      if depth > 64 then error("secret check exceeds 64 authored type nesting levels") end
      local def, names = decl.def, {}
      if type(def) ~= "table" then return end
      if def.typename == "generic" then
         for _, argument in ipairs(def.typeargs or {}) do names[argument.typearg] = true end
         def = def.t
      end
      if def.typename ~= "record" and def.typename ~= "interface" then return end
      for _, name in ipairs(field_names(def)) do
         local member = def.fields[name]
         if member and member.typename == "typedecl" then
            local mentioned = next(names) and watched_typevar(member.def, names, {}, 0)
            if mentioned then
               local at = member.def.y and member.def or origin
               report(at, "type declaration", ": nested type " .. name .. " mentions type parameter " .. mentioned
                  .. " of its generic owner, which Teal 0.24.8 does not instantiate there")
            end
            check_nested_types(member, origin, depth)
         end
      end
   end
   local function runtime_refinement(typ, origin, seen, depth)
      if type(typ) ~= "table" then error("compiler omitted runtime refinement type") end
      seen, depth = seen or {}, (depth or 0) + 1
      if depth > 64 or seen[typ] then error("cannot resolve bounded runtime refinement type") end
      seen[typ] = true
      if typ.typename == "nominal" then
         if not is_secret(typ) and not is_data(typ) then
            runtime_refinement(assert(typ.found, "compiler omitted resolved refinement type"), origin, seen, depth)
         end
      elseif typ.typename == "typedecl" or typ.typename == "generic" then runtime_refinement(typ.def or typ.t, origin, seen, depth)
      elseif typ.typename == "union" then
         for _, member in ipairs(typ.types) do runtime_refinement(member, origin, seen, depth) end
      elseif typ.typename == "record" then
         local predicate = typ.meta_fields and typ.meta_fields.__is
         if predicate and predicate.macroexp then check_record_refinement(predicate, typ, origin) end
      end
      seen[typ] = nil
   end
   local function prepass(node, seen, depth)
      if type(node) ~= "table" or seen[node] then return end
      visit()
      depth = (depth or 0) + 1
      if depth > 256 then error("secret check exceeds 256 AST nesting levels") end
      seen[node] = true
      node = original_is[node] or node
      if node.kind == "macroexp" or node.kind == "local_macroexp" then
         error("unsupported secret-check macro expansion")
      end
      if node.kind == "newtype" then
         authored_type_macros(node.newtype, {}, 0)
         check_nested_types(node.newtype, node, 0)
      end
      if node.kind == "function" or node.kind == "local_function" or node.kind == "global_function" or node.kind == "record_function" then
         implemented_functions[tostring(node.y) .. ":" .. tostring(node.x)] = true
      end
      if node.kind == "op" and node.op.op == "@funcall" then
         for _, argument in ipairs(node.e2 or {}) do call_arguments[argument] = true end
      end
      if node.kind == "statements" then
         for _, child in ipairs(node) do
            if child.kind == "op" and child.op.op == "@funcall" then statement_calls[child] = true end
         end
      end
      if node.kind == "assignment" then
         for _, variable in ipairs(node.vars or {}) do
            if variable.kind == "variable" then mutated_names[variable.tk] = true end
            if variable.kind == "op" and (variable.op.op == "." or variable.op.op == "@index") then
               local base = variable.e1
               while base.kind == "paren" do base = base.e1 end
               -- Teal accepts only a constant key on `_G`: a field name or a
               -- string literal in any quoting.
               local key = variable.e2.conststr or variable.e2.tk
               if base.kind == "variable" and base.tk == "_G" and key then mutated_names[key] = true end
            end
         end
      elseif node.kind == "global_function" and node.name then
         mutated_names[node.name.tk] = true
      elseif node.kind == "global_declaration" then
         for _, variable in ipairs(node.vars or {}) do mutated_names[variable.tk] = true end
      end
      if node.kind == "literal_table" and node.expected and is_metatable(resolve(node.expected)) then
         for _, item in ipairs(node) do
            local value_type = item.key and item.key.conststr == "__index" and node_type(item.value)
            if value_type and record_table(value_type) then mark_inert(item.value) end
         end
      end
      syntax_children(node, function(child) prepass(child, seen, depth) end)
   end
   local prepared, preparation_error = pcall(prepass, result.ast, {})
   if not prepared then fail_closed(preparation_error) end

   local function invalidate_loop_writes(node, state, seen, depth)
      if type(node) ~= "table" or seen[node] then return end
      visit()
      depth = (depth or 0) + 1
      if depth > 256 then error("secret check exceeds 256 AST nesting levels") end
      seen[node] = true
      if node.kind == "function" or node.kind == "local_function" or node.kind == "global_function" or node.kind == "record_function" then return end
      if node.kind == "assignment" then
         for _, variable in ipairs(node.vars or {}) do
            if variable.kind == "variable" then
               local binding = state.bindings[variable.tk]
               if binding then clear_facts(state, binding) end
            end
         end
      end
      syntax_children(node, function(child) invalidate_loop_writes(child, state, seen, depth) end)
   end

   local function native_policy(typ)
      if not typ or not typ.f or not typ.y or not typ.x then return nil end
      local path = canonical(typ.f)
      local prefix = root .. "/"
      if path:sub(1, #prefix) ~= prefix then return nil end
      local policies = call_policy.declarations[path:sub(#prefix + 1)]
      return policies and policies[tostring(typ.y) .. ":" .. tostring(typ.x)]
   end
   local function callable(typ)
      if typ == nil then return nil end
      local resolved = resolve(typ)
      if resolved.typename == "function" then return resolved end
      return nil
   end
   local function native_guard(node, state)
      while node and node.kind == "paren" do node = node.e1 end
      if not node or node.kind ~= "variable" then return nil end
      local binding = state.bindings[node.tk]
      if binding then return not mutated_names[node.tk] and binding.guard or nil end
      if mutated_names[node.tk] or not guard_symbols[node.tk] then return nil end
      local policy = native_policy(callable(value_type(node)))
      return policy and policy.symbol == node.tk and policy.guard or nil
   end
   local function local_binding(node, state)
      while node and node.kind == "paren" do node = node.e1 end
      if not node or node.kind ~= "variable" then return nil end
      local binding = state.bindings[node.tk]
      if binding and not binding.global then return binding end
   end
   -- typ overrides the node's type for one result of a call expanded into
   -- several values, which shares the call node but not its type.
   local function is_opaque(node, state, typ)
      local binding = local_binding(node, state)
      if binding and binding.opaque then return true end
      typ = typ or value_type(node)
      return opaque(typ) or (typ ~= nil and variable_typed(typ))
   end
   local function proved(node, state, fact)
      local binding = local_binding(node, state)
      local proof = binding and state[fact][binding]
      return proof ~= nil and (not binding.captured_write or proof == state.context.epoch)
   end
   -- A public value is accessible in every context (canaccessvalue is true for
   -- any non-secret value), so publicness also proves value access. It never
   -- proves indexing: a table is not a secret value, yet may hold secret contents.
   local function accessible(node, state, indexed, typ)
      if not is_opaque(node, state, typ) then return true end
      if indexed then return proved(node, state, "indexing") end
      return proved(node, state, "value") or proved(node, state, "public")
   end
   local function require_access(node, state, operation, indexed, typ)
      if node and not accessible(node, state, indexed, typ) then reject(node, operation) end
   end
   -- Publicness implies value access, so one publicness check stands in for both:
   -- a separate access check would only repeat the diagnostic for one missing guard.
   local function require_public(node, state, operation)
      if not proved(node, state, "public") then
         report(node, operation, " requires a same-value `not issecretvalue(x)` proof", "public")
      end
   end
   local function require_kind(node, state, fact)
      if not proved(node, state, fact) then
         report(node, "type assertion", " requires a same-value `type(x) " .. (fact == "table" and "==" or "~=") .. " \"table\"` proof", "kind")
      end
   end
   local function permitted(policy, position)
      if not policy then return false end
      if policy.guard then return true end
      if policy.secret_vararg_from and position >= policy.secret_vararg_from then return true end
      for _, allowed in ipairs(policy.secret_parameters or {}) do
         if allowed == position then return true end
      end
      return false
   end
   -- Collects what each type variable of a trusted generic meets at the call.
   -- `variance` is "in" (the caller's value flows into the variable), "out" (the
   -- variable's value flows to the caller) or "both" (mutable storage).
   local flip = { ["in"] = "out", out = "in", both = "both" }
   local function mentions_typevar(t, seen, depth, own)
      if type(t) ~= "table" then return false end
      depth = (depth or 0) + 1
      if depth > 64 then error("type variable search exceeds type depth") end
      seen = seen or {}
      if seen[t] then return false end
      seen[t] = true
      if t.typename == "typevar" then return own == nil or own(t.typevar) end
      if t.typename == "nominal" then
         for _, value in ipairs(t.typevals or {}) do if mentions_typevar(value, seen, depth, own) then return true end end
         return false
      end
      for _, field in ipairs({ "elements", "keys", "values", "args", "rets" }) do
         if mentions_typevar(t[field], seen, depth, own) then return true end
      end
      for _, member in ipairs(t.tuple or {}) do if mentions_typevar(member, seen, depth, own) then return true end end
      for _, member in ipairs(t.types or {}) do if mentions_typevar(member, seen, depth, own) then return true end end
      for _, member in pairs(t.fields or {}) do if mentions_typevar(member, seen, depth, own) then return true end end
      return false
   end
   -- Returns false when the declared shape holds a type variable that the actual
   -- type could not be matched against, so the caller can fail closed.
   -- A use recorded beneath a function type is marked `callback`.
   -- `own`, when given, names the callee's own type variables; others do not bind.
   local function bind_uses(declared, actual, variance, uses, callback, depth, own)
      depth = (depth or 0) + 1
      if depth > 64 then error("generic binding exceeds type depth") end
      if declared == nil or actual == nil then return true end
      if declared.typename == "typevar" and own and not own(declared.typevar) then return true end
      if declared.typename == "typevar" then
         local list = uses[declared.typevar] or {}
         uses[declared.typevar] = list
         list[#list + 1] = { type = actual, variance = variance, callback = callback }
         return true
      end
      -- Pairs members as leak does, so a vararg cannot hide a use leak compares.
      if declared.typename == "tuple" or actual.typename == "tuple" then
         local declared_tuple = declared.typename == "tuple" and declared or { tuple = { declared } }
         local actual_tuple = actual.typename == "tuple" and actual or { tuple = { actual } }
         local ok = true
         for index = 1, math.max(#actual_tuple.tuple, #declared_tuple.tuple) do
            local member, value = member_of(declared_tuple, index), member_of(actual_tuple, index)
            if member and value and not bind_uses(member, value, variance, uses, callback, depth, own) then ok = false end
         end
         return ok
      end
      -- Two instantiations of one generic record (`Box<U>` given `Box<T>`) bind
      -- their type arguments pairwise; fields are writable, so both ways.
      if declared.typename == "nominal" and actual.typename == "nominal" and declared.typevals and actual.typevals
         and declared.found ~= nil and declared.found == actual.found and #declared.typevals == #actual.typevals then
         local ok = true
         for index, value in ipairs(declared.typevals) do
            if not bind_uses(value, actual.typevals[index], "both", uses, callback, depth, own) then ok = false end
         end
         return ok
      end
      local d, a = resolve(declared), resolve(actual)
      if d.typename == "typevar" then return bind_uses(d, actual, variance, uses, callback, depth, own) end
      -- nil holds nothing the callee could read, write or call.
      if not mentions_typevar(d, nil, nil, own) or a.typename == "nil" then return true end
      local function all(pairs_list, beneath)
         local ok = true
         for _, pair in ipairs(pairs_list) do
            if not bind_uses(pair[1], pair[2], pair[3], uses, callback or beneath, depth, own) then ok = false end
         end
         return ok
      end
      if d.typename == "union" then
         -- Only a type variable member binds; a member that nests one is a use
         -- check_bindings cannot see.
         local modelled = true
         for _, member in ipairs(d.types) do
            local resolved = resolve(member)
            if resolved.typename == "typevar" then bind_uses(member, actual, variance, uses, callback, depth, own)
            elseif mentions_typevar(resolved, nil, nil, own) then modelled = false end
         end
         return modelled
      end
      local actual_elements = (a.typename == "array" or ((a.typename == "record" or a.typename == "interface") and a.elements)) and { a.elements }
         or a.typename == "tupletable" and a.types
      if (d.typename == "array" or d.typename == "tupletable") and actual_elements then
         local list = {}
         if d.typename == "array" then
            for _, member in ipairs(actual_elements) do list[#list + 1] = { d.elements, member, "both" } end
         else
            for index, member in ipairs(d.types) do
               list[#list + 1] = { member, actual_elements[index] or actual_elements[#actual_elements], "both" }
            end
         end
         return all(list)
      elseif d.typename == "map" and a.typename == "map" then
         return all({ { d.keys, a.keys, "both" }, { d.values, a.values, "both" } })
      elseif d.typename == "map" and actual_elements then
         local list = { { d.keys, { typename = "integer" }, "both" } }
         for _, member in ipairs(actual_elements) do list[#list + 1] = { d.values, member, "both" } end
         return all(list)
      elseif d.typename == "function" and a.typename == "function" then
         return all({ { d.args, a.args, flip[variance] }, { d.rets, a.rets, variance } }, true)
      end
      return false
   end
   -- A trusted generic moves values between its uses of a type variable: every
   -- value flowing in must be admitted by every caller-owned position it can flow
   -- out to, under the full (untrusted) rule. The call result is not such a
   -- position: Teal infers it from context (`print(select(1, h))` binds `any`),
   -- so `transport` re-types the result from the arguments and its destination
   -- is checked wherever the call expression flows.
   -- `complete` is false when an argument hid a type variable use from
   -- bind_uses; a Secret bound to any type variable may then reach it unseen.
   local function check_bindings(uses, complete)
      for name, list in pairs(uses) do
         local bound = false
         for _, source in ipairs(list) do
            if source.variance ~= "out" then
               -- A callback's own values do not bind: the callee still chooses
               -- what it passes the callback.
               bound = bound or not source.callback
               if not complete and deep(source.type) then
                  return "generic transport: Secret is bound beside a parameter shape the checker does not model"
               end
               if not complete and holds_variable(source.type) then
                  return "generic transport: type variable is bound beside a parameter shape the checker does not model"
               end
               for _, destination in ipairs(list) do
                  if destination ~= source and destination.variance ~= "in" then
                     local reason = leak(source.type, destination.type, { trusted = false, fresh = false })
                     if reason then return "generic transport: " .. reason end
                  end
               end
            end
         end
         -- A type variable no argument binds is the callee's choice: it reaches
         -- its uses as an unknown type variable, not as a transported value.
         -- A callback result flows into it and was checked above as a source.
         if not bound then
            for _, destination in ipairs(list) do
               local reason = destination.variance ~= "in" and leak({ typename = "typevar", typevar = name }, destination.type, { trusted = false, fresh = false })
               if reason then return "generic transport: " .. reason end
            end
         end
      end
   end
   local function trusted_declaration(typ)
      if stdlib_name(typ) then return true end
      local path = typ.f and canonical(typ.f)
      return path ~= nil and path:sub(1, #root + 11) == root .. "/types/wow/"
   end
   -- The generated `Mixin(target, ...)`, by declaration identity: it copies
   -- its other arguments' fields into its target.
   local function mixin_declaration(typ)
      local global = env.globals.Mixin and env.globals.Mixin.t
      global = global and global.typename == "generic" and global.t or global
      return typ ~= nil and global ~= nil and typ.f == global.f and typ.y == global.y and typ.x == global.x
   end
   -- Trusted sinks that only read their arguments.
   local reading_sinks = { print = true, tostring = true }
   -- A callee runs the builtin `name` only when it names that unshadowed
   -- global, never assigned: `local p = print; p = f` keeps print's declared
   -- type.
   local function builtin_callee(callee, state, name)
      return callee.kind == "variable" and callee.tk == name
         and not state.bindings[name] and not mutated_names[name]
   end
   -- Puts the receiver a callee takes first in `values`: a method call's object,
   -- or a callable record itself, which Teal passes to its `__call`.
   local function receive(callee, callee_type, values)
      if callee.kind == "op" and callee.op.op == ":" then
         table.insert(values, 1, { node = callee.e1, type = value_type(callee.e1) })
      elseif callee_type and resolve(callee_type).typename ~= "function" and resolve(callee_type).typename ~= "poly" then
         local meta = resolve(callee_type).meta_fields
         if not (meta and meta.__call) then error("unsupported callable shape") end
         table.insert(values, 1, { node = callee, type = callee_type })
      end
   end
   -- The signature `callee` runs, with its receiver put first in `values`.
   -- Teal types a `rawget` it runs as the index `t[k]` and records no identity
   -- for it, so that function is "indexing"; nil means no identity was recorded.
   local function run_signature(callee, values)
      local callee_type = value_type(callee)
      local signature = lookup(callee_types, callee) or callable(callee_type)
      if not signature then
         return callee_type and stdlib_member_name(resolve(callee_type)) == "rawget" and "indexing" or nil
      end
      receive(callee, callee_type, values)
      signature = resolve(signature)
      if signature.typename ~= "function" then error("unsupported callable shape") end
      return signature
   end
   -- What a call runs: `signature` (or `indexing` for a `rawget`) on `values`,
   -- in parameter order. pcall and xpcall run their first argument on the values
   -- after their own; `protector` is then the pcall or xpcall signature and
   -- `arguments` the values passed to it. Nil when Teal recorded no callee
   -- identity.
   local function call_shape(node)
      local arguments = argument_values(node.e2)
      local signature = run_signature(node.e1, arguments)
      if signature == nil then return nil end
      if signature == "indexing" then return { indexing = true, values = arguments } end
      local builtin = stdlib_name(signature)
      if builtin ~= "pcall" and builtin ~= "xpcall" then return { signature = signature, values = arguments, callee = node.e1 } end
      local target = node.e2 and node.e2[1]
      if not target then error("missing protected call target") end
      local values = {}
      for index = (builtin == "pcall" and 1 or 2) + 1, #arguments do values[#values + 1] = arguments[index] end
      local target_signature = run_signature(target, values)
      if target_signature == nil then error("compiler omitted protected call target identity") end
      local shape = { values = values, protector = signature, arguments = arguments, callee = target }
      if target_signature == "indexing" then shape.indexing = true else shape.signature = target_signature end
      return shape
   end
   -- A copy of tuple `t` holding `members`.
   local function retupled(t, members)
      local out = setmetatable({}, getmetatable(t))
      for key, value in pairs(t) do out[key] = value end
      out.tuple = members
      return out
   end
   local transported = {}
   transport = function(node)
      local cached = transported[node]
      if cached ~= nil then return cached or nil end
      transported[node] = false
      local shape = call_shape(node)
      -- Teal's own type is the result of a `rawget` it ran: the indexed type.
      if not shape or not shape.signature then return nil end
      local signature, values, protector = shape.signature, shape.values, shape.protector
      -- Teal infers an authored generic's result from context too
      -- (`print(id(x))` binds `any`), so a type variable it carries is
      -- re-typed as well; a Secret passed to it was rejected as an argument.
      local trusted = trusted_declaration(signature)
      local own = not trusted and instance_variable or nil
      if stdlib_name(signature) then signature = stdlib_declaration(signature) end
      local uses = {}
      for position, value in ipairs(values) do
         local parameter = member_of(signature.args, position)
         -- An unmodeled shape is rejected by check_arguments; keep Teal's type.
         if parameter and not bind_uses(parameter, value.type, "in", uses, nil, nil, own)
            and ((trusted and deep(value.type)) or holds_variable(value.type)) then return nil end
      end
      local inferred = lookup(expression_types, node)
      if inferred and protector then
         -- Teal types a protected call as `boolean` before the target's results.
         if inferred.typename ~= "tuple" then error("unsupported protected call result") end
         local rest = {}
         for index = 2, #inferred.tuple do rest[#rest + 1] = inferred.tuple[index] end
         inferred = retupled(inferred, rest)
      end
      if inferred then bind_uses(signature.rets, inferred, "out", uses, nil, nil, own) end
      local bindings, carried = {}, false
      for name, list in pairs(uses) do
         local sources, fallback = {}, nil
         for _, use in ipairs(list) do
            if use.variance == "out" then fallback = fallback or use.type
            else
               sources[#sources + 1] = use.type
               carried = carried or (trusted and deep(use.type)) or holds_variable(use.type)
            end
         end
         bindings[name] = #sources == 1 and sources[1] or #sources > 1 and { typename = "union", types = sources } or fallback
      end
      -- Plain transports keep Teal's own result type.
      if not carried then return nil end
      local result = substitute(signature.rets, bindings, {}, 0)
      if protector then
         local members = { member_of(protector.rets, 1) }
         for _, member in ipairs(result.tuple) do members[#members + 1] = member end
         result = retupled(result, members)
      end
      transported[node] = result
      return result
   end
   -- Stdlib functions that hand arguments back to authored code untyped are not
   -- consuming sinks: from the position given, each argument is judged as
   -- reaching `any` under the full rule, not the method exemption. error's
   -- value reaches pcall/xpcall results and the handler, resume/yield values
   -- reach the other side; assert raises its message and on success returns
   -- every argument, though Teal declares neither.
   local reemits = { error = 1, assert = 2, ["coroutine.resume"] = 1, ["coroutine.yield"] = 1 }
   -- Callbacks whose results come back to authored code untyped: xpcall's
   -- handler as xpcall's results, a coroutine body as resume's. They are judged
   -- under the full rule.
   local reemitting_callbacks = { xpcall = 2, ["coroutine.create"] = 1 }
   -- A coroutine.wrap function hands the arguments of every call after the
   -- first to the coroutine.yield that resumes, as `any`.
   local reemitting_parameters = { ["coroutine.wrap"] = 1 }
   -- `first` is the parameter position values[1] fills.
   -- `reading` marks every argument as only read by the callee.
   local function check_arguments(signature, values, first, origin, reading)
      local policy = native_policy(signature)
      -- Stdlib and generated native declarations have trusted bodies: their type
      -- variables transport values. Authored generic type variables are sinks.
      local ctx = { trusted = trusted_declaration(signature), fresh = false }
      local builtin = stdlib_name(signature) or ""
      local mixin = mixin_declaration(signature)
      local reemitted_from, reemitting_callback = reemits[builtin], reemitting_callbacks[builtin]
      local reemitting_parameter = reemitting_parameters[builtin]
      if stdlib_name(signature) then signature = stdlib_declaration(signature) end
      local uses, complete = {}, true
      local last, count = values[#values], #values
      if last and last.variadic then count = math.max(count, #signature.args.tuple - first + 1) end
      for index = 1, count do
         local position = first + index - 1
         local value = value_at(values, index)
         local reemitted = reemitted_from ~= nil and position >= reemitted_from
         local parameter = reemitted and ANY or member_of(signature.args, position)
         if parameter and not permitted(policy, position) then
            local trusted = ctx.trusted and not reemitted and position ~= reemitting_callback
            -- A trusted callee's type variables are transports only in an argument
            -- whose every use of them check_bindings sees.
            local modelled = trusted and bind_uses(parameter, value.type, "in", uses)
            if trusted and not modelled then
               complete = false
               if deep(value.type) then
                  reject_flow(value.node or origin, "argument " .. position .. ": Secret passes a generic parameter shape the checker does not model")
               elseif holds_variable(value.type, nil, true) then
                  -- A record table nested in the argument is compared by the leak
                  -- below; one bound to a type variable, by check_bindings.
                  reject_flow(value.node or origin, "argument " .. position .. ": type variable passes a generic parameter shape the checker does not model")
               end
            elseif not ctx.trusted and holds_variable(value.type) and not bind_uses(parameter, value.type, "in", {}, nil, nil, instance_variable) then
               reject_flow(value.node or origin, "argument " .. position .. ": type variable passes a generic parameter shape the checker does not model")
            end
            local fresh_ctx = { trusted = trusted, fresh = value.node.kind == "literal_table" or value.node.kind == "function", transport = modelled and "expected" or nil,
               inert = reading, written = mixin and position == 1 or nil }
            local reason = value.expanded and leak(value.type, parameter, fresh_ctx) or flow_leak(value.node, value.type, parameter, fresh_ctx)
            if reason then reject_flow(value.node or origin, "argument " .. position .. ": " .. reason) end
            if position == reemitting_parameter then
               local callback = resolve(value.type)
               for _, body in ipairs(callback.typename == "union" and union_members(callback) or { callback }) do
                  for _, member in ipairs(body.typename == "poly" and body.types or { body }) do
                     local resolved = resolve(member)
                     for index, argument in ipairs(resolved.typename == "function" and resolved.args.tuple or {}) do
                        local leaked = leak(argument, ANY, { trusted = false, fresh = false })
                        if leaked then reject_flow(value.node or origin, "argument " .. position .. ": callback parameter " .. index .. ": " .. leaked) end
                     end
                  end
               end
            end
            -- Marks only parameters this comparison covered. Teal 0.24.8 gives a
            -- literal callback's parameters an expected type only when the
            -- parameter is a function type, so the callback_targets test decides
            -- nothing today; re-check it when upgrading Teal.
            if not value.expanded and value.node.kind == "function" and #callback_targets(resolve(parameter)) > 0 then
               for _, argument in ipairs(value.node.args or {}) do call_checked_parameters[argument] = true end
            end
         end
      end
      if ctx.trusted then
         -- An omitted optional or vararg argument is nil at runtime, so it binds
         -- the type variables of its parameter as an explicit `nil` does. A last
         -- argument with variadic results was compared at every later position.
         if not (last and last.variadic) then
            for position = first + #values, #signature.args.tuple do
               bind_uses(signature.args.tuple[position], { typename = "nil" }, "in", uses)
            end
         end
         local reason = check_bindings(uses, complete)
         if reason then reject_flow(origin, reason) end
      end
   end

   -- One value at a receiver that declares `expected`, or nothing for it.
   local function check_received(value, position, expected)
      local ctx = { trusted = false, fresh = false }
      local reason
      if expected then reason = leak(value.type, expected, ctx) else reason = dropped_leak(value.type, ctx) end
      if reason then reject_flow(value.node, "value " .. position .. ": " .. reason) end
   end
   -- Each return expression is walked against the declared result at its own
   -- position only, so the further results a call or `...` in last position
   -- expands to are compared here. A variadic one fills every later declared
   -- position, and past a fixed declaration its results still reach the caller.
   local function check_return(node, state)
      local declared = state.owner ~= result.ast and state.owner.rets
      if not declared then return end
      local values = argument_values(node.exps)
      local last, count = values[#values], #values
      if last and last.variadic then count = math.max(count, #declared.tuple + (declared.is_va and 0 or 1)) end
      for position = #node.exps + 1, count do
         check_received(value_at(values, position), position, member_of(declared, position))
      end
   end
   -- `local a, b: number, W.Secret<number> = g() as number`: a variable past the
   -- expressions keeps its result, and later results are dropped.
   local function check_declaration_tail(node)
      local values = argument_values(node.exps)
      for position, variable in ipairs(node.vars or {}) do
         local value = value_at(values, position)
         if value and value.past_cast then check_received(value, position, variable_type(variable)) end
      end
   end
   -- A generic `for` keeps three values: the iterator, then the two it passes
   -- to every call of itself.
   local function check_iterator_tail(node)
      local values = argument_values(node.exps)
      local iterator = values[1] and callable(values[1].type)
      for position = 2, 3 do
         local value = value_at(values, position)
         if value and value.past_cast then check_received(value, position, iterator and member_of(iterator.args, position - 1)) end
      end
   end

   local expression, condition, block, statement
   local function list(nodes, state)
      for _, node in ipairs(nodes or {}) do expression(node, state) end
   end
   local function guard_facts(node, state)
      if not node or node.kind ~= "op" or node.op.op ~= "@funcall" then return nil end
      local guard = native_guard(node.e1, state)
      if not guard then return nil end
      local facts = state_copy(state)
      local args = node.e2 or {}
      if guard ~= "values" and #args ~= 1 then return nil end
      if guard == "values" and #args == 0 then return nil end
      for _, argument in ipairs(args) do
         local binding = local_binding(argument, state)
         if not binding then return nil end
         facts[guard == "table" and "indexing" or guard == "secret" and "public" or "value"][binding] = state.context.epoch
      end
      return facts, guard
   end

   -- `type(x) == "table"` (or `~=`, or another kind literal) records the runtime
   -- kind of local `x` for the branch it proves. The `type` call itself needs
   -- access and publicness, checked by check_call.
   local function kind_facts(node, state)
      local call, literal = node.e1, node.e2
      if call.kind == "string" then call, literal = literal, call end
      if literal.kind ~= "string" or call.kind ~= "op" or call.op.op ~= "@funcall" or #(call.e2 or {}) ~= 1 then return nil end
      local callee = lookup(callee_types, call.e1)
      if not callee or stdlib_name(resolve(callee)) ~= "type" then return nil end
      local binding = local_binding(call.e2[1], state)
      if not binding then return nil end
      local matched, unmatched = state_copy(state), state_copy(state)
      if literal.conststr == "table" then
         matched.table[binding], unmatched.nontable[binding] = state.context.epoch, state.context.epoch
      else
         matched.nontable[binding] = state.context.epoch
      end
      if node.op.op == "==" then return matched, unmatched end
      return unmatched, matched
   end

   -- A cast can pass more values than the callee declares parameters, which Teal
   -- rejects without one. Only a body in view proves they are dropped: a
   -- `function(number)` variable may hold `function(n: number, ...: any)`, and a
   -- poly callee's overload was chosen by the cast's width, not the runtime count.
   local function check_surplus(values, signature, callee, state)
      local binding = local_binding(callee, state)
      if signature.args.is_va or (binding and binding.body and not mutated_names[binding.body.name.tk]) then return end
      local policy = native_policy(signature)
      local first, last = #signature.args.tuple + 1, values[#values]
      for index = first, math.max(#values, last and last.variadic and first or 0) do
         local value = value_at(values, index)
         if value and value.past_cast and not permitted(policy, index) then check_received(value, index) end
      end
   end

   local function check_call(node, state)
      local shape = call_shape(node)
      if not shape then error("compiler omitted callee identity") end
      local signature, values = shape.signature, shape.values
      if shape.protector then
         -- Only the builtin runs its target on the values after it: a variable
         -- typed like pcall may hold any function, which receives them as `any`.
         if not builtin_callee(node.e1, state, stdlib_name(shape.protector)) then
            check_arguments(shape.protector, shape.arguments, 1, node)
            return
         end
         if stdlib_name(shape.protector) == "xpcall" then
            check_arguments(shape.protector, { shape.arguments[2] }, 2, node)
         end
      end
      -- A protected call runs its target on `values`, so the rules below are
      -- those of the target.
      if shape.indexing then
         -- `rawget(t, k)`, direct or protected, gets the indexing rules; the
         -- result keeps the indexed type and is compared where it flows.
         if #values < 2 then error("unsupported rawget shape") end
         check_surplus(values, { args = { tuple = { values[1], values[2] } } }, shape.callee or node.e1, state)
         -- Its result is typed as the member `t[k]`, which a caller may write.
         local body = record_table(values[1].type)
         if body then reject_flow(values[1].node, "argument 1: " .. record_table_reason(body)) end
         require_access(values[1].node, state, "operation indexing", true, values[1].type)
         require_access(values[2].node, state, "table key", nil, values[2].type)
         return
      end
      local builtin = stdlib_name(signature)
      -- The rules below that skip or relax the argument checks hold only for
      -- the builtin itself, not for a variable typed like it.
      local genuine = builtin ~= nil and builtin_callee(shape.callee, state, builtin)
      -- `type(x)` and `assert(x)` observe x as a condition does, so a value
      -- typed by a type variable needs the same proof as a Secret.
      if builtin == "type" and genuine then
         for _, value in ipairs(values) do
            if is_opaque(value.node, state, value.type) then
               require_public(value.node, state, "type observation")
            end
         end
         return
      end
      if builtin == "assert" and values[1] then
         require_access(values[1].node, state, "truthiness", nil, values[1].type)
      end
      if builtin == "select" and genuine and values[1] and values[1].node.conststr == "#" then
         -- select("#", ...) observes only the count. Any other string selector
         -- is coerced to an index (`select("1", s)` returns s).
         return
      end
      check_surplus(values, signature, shape.callee, state)
      check_arguments(signature, values, 1, node, genuine and reading_sinks[builtin] or nil)
   end

   condition = function(node, state)
      node = original_is[node] or node
      if node.kind == "paren" then return condition(node.e1, state) end
      if node.kind == "op" then
         local op = node.op.op
         if op == "not" then
            local positive, negative = condition(node.e1, state)
            return negative, positive
         elseif op == "and" then
            local left_true, left_false = condition(node.e1, state)
            local right_true, right_false = condition(node.e2, left_true)
            return right_true, join({ left_false, right_false }, state)
         elseif op == "or" then
            local left_true, left_false = condition(node.e1, state)
            local right_true, right_false = condition(node.e2, left_false)
            return join({ left_true, right_true }, state), right_false
         end
      end
      expression(node, state)
      require_access(node, state, "truthiness")
      local facts, guard = guard_facts(node, state)
      if facts then
         if guard == "secret" then return state_copy(state), facts end
         return facts, state_copy(state)
      end
      if node.kind == "op" and (node.op.op == "==" or node.op.op == "~=") then
         local call, boolean = node.e1, node.e2
         if call.kind == "boolean" then call, boolean = boolean, call end
         if boolean.kind == "boolean" then
            local compared_facts, compared_guard = guard_facts(call, state)
            if compared_guard == "secret" then
               local positive = (boolean.tk == "true") == (node.op.op == "==")
               if positive then return state_copy(state), compared_facts end
               return compared_facts, state_copy(state)
            end
         end
         local positive, negative = kind_facts(node, state)
         if positive then return positive, negative end
      end
      return state_copy(state), state_copy(state)
   end

   local function function_body(node, state)
      local inner = state_copy(state)
      inner.owner = node
      inner.context = { epoch = 0 }
      for _, fact in ipairs(FACTS) do
         if not VALUE_FACTS[fact] then inner[fact] = {}
         else for binding in pairs(inner[fact]) do if binding.captured_write then inner[fact][binding] = nil end end end
      end
      for _, parameter in ipairs(node.args or {}) do
         inner.bindings[parameter.tk] = { owner = node, opaque = opaque(variable_type(parameter)), captured_write = mutated_names[parameter.tk] }
      end
      block(node.body, inner)
   end

   -- Handle lookalikes. A cast can re-type a native handle as an authored record
   -- whose same-named member is plain where the native's is opaque. Routes
   -- through a type lacking the member (W.Object, an empty interface, userdata)
   -- hide it from the structural comparison, so the target's own members are
   -- compared against every generated handle the value could be at runtime.
   local generated_declarations = root .. "/types/wow/wow.d.tl"
   local declaration_files = {}
   local function declared_in(t)
      local file = t.f
      if type(file) ~= "string" then return nil end
      local canonical_file = declaration_files[file]
      if not canonical_file then canonical_file = canonical(file); declaration_files[file] = canonical_file end
      return canonical_file
   end
   local function generated_type(t) return declared_in(t) == generated_declarations end
   local declaration_root = root .. "/types/wow/"
   local function authored_type(t)
      if t.f == "stdlib.d.tl" then return false end
      local file = declared_in(t)
      return file ~= nil and file:sub(1, #declaration_root) ~= declaration_root
   end
   local function record_like(t) return t.typename == "record" or t.typename == "interface" end
   local function generated_bases(t)
      local bases = {}
      for _, base in ipairs(t.interface_list or {}) do
         local resolved = resolve(base)
         if record_like(resolved) and generated_type(resolved) then bases[resolved] = true end
      end
      return bases
   end
   -- Judged by where the member is declared, since inheriting copies it: an
   -- authored override of a generated name is an authored contract.
   generated_handle_member = function(member)
      return callable_member(member) and generated_type(member)
   end
   -- Generated interfaces by member name, with each one's flattened ancestors.
   local handle_members
   local function handles_with(name)
      if not handle_members then
         handle_members = {}
         local module = env.modules.wow
         module = module and module.typename == "typedecl" and module.def or module
         if not module or not module.fields then error("compiler omitted declaration module identity") end
         for _, decl in pairs(module.fields) do
            local handle = decl.typename == "typedecl" and decl.def
            if handle and handle.typename == "interface" and generated_type(handle) then
               local ancestors = generated_bases(handle)
               for member in pairs(handle.fields or {}) do
                  local list = handle_members[member] or {}
                  list[#list + 1] = { handle = handle, ancestors = ancestors }
                  handle_members[member] = list
               end
            end
         end
         for _, list in pairs(handle_members) do
            table.sort(list, function(left, right) return left.handle.y < right.handle.y end)
         end
      end
      return handle_members[name] or {}
   end
   -- Overloads whose arities overlap may both accept a call, since Teal admits
   -- nil in every parameter: Teal types such a call by the first that fits, or
   -- by a later one when the expected result type rules the first out, and
   -- overload pairing assumes overloads accepting the same calls agree. Both
   -- are sound for generated declarations only while such overloads agree on
   -- the opacity of every value the native hands back: each result, and each
   -- parameter of a callback it is passed. Metamethods (`__call`) included.
   local function check_generated_overloads()
      local module = env.modules.wow
      module = module and module.typename == "typedecl" and module.def or module
      if not module or not module.fields then error("compiler omitted declaration module identity") end
      local function arity(f) return f.min_arity or 0, f.args.is_va and math.huge or #f.args.tuple end
      local function differ(x, y)
         return opaque(x) ~= opaque(y) or (x ~= nil and deep(x, nil, nil, function_member)) ~= (y ~= nil and deep(y, nil, nil, function_member))
      end
      -- The first position of `left` and `right` whose opacity differs.
      local function disagreement(left, right, what)
         for index = 1, math.max(#left.tuple, #right.tuple) do
            if differ(member_of(left, index), member_of(right, index)) then return what .. " " .. index end
         end
         return nil
      end
      local function compare_overloads(left, right)
         local reason = disagreement(left.rets, right.rets, "result")
         if reason then return reason end
         for position = 1, math.max(#left.args.tuple, #right.args.tuple) do
            local x, y = member_of(left.args, position), member_of(right.args, position)
            for _, left_callback in ipairs(x and callback_targets(resolve(x)) or {}) do
               for _, right_callback in ipairs(y and callback_targets(resolve(y)) or {}) do
                  reason = disagreement(left_callback.args, right_callback.args, "parameter " .. position .. " callback parameter")
                  if reason then return reason end
               end
            end
         end
         return nil
      end
      local seen = {}
      local function scan(t, path)
         if seen[t] then return end
         seen[t] = true
         for _, members in ipairs({ { t.fields or {}, field_names(t) }, { t.meta_fields or {}, field_names({ fields = t.meta_fields }) } }) do
            local fields, names = members[1], members[2]
            for _, name in ipairs(names) do
               local field = fields[name]
               if not field then
               elseif field.typename == "typedecl" and field.def and record_like(field.def) then
                  scan(field.def, path .. "." .. name)
               elseif field.typename == "poly" then
                  for i = 1, #field.types do
                     for j = i + 1, #field.types do
                        local left, right = resolve(field.types[i]), resolve(field.types[j])
                        local left_min, left_max = arity(left)
                        local right_min, right_max = arity(right)
                        if math.max(left_min, right_min) <= math.min(left_max, right_max) then
                           local reason = compare_overloads(left, right)
                           if reason then
                              error("generated overloads of " .. path .. "." .. name .. " disagree on the opacity of " .. reason)
                           end
                        end
                     end
                  end
               end
            end
         end
      end
      scan(module, "wow")
   end
   -- `function M:Name` bodies by owner type: the record itself (`Mixin(f, R)`) or
   -- a table whose inferred record type is M (`local M = {}`). Other modules'
   -- nodes carry no types, so there only top-level methods of a top-level record
   -- declared in the same module count.
   local method_bodies = {}
   local function add_body(owner, name)
      if not owner or not name then return end
      local bodies = method_bodies[owner] or {}
      bodies[name] = true
      method_bodies[owner] = bodies
   end
   local function collect_bodies(node, seen, depth)
      if type(node) ~= "table" or seen[node] then return end
      visit()
      depth = (depth or 0) + 1
      if depth > 256 then error("secret check exceeds 256 AST nesting levels") end
      seen[node] = true
      node = original_is[node] or node
      if node.kind == "record_function" and node.fn_owner.kind == "type_identifier" then
         local owner = lookup(expression_types, node.fn_owner)
         add_body(owner and resolve(owner), node.name and node.name.tk)
      end
      syntax_children(node, function(child) collect_bodies(child, seen, depth) end)
   end
   local function collect_module_bodies(ast)
      local records = {}
      for _, statement in ipairs(ast) do
         local value = (statement.kind == "local_type" or statement.kind == "global_type") and statement.value
         local decl = value and value.kind == "newtype" and value.newtype
         if decl and decl.typename == "typedecl" and statement.var and statement.var.tk then
            local def = resolve(decl)
            if record_like(def) then records[statement.var.tk] = def end
         end
      end
      for _, statement in ipairs(ast) do
         local owner = statement.kind == "record_function" and statement.fn_owner
         if owner and owner.kind == "type_identifier" then add_body(records[owner.tk], statement.name and statement.name.tk) end
      end
   end
   do
      local collected, collection_error = pcall(function()
         collect_bodies(result.ast, {})
         for name, loaded in pairs(env.loaded) do
            if name ~= filename and loaded.ast and authored_type({ f = name }) then collect_module_bodies(loaded.ast) end
         end
      end)
      if not collected then fail_closed(collection_error) end
   end
   local function mixin_call(node)
      while node.kind == "paren" do node = node.e1 end
      if node.kind ~= "op" or node.op.op ~= "@funcall" then return nil end
      if not mixin_declaration(lookup(callee_types, node.e1)) then return nil end
      return node
   end
   -- Members of `target` supplied by a checked body of a mixin argument of a
   -- syntactic `Mixin(x, ...)` operand. A separate mixin table's member must
   -- itself flow into the target's member.
   local function mixin_bodies(operand, target)
      local call = mixin_call(operand)
      local exempt = {}
      if not call then return exempt end
      for index = 2, #call.e2 do
         local argument = call.e2[index]
         local typ = lookup(expression_types, argument)
         local owner = typ and resolve(typ)
         for name in pairs(owner and method_bodies[owner] or {}) do
            local member = target.fields and target.fields[name]
            if member then
               if owner == target then exempt[name] = true
               elseif owner.fields and owner.fields[name]
                  and not leak(owner.fields[name], member, { trusted = false, fresh = false, cast = true }) then
                  exempt[name] = true
               end
            end
         end
      end
      return exempt
   end
   local function lookalike_cast(source, target)
      if not record_like(target) or not authored_type(target) then return false end
      if next(generated_bases(target)) or target.is_userdata then return true end
      if source.is_userdata then return true end
      if not record_like(source) then return false end
      if generated_type(source) then return true end
      -- Any authored source: an interface chain of authored bases can hide a
      -- handle. Candidates are narrowed by admits.
      return authored_type(source)
   end
   -- A value held by an authored type entered through a compared cast, so each of
   -- the type's members admits the value's same-named member. An empty interface
   -- admits every generated type.
   local admitted = {}
   local function admits(source, handle)
      local cache = admitted[source] or {}
      admitted[source] = cache
      if cache[handle] == nil then
         local ok = true
         for name, member in pairs(source.fields or {}) do
            local field = handle.fields and handle.fields[name]
            if not field or invariant(field, member, { trusted = false, fresh = false, cast = true }, 0, name) then
               ok = false
               break
            end
         end
         cache[handle] = ok
      end
      return cache[handle]
   end
   -- A table built here, bare or through setmetatable, is not a handle.
   local function fresh_table(node)
      while node.kind == "paren" do node = node.e1 end
      if node.kind == "literal_table" then return true end
      if node.kind ~= "op" or node.op.op ~= "@funcall" or stdlib_name(lookup(callee_types, node.e1)) ~= "setmetatable" then return false end
      local first = node.e2[1]
      while first and first.kind == "paren" do first = first.e1 end
      return first ~= nil and first.kind == "literal_table"
   end
   -- The runtime value is a subtype of every generated type both sides name.
   local function runtime_bases(source, target)
      local bases = generated_bases(target)
      if record_like(source) and generated_type(source) then bases[source] = true
      elseif record_like(source) then
         for base in pairs(generated_bases(source)) do bases[base] = true end
      end
      return bases
   end
   local function check_lookalike(node, source, target, exempt)
      if fresh_table(node.e1) or not lookalike_cast(source, target) then return end
      local authored_source = record_like(source) and authored_type(source)
      local inherited_from = generated_bases(target)
      local bases = runtime_bases(source, target)
      local names = {}
      for name in pairs(target.fields or {}) do names[#names + 1] = name end
      table.sort(names)
      for _, name in ipairs(names) do
         local member = target.fields[name]
         local inherited = false
         for base in pairs(inherited_from) do if base.fields and base.fields[name] == member then inherited = true end end
         if not inherited and not exempt[name] then
            for _, candidate in ipairs(handles_with(name)) do
               local reachable = not authored_source or admits(source, candidate.handle)
               for base in pairs(bases) do
                  if candidate.handle ~= base and not candidate.ancestors[base] then reachable = false end
               end
               local reason = reachable and invariant(candidate.handle.fields[name], member,
                  { trusted = false, fresh = false, cast = true }, 0, "lookalike member " .. name)
               if reason then
                  report(node.e2, "unwrap target", ": " .. reason .. " (as on " .. type_text(candidate.handle) .. ")")
                  return
               end
            end
         end
      end
   end

   -- `x as T` from an opaque `x` unwraps one Secret arm: T must equal its
   -- payload. When the other arms differ in kind (SecretData holds scalars and
   -- tables), a same-value `type(x)` proof selects the arm, so a public table with
   -- secret contents never unwraps as Atom and a scalar never unwraps as a table.
   local function check_unwrap_opaque(node, state, operand, target, operand_type, at)
      local payloads, plain = arms(operand_type)
      local declared
      for _, candidate in ipairs(payloads) do
         if same_type(candidate, target) then declared = candidate; break end
      end
      if not declared then
         report(node.e2, "unwrap target", ": " .. at .. type_text(target) .. " is not the declared payload of " .. type_text(operand_type))
         return
      end
      local wants_table, other_table, other_scalar = table_capable(declared, true), false, false
      for _, list in ipairs({ payloads, plain }) do
         for _, member in ipairs(list) do
            if member ~= declared and resolve(member).typename ~= "nil" then
               if table_capable(member, true) then other_table = true else other_scalar = true end
            end
         end
      end
      if is_secret_table(declared) then
         require_access(operand, state, "type assertion", true)
         if other_scalar or other_table then
            -- From SecretData: the Blizzard chain canaccessvalue, issecretvalue,
            -- type(x) == "table", canaccesstable.
            require_public(operand, state, "type assertion")
            require_kind(operand, state, "table")
         end
         return
      end
      require_public(operand, state, "type assertion")
      if wants_table and other_scalar then require_kind(operand, state, "table")
      elseif not wants_table and other_table then require_kind(operand, state, "nontable") end
   end
   -- A cast into an opaque type re-types an existing value, so a table or
   -- function must already be exactly the payload of some arm: `{number} as
   -- Secret<SecretTable>` would let later writes put secrets into a plain alias.
   -- Scalars carry no alias. Sources of unknown shape are judged by check_unknown_source.
   local function unknown_shape(t)
      if is_data(t) then return true end
      local resolved = resolve(t)
      local name = resolved.typename
      if name == "any" or name == "typevar" or name == "typearg" then return true end
      if name ~= "union" then return false end
      for _, member in ipairs(resolved.types) do
         if unknown_shape(member) then return true end
      end
      return false
   end
   local function check_wrap(node, operand_type, target)
      local payloads, plain = arms(target)
      local function wraps(t)
         local name = resolve(t).typename
         if is_data(t) or name == "any" or name == "typevar" or name == "typearg" or name == "nil" or scalar_target(t) then return true end
         for _, list in ipairs({ payloads, plain }) do
            for _, member in ipairs(list) do
               if same_type(t, member) then return true end
            end
         end
         return false
      end
      if wraps(operand_type) then return end
      -- Unknown arms enter conservatively; every other arm must still wrap.
      local resolved = resolve(operand_type)
      for _, member in ipairs(resolved.typename == "union" and resolved.types or { operand_type }) do
         if not wraps(member) then
            report(node.e2, "unwrap target", ": " .. type_text(member) .. " is not the payload of " .. type_text(target))
            return
         end
      end
   end
   -- A cast from a value of unknown shape (any, Data, a type variable) cannot
   -- show that an existing table or function was built for secrets, so an opaque
   -- position that accepts writes would alias plain storage:
   -- `(t as any) as {MaybeSecret<number>}` admits secret writes into `{number}`.
   -- Top-level and result positions only read. `W.SecretData` and
   -- `W.Secret<W.SecretTable>` are the conservative entry for unknown payloads.
   local function writable_secret(t, seen, depth)
      depth = (depth or 0) + 1
      if depth > 64 then error("cast target exceeds type depth") end
      if is_secret(t) then
         local inner = payload(t)
         return not is_secret_table(inner) and writable_secret(inner, seen, depth)
      end
      if is_data(t) then return nil end
      local resolved = resolve(t)
      if seen[resolved] then return nil end
      seen[resolved] = true
      local name = resolved.typename
      if name == "union" then
         for _, member in ipairs(resolved.types) do
            local position = writable_secret(member, seen, depth)
            if position then return position end
         end
      elseif name == "array" or name == "map" or name == "tupletable" then
         if deep(resolved) then return "container element" end
      elseif name == "function" then
         if deep(resolved.args) then return "function parameter" end
         -- A returned container is a writable alias too.
         for _, result in ipairs(resolved.rets and resolved.rets.tuple or {}) do
            local position = writable_secret(result, seen, depth)
            if position then return "function result " .. position end
         end
      elseif record_like(resolved) then
         if resolved.elements and deep(resolved.elements) then return "container element" end
         local names = {}
         for field_name in pairs(resolved.fields or {}) do names[#names + 1] = field_name end
         table.sort(names)
         for _, field_name in ipairs(names) do
            local field = resolved.fields[field_name]
            if callable_member(field) then
               -- Generated method contracts are the native's own; the value is
               -- not authored storage. Lookalikes reached from `any` are a documented limit.
               if not generated_type(resolved) and deep(field) then
                  local member = resolve(field)
                  local position = member.typename == "function" and writable_secret(member, seen, depth)
                  if position then return "method " .. field_name .. " " .. position:gsub("^function ", "") end
               end
            elseif deep(field) then return "field " .. field_name end
         end
      end
      return nil
   end
   local function check_unknown_source(node, operand_type, target)
      if not unknown_shape(operand_type) then return false end
      local position = writable_secret(target, {})
      if position then
         report(node.e2, "unwrap target", ": " .. position .. " of " .. type_text(target)
            .. " is opaque; a value of unknown shape enters only as W.SecretData or W.Secret<W.SecretTable>")
         return true
      end
      return false
   end
   -- One position of a cast. `operand_type` is the value at that position; proofs
   -- live on the operand node, so a result of a call expanded past the first has
   -- none and only the rules that need no proof can pass it.
   local function check_cast_position(node, state, operand_type, target, position)
      -- Casting to the identical type changes nothing, whatever the type holds.
      if same_type(operand_type, target) then return end
      local operand = node.e1
      local at = position > 1 and "value " .. position .. ": " or ""
      if is_opaque(operand, state, operand_type) and opaque(operand_type) then
         check_unwrap_opaque(node, state, operand, target, operand_type, at)
         return
      end
      -- A target that is itself type-variable-typed or opaque keeps the rules a
      -- Secret needs; any other target would read a possible Secret as plain.
      if variable_typed(operand_type) and not variable_typed(target) and not opaque(target) then
         require_public(operand, state, "type assertion")
      end
      if check_unknown_source(node, operand_type, target) then return end
      if opaque(target) then check_wrap(node, operand_type, target) end
      local source, resolved_target = resolve(operand_type), resolve(target)
      local exempt = position == 1 and mixin_bodies(operand, resolved_target) or {}
      if table_capable(target, false) and not table_capable(operand_type, true) and not source.is_userdata then
         report(node.e2, "unwrap target", ": " .. at .. type_text(operand_type) .. " holds no table")
      elseif deep(operand_type) or deep(target) or (holds_variable(operand_type) and not scalar_target(target)) then
         -- Containers are mutable aliases, so a cast compares them invariantly in
         -- both directions: `plain as {MaybeSecret<T>}` would admit secret writes.
         -- A type variable read as a scalar takes the publicness proof above;
         -- anywhere else it may hold a Secret the target types plain.
         local reason = leak(operand_type, target, { trusted = false, fresh = false, cast = true, exempt = exempt })
         if reason then report(node.e2, "unwrap target", ": " .. at .. reason); return end
      end
      check_lookalike(node, source, resolved_target, exempt)
   end
   local function check_unwrap(node, state)
      if not node.e2.casttype then error("compiler omitted cast target") end
      local targets = cast_targets(node)
      check_cast_position(node, state, value_type(node.e1), targets.tuple[1], 1)
      local results = runtime_results(node.e1)
      -- Teal types a chunk's `...` as strings; the client passes the addon name
      -- and the addon's table, so past the name there is no type to judge.
      if not results or (node.e1.kind == "..." and state.owner == result.ast) then return end
      local count = #targets.tuple
      if targets.is_va then count = math.max(count, #results.tuple) end
      for position = 2, count do
         local actual = member_of(results, position)
         if actual then check_cast_position(node, state, actual, member_of(targets, position), position) end
      end
   end

   -- A record table's parameter-typed fields are shared by every instance
   -- through `__index`, so writing one would type the value for every
   -- instantiation at once; reading one is rejected too, since a container read
   -- from it is written in place. A method is read to be called with an
   -- instance, which the call compares; assigning one is a write.
   -- The member is judged at its declared type: Teal gives a read fresh type
   -- variables (`T@12`), as it does a generic referenced by name. Teal 0.24.8
   -- indexes a record only by a member name or string literal.
   local assigned_members = {}
   local function check_record_table(node)
      local base = node_type(node.e1)
      if base and base.typename == "tuple" then base = base.tuple[1] end
      local body = base and record_table(base)
      if not body then return end
      local key = node.op.op == "@index" and node.e2.conststr or node.op.op ~= "@index" and node.e2.tk
      if not key then error("compiler indexed a record table by a computed key") end
      local member = body.fields and body.fields[key]
      if member == nil then error("compiler omitted record table member " .. key) end
      if not holds_variable(member) or (callable_member(member) and not assigned_members[node]) then return end
      report(node, "record table member", ": " .. record_table_name(body) .. "." .. key
         .. " has a type-parameter type every instance shares; read and write it through an instance")
   end
   -- A method body defined inside a function can capture values of that
   -- function's instantiation (an outer `T`, or `self`'s) and return them at
   -- the record's parameter types, which every instance shares.
   local function check_record_method(node, state)
      if state.owner == result.ast then return end
      local owner = lookup(expression_types, node.fn_owner)
      if owner == nil then error("compiler omitted record method owner identity") end
      local body = record_table(owner)
      if not body then return end
      local name = node.name and node.name.tk
      local member = name and body.fields and body.fields[name]
      if member and not holds_variable(member) then return end
      report(node, "record table method", ": " .. record_table_name(body) .. "." .. tostring(name)
         .. " is defined inside a function, where it can capture values at the record's type parameters; define it at chunk level")
   end

   expression = function(node, state)
      if not node then return end
      node = original_is[node] or node
      visit()
      if node.kind == "variable" then
         local binding = local_binding(node, state)
         if binding and binding.opaque and not opaque(value_type(node)) then
            -- Teal narrowed an opaque binding (for example after `v == 1`);
            -- reading it as plain is a flow that needs a same-value publicness proof.
            local scalar, marker = scalar_target(value_type(node))
            if not scalar or marker then reject_flow(node, "narrowed Secret is not a scalar")
            elseif not proved(node, state, "public") then
               reject_flow(node, "narrowed Secret lacks same-value publicness proof")
            end
         elseif not binding then
            -- Teal keeps an assignment's narrowing of a global across calls that
            -- may rewrite it, and globals carry no proofs, so a narrowed read of an
            -- opaque global is never public.
            local global = (not state.bindings[node.tk]) or state.bindings[node.tk].global
            local declared = global and env.globals[node.tk]
            if declared and declared.t and opaque(declared.t) and not opaque(value_type(node)) then
               reject_flow(node, "narrowed global Secret has no publicness proof")
            end
         end
      end
      if node.kind == "function" then function_body(node, state); return end
      if node.kind == "paren" then expression(node.e1, state); return end
      if node.kind == "op" then
         local op = node.op.op
         if op == "and" or op == "or" then
            local positive, negative = condition(node.e1, state)
            expression(node.e2, op == "and" and positive or negative)
            return
         elseif op == "not" then condition(node.e1, state); return end
         expression(node.e1, state)
         if node.e2 and node.e2.kind ~= "cast" then expression(node.e2, state) end
         if op == "==" or op == "~=" or op == "<" or op == ">" or op == "<=" or op == ">=" then
            require_access(node.e1, state, "comparison"); require_access(node.e2, state, "comparison")
         elseif op == "is" then
            runtime_refinement(node.e2.casttype, node)
            if is_opaque(node.e1, state) then
               local scalar, marker = scalar_target(node.e2.casttype)
               if not scalar or marker then reject(node.e2, "non-scalar narrowing")
               else require_public(node.e1, state, "type narrowing") end
            end
         elseif op == "as" then
            check_unwrap(node, state)
         elseif op == "@index" then
            require_access(node.e1, state, "operation indexing", true); require_access(node.e2, state, "table key")
            check_record_table(node)
         elseif op == "." or op == ":" then
            require_access(node.e1, state, "operation indexing", true)
            check_record_table(node)
         elseif op == "@funcall" then
            require_access(node.e1, state, "operation call")
            check_call(node, state)
            if not native_guard(node.e1, state) then state.context.epoch = state.context.epoch + 1 end
         elseif op == "#" or op == "-" or op == "~" or op == "+" or op == "*" or op == "/" or op == "//" or op == "%" or op == "^"
            or op == ".." or op == "&" or op == "|" or op == "<<" or op == ">>" then
            require_access(node.e1, state, "operation " .. op); require_access(node.e2, state, "operation " .. op)
         else error("unsupported secret-check operator: " .. tostring(op))
         end
         return
      elseif node.kind == "literal_table" then
         for _, item in ipairs(node) do
            expression(item.key, state); require_access(item.key, state, "table key")
            expression(item.value, state)
         end
         -- The table's own type receives what a cast in last position passes on:
         -- with no expected type, or a generic parameter's, nothing else does.
         local tail = node[#node]
         if tail and tail.key_parsed == "implicit" and tail.value.kind == "op" and tail.value.op.op == "as" then
            local reason = literal_leak(node, node_type(node), { trusted = false, fresh = true })
            if reason then reject_flow(node, reason) end
         end
         return
      elseif node.kind == "expression_list" then list(node, state); return
      elseif node.kind ~= "variable" and node.kind ~= "identifier" and node.kind ~= "argument"
         and node.kind ~= "string" and node.kind ~= "number" and node.kind ~= "integer"
         and node.kind ~= "boolean" and node.kind ~= "nil" and node.kind ~= "..."
         and node.kind ~= "cast" and node.kind ~= "type_identifier" and node.kind ~= "pragma" then
         error("unsupported secret-check expression: " .. tostring(node.kind))
      end
   end

   local function scoped(body, state)
      local nested = state_copy(state)
      local outcome = block(body, nested)
      if outcome then outcome.bindings = state.bindings end
      return outcome
   end
   local function iterator_result_opaque(typ, depth)
      depth = (depth or 0) + 1
      if depth > 64 then error("cannot resolve bounded iterator result identity") end
      if typ == nil then error("compiler omitted iterator identity") end
      if typ.typename == "tuple" then typ = typ.tuple[1] end
      local resolved = resolve(typ)
      if resolved.typename == "function" then return opaque(resolved.rets.tuple[1]) end
      if resolved.typename == "poly" or resolved.typename == "union" then
         for _, member in ipairs(resolved.types) do if iterator_result_opaque(member, depth) then return true end end
      end
      return false
   end
   -- `pairs(t)` and `next, t` yield the keys of a table the caller could already
   -- index, which needed canaccesstable when it came from a Secret; a key typed
   -- SecretData still needs its own proofs before any operation.
   local function table_iteration(node)
      local callee = node.kind == "op" and node.op.op == "@funcall" and lookup(callee_types, node.e1)
      if callee then return stdlib_name(resolve(callee)) == "pairs" end
      local typ = node.kind == "variable" and value_type(node)
      return typ and stdlib_member_name(resolve(typ)) == "next" or false
   end
   -- Assigning a public scalar to an opaque local proves the binding public, so
   -- `if issecretvalue(a) then a = 1 end` leaves `a` public on both paths. Only
   -- a positional expression whose type is exactly a scalar or nil counts: `any`
   -- (handler payloads) and unions may still carry a secret at runtime. A local
   -- declaration needs no rule: Teal narrows `local a: MaybeSecret<T> = 1` to T.
   local PUBLIC_SCALARS = { number = true, integer = true, string = true, boolean = true, ["nil"] = true }
   local function store_public(binding, value, state)
      if binding.global or not binding.opaque or not value then return end
      local typ = node_type(value)
      if typ and typ.typename == "tuple" then typ = typ.tuple[1] end
      if typ and PUBLIC_SCALARS[resolve(typ).typename] then state.public[binding] = state.context.epoch end
   end
   statement = function(node, state)
      local kind = node.kind
      if kind == "local_declaration" or kind == "global_declaration" then
         list(node.exps, state)
         check_declaration_tail(node)
         local aliases = {}
         for index in ipairs(node.vars or {}) do aliases[index] = native_guard(node.exps and node.exps[index], state) end
         for index, variable in ipairs(node.vars or {}) do
            state.bindings[variable.tk] = { owner = state.owner, guard = aliases[index], opaque = opaque(variable_type(variable)),
               captured_write = mutated_names[variable.tk], global = kind == "global_declaration" }
         end
      elseif kind == "assignment" then
         list(node.exps, state)
         -- Lua leaves the order of a multiple assignment unspecified, so a target
         -- named twice gets no proof from either value.
         local targets = {}
         for _, variable in ipairs(node.vars or {}) do
            if variable.kind == "variable" then targets[variable.tk] = (targets[variable.tk] or 0) + 1 end
         end
         -- Lua evaluates every target subexpression before assigning, so
         -- `a, t[a as number] = 1, 2` indexes with the old `a`: check targets first.
         for _, variable in ipairs(node.vars or {}) do
            if variable.kind ~= "variable" then
               assigned_members[variable] = true
               expression(variable, state)
            end
         end
         for index, variable in ipairs(node.vars or {}) do
            local binding = variable.kind == "variable" and state.bindings[variable.tk]
            if binding then
               if binding.owner ~= state.owner then binding.captured_write = true end
               clear_facts(state, binding)
               if targets[variable.tk] == 1 then store_public(binding, node.exps and node.exps[index], state) end
            end
         end
      elseif kind == "local_function" or kind == "global_function" or kind == "record_function" then
         if node.name and node.name.tk and kind == "local_function" then state.bindings[node.name.tk] = { owner = state.owner, body = node } end
         if kind == "record_function" then check_record_method(node, state) end
         function_body(node, state)
      elseif kind == "if" then
         local remaining, outcomes = state_copy(state), {}
         local has_else = false
         for _, branch in ipairs(node.if_blocks) do
            local positive, negative
            if branch.exp then positive, negative = condition(branch.exp, remaining)
            else positive, has_else = remaining, true end
            local outcome = scoped(branch.body, positive)
            if outcome then outcomes[#outcomes + 1] = outcome end
            if negative then remaining = negative end
         end
         if not has_else then outcomes[#outcomes + 1] = remaining end
         return join(outcomes, state)
      elseif kind == "while" then
         invalidate_loop_writes(node.body, state, {})
         local positive = condition(node.exp, state)
         local outcome = scoped(node.body, positive)
         return join(outcome and { state, outcome } or { state }, state)
      elseif kind == "repeat" then
         invalidate_loop_writes(node.body, state, {})
         local outcome = block(node.body, state_copy(state)) or state_copy(state)
         condition(node.exp, outcome)
         outcome.bindings = state.bindings
         return join({ state, outcome }, state)
      elseif kind == "do" then return scoped(node.body, state)
      elseif kind == "fornum" or kind == "forin" then
         expression(node.from, state); expression(node.to, state); expression(node.step, state); list(node.exps, state)
         if kind == "fornum" then
            require_access(node.from, state, "numeric for bound")
            require_access(node.to, state, "numeric for bound")
            require_access(node.step, state, "numeric for step")
         elseif iterator_result_opaque(node_type(node.exps[1])) and not table_iteration(node.exps[1]) then
            reject(node.exps[1], "generic for nil test")
         end
         if kind == "forin" then
            check_iterator_tail(node)
            state.context.epoch = state.context.epoch + 1
         end
         local inner = state_copy(state)
         invalidate_loop_writes(node.body, inner, {})
         if node.var then inner.bindings[node.var.tk] = { owner = state.owner } end
         for _, variable in ipairs(node.vars or {}) do inner.bindings[variable.tk] = { owner = state.owner } end
         local outcome = scoped(node.body, inner)
         return join(outcome and { state, outcome } or { state }, state)
      elseif kind == "return" then list(node.exps, state); check_return(node, state); return nil
      elseif kind == "break" or kind == "goto" or kind == "label" then
         -- Loop exits are not function exits. Clear proofs rather than using a
         -- break/goto as evidence that a failed guard cannot reach later code.
         for _, fact in ipairs(FACTS) do state[fact] = {} end
      elseif kind == "local_type" or kind == "global_type" then
         -- A local type shadows a global of the same name; without a binding its
         -- reads would be checked against the global's declared type.
         if kind == "local_type" and node.var and node.var.tk then state.bindings[node.var.tk] = { owner = state.owner } end
         return state
      else expression(node, state) end
      return state
   end
   block = function(body, state)
      for _, node in ipairs(body or {}) do
         state = statement(node, state)
         if not state then return nil end
      end
      return state
   end

   -- Opaque-flow rule at every point where Teal compared an expression against an
   -- expected type. Call arguments are compared per callee by check_call, where
   -- native policy and stdlib transports are known.
   local literal_kinds = { string = true, number = true, integer = true, boolean = true, ["nil"] = true }
   local function flows(node, seen, depth)
      if type(node) ~= "table" or seen[node] then return end
      visit()
      depth = (depth or 0) + 1
      if depth > 256 then error("secret check exceeds 256 AST nesting levels") end
      seen[node] = true
      local expected = node.expected
      if node.kind and expected and not literal_kinds[node.kind] and not call_arguments[node]
         and not call_checked_parameters[node] and expected.typename ~= "boolean_context" then
         local actual = node_type(node)
         if actual == nil then error(string.format("compiler omitted expression identity at %d:%d", node.y or 0, node.x or 0)) end
         local ctx = { trusted = false, fresh = node.kind == "literal_table" or node.kind == "function", inert = inert_positions[node] }
         -- A literal callback's parameter receives the contextual type, so the
         -- flow runs from the expected type into the declared one. Teal assigns
         -- that type before inferring a generic callee, leaving its type
         -- variables unbound; check_arguments, which knows whether they are
         -- transports or sinks, owns a parameter it already compared.
         local reason
         if node.kind == "argument" then reason = leak(expected, actual, ctx)
         else reason = flow_leak(node, actual, expected, ctx) end
         if reason then reject_flow(node, reason) end
      end
      -- Where Teal compared a call against nothing (an unannotated local, a table
      -- item), it gave the call its own result type, which later checks read. A
      -- re-typed result must be admitted by that type. An argument of an
      -- overloaded callee (`table.insert`, `select`) has no expected type either,
      -- so it is checked here as well as by check_arguments.
      if not expected and node.kind == "op" and node.op.op == "@funcall" and not statement_calls[node] then
         local transported = transport(node)
         if transported then
            local inferred = lookup(expression_types, node)
            if inferred == nil then error(string.format("compiler omitted call result identity at %d:%d", node.y or 0, node.x or 0)) end
            local reason = leak(transported, inferred, { trusted = false, fresh = false })
            if reason then reject_flow(node, reason) end
         end
      end
      syntax_children(node, function(child) flows(child, seen, depth) end)
   end

   local nesting = 0
   local function bounded(fn)
      return function(...)
         visit()
         nesting = nesting + 1
         if nesting > 256 then error("secret check exceeds 256 operation nesting levels") end
         local first, second = fn(...)
         nesting = nesting - 1
         return first, second
      end
   end
   expression, condition, block, statement = bounded(expression), bounded(condition), bounded(block), bounded(statement)

   local ok, inspection_error = pcall(function()
      if not generated_overloads_checked then
         check_generated_overloads()
         generated_overloads_checked = true
      end
      local initial = { bindings = {}, owner = result.ast, context = { epoch = 0 } }
      for _, fact in ipairs(FACTS) do initial[fact] = {} end
      block(result.ast, initial)
      flows(result.ast, {})
   end)
   if not ok then fail_closed(inspection_error) end
   table.sort(diagnostics, function(a, b) return a.y == b.y and (a.x == b.x and a.message < b.message or a.x < b.x) or a.y < b.y end)
   for _, diagnostic in ipairs(diagnostics) do
      io.stderr:write(string.format("%s:%d:%d: %s\n", filename, diagnostic.y, diagnostic.x, diagnostic.message))
   end
   return #diagnostics == 0
end

local failed = false
for _, source in ipairs(sources) do
   local ok, passed = pcall(check_source, source)
   if not ok then
      local message = type(passed) == "table" and passed[FAILED_CLOSED] or tostring(passed)
      io.stderr:write(source .. ":1:1: secret check failed closed: " .. message .. "\n")
      passed = false
   end
   if not passed then failed = true end
end
if failed then os.exit(1) end
