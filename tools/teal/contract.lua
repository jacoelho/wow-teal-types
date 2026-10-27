-- This contract is shared by the consumer preflight and the secrecy checker.
return function(tl)
   if type(tl) ~= "table" or type(tl.version) ~= "function" or tl.version() ~= "0.24.8-wow.1"
      or type(tl.wow_capabilities) ~= "table" or tl.wow_capabilities.unknown ~= 1
      or tl.wow_capabilities.type_report ~= 1 or tl.wow_capabilities.lua_target ~= "5.1"
      or type(tl.typecodes) ~= "table" or tl.typecodes.PUBLIC_UNKNOWN ~= 0x04000000
      or type(tl.new_type_reporter) ~= "function" or type(tl.get_types) ~= "function" then
      error("paired Teal 0.24.8-wow.1 unknown/report capabilities required")
   end
   return "unknown=1;type_report=1;lua_target=5.1"
end
