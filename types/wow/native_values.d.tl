-- Static restriction markers; these do not assert runtime userdata carriers.
-- Native accessibility checks own permission to inspect the underlying values.
local record OpaqueNumber
   is userdata
end
local record OpaqueText
   is userdata
end
local record Values
   type OpaqueNumber = OpaqueNumber
   type OpaqueText = OpaqueText
   type NumberLike = number | OpaqueNumber
   type TextLike = string | OpaqueText
end
return Values
