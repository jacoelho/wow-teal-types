local type W = require("wow")

local panel: W.Frame = CreateFrame("Frame", "WowTealExampleFrame", UIParent)
panel:SetSize(220, 32)
panel:SetPoint("CENTER")

local label = panel:CreateFontString(nil, "OVERLAY", "GameFontNormal")
label:SetAllPoints(panel)
label:SetText("WoW Teal example")

-- Forward the possibly secret result to a native sink without inspecting it.
local health: W.NumberLike = UnitHealth("player")
local bar = CreateFrame("StatusBar", nil, panel) as W.StatusBar
bar:SetSize(220, 8)
bar:SetPoint("TOP", panel, "BOTTOM")
bar:SetStatusBarTexture("Interface\\Buttons\\WHITE8X8")
bar:SetValue(health)

-- Ordinary Lua arithmetic requires independent permission, secrecy and shape checks.
if canaccessvalue(health) and not issecretvalue(health) and health is number then
   local public_health: number = health
   print("WoW Teal public health sample:", public_health + 0)
end
