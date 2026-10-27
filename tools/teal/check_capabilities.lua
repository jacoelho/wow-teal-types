local directory = assert(arg[0]:match("^(.*)[/\\]check_capabilities.lua$"), "cannot locate compiler contract")
local validate = assert(loadfile(directory .. "/contract.lua"))()
print(validate(require("tl")))
