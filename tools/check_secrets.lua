-- Teal's type report identifies opaque values; this pass supplies operation rules
-- that the compiler does not enforce. It is not data-flow analysis through any.
local tl = require("tl")
local environment, filename = arg[1], arg[2]
if not environment or not filename then
   io.stderr:write("usage: tl [include flags] run check_secrets.lua -- ENV SOURCE\n")
   os.exit(2)
end
if tl.version() ~= "0.24.8" then
   io.stderr:write("secret check requires Teal 0.24.8\n")
   os.exit(2)
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
if not root then error("cannot locate catalogue root from checker path") end
local marker_locations = {}
for file, names in pairs({
   [root .. "/types/wow/native_values.d.tl"] = { OpaqueNumber = true, OpaqueText = true },
   [root .. "/types/wow/wow.d.tl"] = { OpaqueValue = true },
}) do
   local source, count = assert(io.open(file, "r")), 0
   local line_number = 0
   for line in source:lines() do
      line_number = line_number + 1
      local name = line:match("^%s*local%s+record%s+(%w+)") or line:match("^%s*record%s+(%w+)")
      if names[name] then
         marker_locations[file .. ":" .. line_number] = name
         count = count + 1
      end
   end
   source:close()
   local expected = 0
   for _ in pairs(names) do expected = expected + 1 end
   if count ~= expected then error("opaque marker definitions changed: " .. file) end
end

local syntax_fields = { "e1", "e2", "vars", "exps", "body", "args", "name", "fn_owner", "if_blocks", "exp", "var", "from", "to", "step", "key", "value", "macrodef" }
local function syntax_children(node, consume)
   for _, child in ipairs(node) do if type(child) == "table" then consume(child) end end
   for _, field in ipairs(syntax_fields) do
      local child = node[field]
      if type(child) == "table" then consume(child) end
   end
end
local original_is, capture_error = {}, nil
local original_parse = tl.parse
tl.parse = function(input, source_name, parse_lang)
   local ast, errors, required = original_parse(input, source_name, parse_lang)
   if source_name == filename then
      -- The compiler replaces `is Record` nodes with expanded where bodies.
      -- Keep their original operands while retaining the same type objects,
      -- which this one strict check resolves before policy inspection.
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
local opts = { defaults = { feat_lax = "off", feat_arity = "on", gen_compat = "off", gen_target = "5.1" } }
if environment ~= "-" then opts.predefined_modules = { environment } end
local env, env_error = tl.new_env(opts)
if not env then io.stderr:write(tostring(env_error) .. "\n"); os.exit(1) end
env.report_types = true
-- The public report omits tuple expressions, including native getter calls.
-- Keep their first-result identity through the same compiler collector rather
-- than inferring it from a function signature or checking the source twice.
local tuple_results = {}
env.reporter = tl.new_type_reporter()
local get_collector = env.reporter.get_collector
env.reporter.get_collector = function(reporter, source_name)
   local collector = get_collector(reporter, source_name)
   local store_type = collector.store_type
   collector.store_type = function(y, x, typ)
      if source_name == filename and typ and typ.typename == "tuple" and typ.tuple[1] then
         tuple_results[y] = tuple_results[y] or {}
         tuple_results[y][x] = typ.tuple[1]
      end
      return store_type(y, x, typ)
   end
   return collector
end
local result, process_error = tl.process(filename, env)
tl.parse = original_parse
if not result then io.stderr:write(tostring(process_error) .. "\n"); os.exit(1) end
if capture_error then io.stderr:write(filename .. ":1:1: secret check failed closed: " .. tostring(capture_error) .. "\n"); os.exit(1) end
-- Marker identity must belong to the declarations actually loaded by Teal.
local native_environment = arg[3] or (environment == "wow_secure_env" and "wow_secure_env" or "wow_env")
if native_environment ~= "wow_env" and native_environment ~= "wow_secure_env" then
   error("unsupported native environment")
end
for _, name in ipairs({ "wow", "native_values", "wow_env", "wow_secure_env" }) do
   local actual = env.module_filenames[name]
   local required = name == "wow" or name == "native_values" or name == native_environment
   if (required and not actual) or (env.modules[name] and not actual)
      or (actual and canonical(actual) ~= root .. "/types/wow/" .. name .. ".d.tl") then
      io.stderr:write(filename .. ":1:1: declaration identity mismatch: " .. name .. "\n")
      os.exit(1)
   end
end
local failed = false
for _, loaded_name in ipairs(env.loaded_order) do
   local loaded = env.loaded[loaded_name]
   for _, category in ipairs({ "syntax_errors", "type_errors", "warnings" }) do
      for _, diagnostic in ipairs(loaded[category] or {}) do
         io.stderr:write(string.format("%s:%d:%d: %s%s\n", diagnostic.filename or loaded_name,
            diagnostic.y or 1, diagnostic.x or 1, category == "warnings" and "warning: " or "", diagnostic.msg))
         if category ~= "warnings" then failed = true end
      end
   end
end
if failed then os.exit(1) end
local report = assert(tl.get_types(result))
local positions = assert(report.by_pos[filename], "compiler omitted source type report")
local type_cache, type_visiting = {}, {}
local function opaque_type(id, depth)
   if not id then return false end
   if type_cache[id] ~= nil then return type_cache[id] end
   if (depth or 0) > 64 or type_visiting[id] then error("cannot resolve bounded opaque type identity") end
   local info = assert(report.types[id], "compiler omitted type identity")
   type_visiting[id] = true
   local opaque = info.file and info.y and marker_locations[canonical(info.file) .. ":" .. info.y] ~= nil
   if not opaque and info.ref then opaque = opaque_type(info.ref, (depth or 0) + 1) end
   if not opaque and info.t == tl.typecodes.UNION then
      for _, member in ipairs(info.types or {}) do
         if opaque_type(member, (depth or 0) + 1) then opaque = true; break end
      end
   end
   type_visiting[id] = nil
   type_cache[id] = not not opaque
   return not not opaque
end
local function expression_type(node)
   local row = node and positions[node.y]
   if node and (node.kind == "..." or (node.kind == "op" and node.op.op == "@funcall")) then
      local tuple_row = tuple_results[node.y]
      local first = tuple_row and tuple_row[node.x]
      if not first then error("compiler omitted call result identity") end
      return env.reporter:get_typenum(first)
   end
   return row and row[node.x]
end
local function opaque(node)
   return opaque_type(expression_type(node))
end
local function opaque_iterator_result(id, seen, depth)
   if not id then error("compiler omitted iterator identity") end
   seen, depth = seen or {}, (depth or 0) + 1
   if depth > 64 or seen[id] then error("cannot resolve bounded iterator result identity") end
   seen[id] = true
   local info = assert(report.types[id], "compiler omitted iterator result type")
   local result = false
   if info.ref then result = opaque_iterator_result(info.ref, seen, depth)
   elseif info.t == tl.typecodes.FUNCTION then
      result = info.rets and info.rets[1] and opaque_type(info.rets[1][1]) or false
   elseif info.t == tl.typecodes.POLY or info.t == tl.typecodes.UNION then
      for _, member in ipairs(info.types or {}) do
         if opaque_iterator_result(member, seen, depth) then result = true; break end
      end
   end
   seen[id] = nil
   return result
end
local function scalar_target(id, seen, depth)
   if not id then error("compiler omitted conversion type identity") end
   seen, depth = seen or {}, (depth or 0) + 1
   if depth > 64 or seen[id] then error("cannot resolve bounded scalar type identity") end
   seen[id] = true
   local info = assert(report.types[id], "compiler omitted conversion type")
   local scalar, marker = false, false
   if info.file and info.y and marker_locations[canonical(info.file) .. ":" .. info.y] then marker = true
   elseif info.ref then scalar, marker = scalar_target(info.ref, seen, depth)
   elseif info.t == tl.typecodes.UNION then
      scalar = true
      for _, member in ipairs(info.types or {}) do
         local member_scalar, member_marker = scalar_target(member, seen, depth)
         if not member_scalar and not member_marker then scalar, marker = false, false; break end
         marker = marker or member_marker
      end
   elseif info.t == tl.typecodes.NUMBER or info.t == tl.typecodes.INTEGER or info.t == tl.typecodes.STRING
      or info.t == tl.typecodes.BOOLEAN or info.t == tl.typecodes.NIL or info.t == tl.typecodes.ENUM then scalar = true
   end
   seen[id] = nil
   return scalar, marker
end

local diagnostics, reported, visits = {}, {}, 0
local function visit()
   visits = visits + 1
   if visits > 2000000 then error("secret check exceeds two million AST visits") end
end
local function reject(node, operation)
   local key = tostring(node.y) .. ":" .. tostring(node.x) .. ":" .. operation
   if not reported[key] then
      reported[key] = true
      diagnostics[#diagnostics + 1] = { y = node.y or 1, x = node.x or 1,
         message = "secret-value " .. operation .. " requires same-value native accessibility proof" }
   end
end
local function copy(map)
   local out = {}; for key, value in pairs(map) do out[key] = value end; return out
end
local function state_copy(state)
   return { bindings = copy(state.bindings), value = copy(state.value), indexing = copy(state.indexing), public = copy(state.public), owner = state.owner, context = state.context }
end
local function join(states, fallback)
   if #states == 0 then return nil end
   local out = state_copy(fallback or states[1])
   for _, facts in ipairs({ "value", "indexing", "public" }) do
      out[facts] = copy(states[1][facts])
      for binding in pairs(out[facts]) do
         for index = 2, #states do if not states[index][facts][binding] then out[facts][binding] = nil; break end end
      end
   end
   return out
end
local guard_names = { canaccessvalue = "value", canaccessallvalues = "values", canaccesstable = "table", issecretvalue = "secret" }
local mutated_names = {}
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
         local id = env.reporter:get_typenum(field)
         if opaque_type(id) then reject(origin or node, "record refinement field inspection")
         else
            local scalar, marker = scalar_target(id)
            if not scalar or marker then error("record refinement requires a public scalar discriminator") end
         end
      elseif op == "not" then predicate(node.e1, depth + 1)
      elseif op == "and" or op == "or" or op == "==" or op == "~=" or op == "<" or op == ">" or op == "<=" or op == ">=" then
         predicate(node.e1, depth + 1); predicate(node.e2, depth + 1)
      elseif op == "is" then
         predicate(node.e1, depth + 1)
         local scalar, marker = scalar_target(env.reporter:get_typenum(node.e2.casttype))
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
local function runtime_refinement(typ, origin, seen, depth)
   if type(typ) ~= "table" then error("compiler omitted runtime refinement type") end
   seen, depth = seen or {}, (depth or 0) + 1
   if depth > 64 or seen[typ] then error("cannot resolve bounded runtime refinement type") end
   seen[typ] = true
   if typ.typename == "nominal" then runtime_refinement(assert(typ.found, "compiler omitted resolved refinement type"), origin, seen, depth)
   elseif typ.typename == "typedecl" or typ.typename == "generic" then runtime_refinement(typ.def, origin, seen, depth)
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
   if node.kind == "newtype" then authored_type_macros(node.newtype, {}, 0) end
   if node.kind == "assignment" then
      for _, variable in ipairs(node.vars or {}) do
         if variable.kind == "variable" then mutated_names[variable.tk] = true end
         if variable.kind == "op" and (variable.op.op == "." or variable.op.op == "@index")
            and variable.e1.kind == "variable" and variable.e1.tk == "_G" then
            local key = variable.e2 and variable.e2.tk
            if key then mutated_names[key:gsub("^['\"]", ""):gsub("['\"]$", "")] = true end
         end
      end
   elseif node.kind == "global_function" and node.name then
      mutated_names[node.name.tk] = true
   elseif node.kind == "global_declaration" then
      for _, variable in ipairs(node.vars or {}) do mutated_names[variable.tk] = true end
   end
   syntax_children(node, function(child) prepass(child, seen, depth) end)
end
local prepared, preparation_error = pcall(prepass, result.ast, {})
if not prepared then
   io.stderr:write(filename .. ":1:1: secret check failed closed: " .. tostring(preparation_error) .. "\n")
   os.exit(1)
end

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
            if binding then state.value[binding], state.indexing[binding], state.public[binding] = nil, nil, nil end
         end
      end
   end
   syntax_children(node, function(child) invalidate_loop_writes(child, state, seen, depth) end)
end

local function native_guard(node, state)
   while node and node.kind == "paren" do node = node.e1 end
   if not node or node.kind ~= "variable" then return nil end
   local binding = state.bindings[node.tk]
   if binding then return not mutated_names[node.tk] and binding.guard or nil end
   return not mutated_names[node.tk] and guard_names[node.tk] or nil
end
local function local_binding(node, state)
   while node and node.kind == "paren" do node = node.e1 end
   if not node or node.kind ~= "variable" then return nil end
   local binding = state.bindings[node.tk]
   if binding and not binding.global then return binding end
end
local function is_opaque(node, state)
   local binding = local_binding(node, state)
   return binding and binding.opaque or opaque(node)
end
local function accessible(node, state, indexed)
   if not is_opaque(node, state) then return true end
   local binding = local_binding(node, state)
   local proof = binding and state[indexed and "indexing" or "value"][binding]
   return proof ~= nil and (not binding.captured_write or proof == state.context.epoch)
end
local function require_access(node, state, operation, indexed)
   if node and not accessible(node, state, indexed) then reject(node, operation) end
end
local function require_public(node, state)
   local binding = local_binding(node, state)
   local proof = binding and state.public[binding]
   if proof == nil or (binding.captured_write and proof ~= state.context.epoch) then
      reject(node, "scalar conversion publicness")
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
   end
   return state_copy(state), state_copy(state)
end

local function function_body(node, state)
   local inner = state_copy(state)
   inner.owner, inner.value, inner.indexing = node, {}, {}
   inner.context = { epoch = 0 }
   for binding in pairs(inner.public) do if binding.captured_write then inner.public[binding] = nil end end
   for _, parameter in ipairs(node.args or {}) do
      inner.bindings[parameter.tk] = { owner = node, opaque = opaque(parameter), captured_write = mutated_names[parameter.tk] }
   end
   block(node.body, inner)
end

expression = function(node, state)
   if not node then return end
   node = original_is[node] or node
   visit()
   if node.kind == "variable" then
      local binding = local_binding(node, state)
      if binding and binding.opaque and not opaque(node) then
         local scalar, marker = scalar_target(expression_type(node))
         if not scalar or marker then reject(node, "non-scalar flow conversion")
         else require_access(node, state, "scalar flow conversion"); require_public(node, state) end
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
            local scalar, marker = scalar_target(expression_type(node.e2))
            if not scalar or marker then reject(node.e2, "non-scalar narrowing")
            else require_access(node.e1, state, "type narrowing"); require_public(node.e1, state) end
         end
      elseif op == "as" then
         if is_opaque(node.e1, state) then
            local scalar, marker = scalar_target(expression_type(node.e2))
            if not scalar and not marker then reject(node.e2, "non-scalar assertion")
            elseif not marker then
               require_access(node.e1, state, "type assertion"); require_public(node.e1, state)
               local already_scalar, still_marker = scalar_target(expression_type(node.e1))
               if not already_scalar or still_marker then reject(node.e1, "scalar assertion shape") end
            end
         end
      elseif op == "@index" then
         require_access(node.e1, state, "indexing", true); require_access(node.e2, state, "table key")
      elseif op == "." or op == ":" then
         require_access(node.e1, state, "indexing", true)
      elseif op == "@funcall" then
         require_access(node.e1, state, "call")
         if not native_guard(node.e1, state) then state.context.epoch = state.context.epoch + 1 end
      elseif op == "#" or op == "-" or op == "~" or op == "+" or op == "*" or op == "/" or op == "//" or op == "%" or op == "^"
         or op == ".." or op == "&" or op == "|" or op == "<<" or op == ">>" then
         require_access(node.e1, state, "operator " .. op); require_access(node.e2, state, "operator " .. op)
      else error("unsupported secret-check operator: " .. tostring(op))
      end
      return
   elseif node.kind == "literal_table" then
      for _, item in ipairs(node) do
         expression(item.key, state); require_access(item.key, state, "table key")
         expression(item.value, state)
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
statement = function(node, state)
   local kind = node.kind
   if kind == "local_declaration" or kind == "global_declaration" then
      list(node.exps, state)
      local aliases = {}
      for index in ipairs(node.vars or {}) do aliases[index] = native_guard(node.exps and node.exps[index], state) end
      for index, variable in ipairs(node.vars or {}) do
         state.bindings[variable.tk] = { owner = state.owner, guard = aliases[index], opaque = opaque(variable),
            captured_write = mutated_names[variable.tk], global = kind == "global_declaration" }
      end
   elseif kind == "assignment" then
      list(node.exps, state)
      for _, variable in ipairs(node.vars or {}) do
         if variable.kind == "variable" then
            local binding = state.bindings[variable.tk]
            if binding then
               if binding.owner ~= state.owner then binding.captured_write = true end
               state.value[binding], state.indexing[binding], state.public[binding] = nil, nil, nil
            end
         else expression(variable, state) end
      end
   elseif kind == "local_function" or kind == "global_function" or kind == "record_function" then
      if node.name and node.name.tk and kind == "local_function" then state.bindings[node.name.tk] = { owner = state.owner } end
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
      elseif opaque_iterator_result(expression_type(node.exps[1])) then
         reject(node.exps[1], "generic for nil test")
      end
      if kind == "forin" then state.context.epoch = state.context.epoch + 1 end
      local inner = state_copy(state)
      invalidate_loop_writes(node.body, inner, {})
      if node.var then inner.bindings[node.var.tk] = { owner = state.owner } end
      for _, variable in ipairs(node.vars or {}) do inner.bindings[variable.tk] = { owner = state.owner } end
      local outcome = scoped(node.body, inner)
      return join(outcome and { state, outcome } or { state }, state)
   elseif kind == "return" then list(node.exps, state); return nil
   elseif kind == "break" or kind == "goto" or kind == "label" then
      -- Loop exits are not function exits. Clear proofs rather than using a
      -- break/goto as evidence that a failed guard cannot reach later code.
      state.value, state.indexing, state.public = {}, {}, {}
   elseif kind == "local_type" or kind == "global_type" then
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

local ok, inspection_error = pcall(block, result.ast, { bindings = {}, value = {}, indexing = {}, public = {}, owner = result.ast, context = { epoch = 0 } })
if not ok then
   io.stderr:write(filename .. ":1:1: secret check failed closed: " .. tostring(inspection_error) .. "\n")
   os.exit(1)
end
table.sort(diagnostics, function(a, b) return a.y == b.y and (a.x == b.x and a.message < b.message or a.x < b.x) or a.y < b.y end)
for _, diagnostic in ipairs(diagnostics) do
   io.stderr:write(string.format("%s:%d:%d: %s\n", filename, diagnostic.y, diagnostic.x, diagnostic.message))
end
if #diagnostics > 0 then os.exit(1) end
