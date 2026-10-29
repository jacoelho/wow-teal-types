-- Inspect generated syntax, never addon statements. A type-only import is erased.
-- `tl run` has already loaded the compiler. A `require("tl")` here would make
-- `tl run` type-check all of tl.lua before this script starts.
local tl = package.loaded["tl"]
assert(type(tl) == "table" and type(tl.parse) == "function", "run with `tl run`")
local filename = assert(arg[1], "expected generated Lua path")
local input = assert(io.open(filename, "rb"))
local text = input:read(4 * 1024 * 1024 + 1)
input:close()
assert(#text <= 4 * 1024 * 1024, "generated Lua exceeds 4 MiB")
local ast, errors = tl.parse(text, filename, "lua")
assert(ast and #errors == 0, "cannot parse generated Lua")
local forbidden = { wow = true, native_values = true, wow_env = true, wow_secure_env = true, compat53 = true }
local fields = { "e1", "e2", "vars", "exps", "body", "args", "name", "fn_owner", "if_blocks", "exp", "var", "from", "to", "step", "key", "value" }
local seen, count = {}, 0
local function unparen(node)
   local depth = 0
   while node and node.kind == "paren" do
      depth = depth + 1
      assert(depth <= 256, "runtime import parentheses exceed syntax bounds")
      node = node.e1
   end
   return node
end
local function literal(node)
   node = unparen(node)
   if not node or node.kind ~= "string" then return nil end
   -- The parser owns this single string token. Decoding it cannot run a statement.
   return assert((loadstring or load)("return " .. node.tk, "@module-literal"))()
end
local function inspect(node, depth)
   if seen[node] then return end
   seen[node], count = true, count + 1
   assert(depth <= 256 and count <= 2000000, "runtime import scan exceeds syntax bounds")
   local called = unparen(node.e1)
   local require_call = called and called.kind == "variable" and called.tk == "require"
   if called and called.kind == "op" and (called.op.op == "." or called.op.op == "@index") then
      local receiver = unparen(called.e1)
      local name = called.op.op == "." and called.e2.tk or literal(called.e2)
      require_call = receiver and receiver.kind == "variable" and receiver.tk == "_G" and name == "require"
   end
   if node.kind == "op" and node.op.op == "@funcall" and require_call and node.e2 then
      local name = literal(node.e2[1])
      if forbidden[name] then
         io.stderr:write(string.format("%s:%d:%d: runtime declaration/compatibility import: %s\n", filename, node.y, node.x, name))
         os.exit(1)
      end
   end
   for _, child in ipairs(node) do if type(child) == "table" then inspect(child, depth + 1) end end
   for _, key in ipairs(fields) do if type(node[key]) == "table" then inspect(node[key], depth + 1) end end
end
inspect(ast, 0)
