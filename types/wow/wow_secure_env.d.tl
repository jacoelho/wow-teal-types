-- Generated public native environment; addon fragments add only owned globals.
local type W = require("wow")
local type CatalogueTable = {unknown: unknown}
global ACCOUNT_WIDE_FONT_COLOR: W.Global_ACCOUNT_WIDE_FONT_COLOR
global ACHIEVEMENT_COLOR: W.Global_ACHIEVEMENT_COLOR
global ACHIEVEMENT_COMPLETE_COLOR: W.Global_ACHIEVEMENT_COMPLETE_COLOR
global ACHIEVEMENT_INCOMPLETE_COLOR: W.Global_ACHIEVEMENT_INCOMPLETE_COLOR
global ACTIONBAR_HOTKEY_FONT_COLOR: W.Global_ACTIONBAR_HOTKEY_FONT_COLOR
global ADVENTURES_BUFF_BLUE: W.Global_ADVENTURES_BUFF_BLUE
global ADVENTURES_COMBAT_LOG_BLUE: W.Global_ADVENTURES_COMBAT_LOG_BLUE
global ADVENTURES_COMBAT_LOG_GREY: W.Global_ADVENTURES_COMBAT_LOG_GREY
global ADVENTURES_COMBAT_LOG_ORANGE: W.Global_ADVENTURES_COMBAT_LOG_ORANGE
global ADVENTURES_COMBAT_LOG_YELLOW: W.Global_ADVENTURES_COMBAT_LOG_YELLOW
global ADVENTURES_HEALING_GREEN: W.Global_ADVENTURES_HEALING_GREEN
global ARCHAEOLOGY_BROWN: W.Global_ARCHAEOLOGY_BROWN
global ARCHAEOLOGY_LIGHT_BROWN: W.Global_ARCHAEOLOGY_LIGHT_BROWN
global AREA_DESCRIPTION_FONT_COLOR: W.Global_AREA_DESCRIPTION_FONT_COLOR
global AREA_NAME_FONT_COLOR: W.Global_AREA_NAME_FONT_COLOR
global ARENA_NAME_FONT_COLOR: W.Global_ARENA_NAME_FONT_COLOR
global ARTIFACT_BAR_COLOR: W.Global_ARTIFACT_BAR_COLOR
global ARTIFACT_GOLD_COLOR: W.Global_ARTIFACT_GOLD_COLOR
global AZERITE_ESSENCE_COLOR: W.Global_AZERITE_ESSENCE_COLOR
global AZERITE_SUBTEXT_COLOR: W.Global_AZERITE_SUBTEXT_COLOR
global AbandonSkill: function(number): unknown...
global AbbreviateLargeNumbers: function(W.NumberLike, ? W.NumberAbbrevOptions): string
global AbbreviateNumbers: function(W.NumberLike, ? W.NumberAbbrevOptions): string
global AcceptAreaSpiritHeal: function()
global AcceptArenaTeam: function()
global AcceptBattlefieldPort: function(number, boolean): unknown...
global AcceptDuel: function(): unknown...
global AcceptGroup: function(): unknown...
global AcceptGuild: function()
global AcceptProposal: function(): unknown...
global AcceptQuest: function(): unknown...
global AcceptResurrect: function()
global AcceptSpellConfirmationPrompt: function(number): unknown...
global AcceptTrade: function(): unknown...
global AccumulateIf: function({unknown: unknown}, unknown): number
global AchievementCriteriaFont: W.Global_AchievementCriteriaFont
global AchievementDateFont: W.Global_AchievementDateFont
global AchievementDescriptionFont: W.Global_AchievementDescriptionFont
global AchievementFont_Small: W.Global_AchievementFont_Small
global AchievementPointsFont: W.Global_AchievementPointsFont
global AchievementPointsFontSmall: W.Global_AchievementPointsFontSmall
global AcknowledgeAutoAcceptQuest: function(): unknown...
global AcknowledgeSurvey: function(unknown): unknown...
global ActionBindsItem: function()
global ActionButton_HideOverlayGlow: function({unknown: unknown})
global ActionButton_ShowOverlayGlow: function({unknown: unknown})
global AddAutoQuestPopUp: function(number, string): unknown...
global AddChatWindowChannel: function(number, string): unknown...
global AddChatWindowMessages: function(number, string): unknown...
global AddSourceLocationExclude: function(string)
global Ambiguate: function(W.TextLike, string): string
global AntiAliasingSupported: function(): unknown...
global ArchaeologyGetIconInfo: function(number): unknown...
global ArchaeologyMapUpdateAll: function(number): number
global ArcheologyGetVisibleBlobID: function(number): number
global AreClassRolesSoftSuggestions: function(): boolean
global AreColorsEqual: function(W.colorRGBA, W.colorRGBA): boolean
global AreDangerousScriptsAllowed: function(): boolean
global AreTalentsLocked: function(): unknown...
global AreVector2DEqual: function(W.Vector2DMixin, W.Vector2DMixin): boolean
global AreVector3DEqual: function(W.Vector3DMixin, W.Vector3DMixin): boolean
global ArenaTeamDisband: function(number)
global ArenaTeamInviteByName: function(number, string)
global ArenaTeamLeave: function(number)
global ArenaTeamSetLeaderByName: function(number, string)
global ArenaTeamUninviteByName: function(number, string)
global ArtifactAppearanceSetHighlightFont: W.Global_ArtifactAppearanceSetHighlightFont
global ArtifactAppearanceSetNormalFont: W.Global_ArtifactAppearanceSetNormalFont
global AscendStop: function(): unknown...
global AssistUnit: function(? string, ? boolean)
global AttachGlyphToSpell: function(unknown): unknown...
global AttackTarget: function()
global AuraUtil: W.Global_AuraUtil
global AutoChooseCurrentGraphicsSetting: function(): unknown...
global AutoEquipCursorItem: function()
global AutoLootMailItem: function(unknown): unknown...
global AutoStoreGuildBankItem: function(number, number): unknown...
global BANK_BAG_BACKGROUND_COLOR: W.Global_BANK_BAG_BACKGROUND_COLOR
global BATTLENET_FONT_COLOR: W.Global_BATTLENET_FONT_COLOR
global BATTLEPET_ABILITY_LINK_COLOR: W.Global_BATTLEPET_ABILITY_LINK_COLOR
global BIND_TRADE_TOOLTIP_COLOR: W.Global_BIND_TRADE_TOOLTIP_COLOR
global BLACK_FONT_COLOR: W.Global_BLACK_FONT_COLOR
global BLUE_FONT_COLOR: W.Global_BLUE_FONT_COLOR
global BNAcceptFriendInvite: function(unknown): unknown...
global BNCheckBattleTagInviteToGuildMember: function(unknown): unknown...
global BNCheckBattleTagInviteToUnit: function(unknown): unknown...
global BNConnected: function(): boolean
global BNDeclineFriendInvite: function(unknown): unknown...
global BNFeaturesEnabled: function(): unknown...
global BNFeaturesEnabledAndConnected: function(): unknown...
global BNGetBlockedInfo: function(unknown): unknown...
global BNGetDisplayName: function(unknown): unknown...
global BNGetFOFInfo: function(boolean, boolean, number): number, string, boolean
global BNGetFriendIndex: function(number): number
global BNGetFriendInviteInfo: function(number): number, number, boolean, string | nil, number
global BNGetInfo: function(): number | nil, string, number, string, boolean, boolean, boolean
global BNGetNumBlocked: function(): unknown...
global BNGetNumFOF: function(unknown, unknown, unknown): unknown...
global BNGetNumFriendInvites: function(): unknown...
global BNGetNumFriends: function(): number, number, number, number
global BNGetSelectedBlock: function(): unknown...
global BNGetSelectedFriend: function(): unknown...
global BNInviteFriend: function(unknown): unknown...
global BNIsBlocked: function(unknown): unknown...
global BNIsFriend: function(unknown): unknown...
global BNIsSelf: function(unknown): unknown...
global BNRemoveFriend: function(unknown): unknown...
global BNRequestFOFInfo: function(unknown): unknown...
global BNRequestInviteFriend: function(unknown, unknown, unknown, unknown): unknown...
global BNSendFriendInvite: function(unknown, unknown): unknown...
global BNSendFriendInviteByID: function(unknown, unknown): unknown...
global BNSendVerifiedBattleTagInvite: function(): unknown...
global BNSetBlocked: function(unknown, unknown): unknown...
global BNSetFriendFavoriteFlag: function(number, boolean): unknown...
global BNSetFriendNote: function(number, string): unknown...
global BNSetSelectedBlock: function(unknown): unknown...
global BNSetSelectedFriend: function(unknown): unknown...
global BNSummonFriendByIndex: function(unknown): unknown...
global BNTokenFindName: function(unknown): unknown...
global BRIGHTBLUE_FONT_COLOR: W.Global_BRIGHTBLUE_FONT_COLOR
global BRONZE_MATERIAL_TEXT_COLOR: W.Global_BRONZE_MATERIAL_TEXT_COLOR
global BRONZE_MATERIAL_TITLETEXT_COLOR: W.Global_BRONZE_MATERIAL_TITLETEXT_COLOR
global BattlefieldMgrEntryInviteResponse: function(unknown, unknown): unknown...
global BattlefieldMgrExitRequest: function(unknown): unknown...
global BattlefieldMgrQueueInviteResponse: function(unknown, unknown): unknown...
global BattlefieldMgrQueueRequest: function(): unknown...
global BeginTrade: function()
global BindEnchant: function()
global BossEmoteNormalHuge: W.Global_BossEmoteNormalHuge
global BreakUpLargeNumbers: function(W.NumberLike, ? boolean): string
global BuyGuildBankTab: function(): unknown...
global BuyGuildCharter: function(string): unknown...
global BuyMerchantItem: function(number, ? number): unknown...
global BuyTrainerService: function(number): unknown...
global BuybackItem: function(number): unknown...
global CAMPAIGN_COMPLETE_COLOR: W.Global_CAMPAIGN_COMPLETE_COLOR
global CANDLE_MAJOR_FACTION_COLOR: W.Global_CANDLE_MAJOR_FACTION_COLOR
global CENTAUR_MAJOR_FACTION_COLOR: W.Global_CENTAUR_MAJOR_FACTION_COLOR
global CHALLENGE_MODE_TOAST_TITLE_COLOR: W.Global_CHALLENGE_MODE_TOAST_TITLE_COLOR
global COMBAT_TEXT_PET_COLOR: W.Global_COMBAT_TEXT_PET_COLOR
global COMBAT_TEXT_SPELL_COLOR: W.Global_COMBAT_TEXT_SPELL_COLOR
global COMMON_GRAY_COLOR: W.Global_COMMON_GRAY_COLOR
global CONTEXT_FEEDBACK_COLOR: W.Global_CONTEXT_FEEDBACK_COLOR
global CORRUPTION_COLOR: W.Global_CORRUPTION_COLOR
global C_AccessibilityOptions: W.Global_C_AccessibilityOptions
global C_AccountInfo: W.Global_C_AccountInfo
global C_AccountServices: W.Global_C_AccountServices
global C_AccountStore: W.Global_C_AccountStore
global C_AchievementInfo: W.Global_C_AchievementInfo
global C_AchievementTelemetry: W.Global_C_AchievementTelemetry
global C_ActionBar: W.Global_C_ActionBar
global C_AddOnProfiler: W.Global_C_AddOnProfiler
global C_AddOns: W.Global_C_AddOns
global C_AdventureJournal: W.Global_C_AdventureJournal
global C_AdventureMap: W.Global_C_AdventureMap
global C_AlliedRaces: W.Global_C_AlliedRaces
global C_AnimaDiversion: W.Global_C_AnimaDiversion
global C_ArdenwealdGardening: W.Global_C_ArdenwealdGardening
global C_AreaPoiInfo: W.Global_C_AreaPoiInfo
global C_ArrowCalloutManager: W.Global_C_ArrowCalloutManager
global C_ArtifactUI: W.Global_C_ArtifactUI
global C_AssistedCombat: W.Global_C_AssistedCombat
global C_AuctionHouse: W.Global_C_AuctionHouse
global C_AuraContainerUtil: W.Global_C_AuraContainerUtil
global C_AutoComplete: W.Global_C_AutoComplete
global C_AzeriteEmpoweredItem: W.Global_C_AzeriteEmpoweredItem
global C_AzeriteEssence: W.Global_C_AzeriteEssence
global C_AzeriteItem: W.Global_C_AzeriteItem
global C_Bank: W.Global_C_Bank
global C_BarberShop: W.Global_C_BarberShop
global C_BarberShopInternal: W.Global_C_BarberShopInternal
global C_BattleNet: W.Global_C_BattleNet
global C_BattlePet: W.Global_C_BattlePet
global C_BehavioralMessaging: W.Global_C_BehavioralMessaging
global C_BlackMarket: W.Global_C_BlackMarket
global C_BlackMarketInfo: W.Global_C_BlackMarketInfo
global C_Browser: W.Global_C_Browser
global C_CVar: W.Global_C_CVar
global C_Calendar: W.Global_C_Calendar
global C_CampaignInfo: W.Global_C_CampaignInfo
global C_CatalogShop: W.Global_C_CatalogShop
global C_ChallengeMode: W.Global_C_ChallengeMode
global C_CharacterServices: W.Global_C_CharacterServices
global C_CharacterServicesPublic: W.Global_C_CharacterServicesPublic
global C_ChatBubbles: W.Global_C_ChatBubbles
global C_ChatInfo: W.Global_C_ChatInfo
global C_ChromieTime: W.Global_C_ChromieTime
global C_CinematicList: W.Global_C_CinematicList
global C_ClassColor: W.Global_C_ClassColor
global C_ClassTalents: W.Global_C_ClassTalents
global C_ClassTrial: W.Global_C_ClassTrial
global C_ClickBindings: W.Global_C_ClickBindings
global C_ClientScene: W.Global_C_ClientScene
global C_Club: W.Global_C_Club
global C_ClubFinder: W.Global_C_ClubFinder
global C_ColorOverrides: W.Global_C_ColorOverrides
global C_ColorUtil: W.Global_C_ColorUtil
global C_CombatAudioAlert: W.Global_C_CombatAudioAlert
global C_CombatLog: W.Global_C_CombatLog
global C_CombatLogInternal: W.Global_C_CombatLogInternal
global C_CombatText: W.Global_C_CombatText
global C_Commentator: W.Global_C_Commentator
global C_CompactUnitFrames: W.Global_C_CompactUnitFrames
global C_ConfigurationWarnings: W.Global_C_ConfigurationWarnings
global C_ConsoleScriptCollection: W.Global_C_ConsoleScriptCollection
global C_Container: W.Global_C_Container
global C_ContentTracking: W.Global_C_ContentTracking
global C_ContributionCollector: W.Global_C_ContributionCollector
global C_CooldownViewer: W.Global_C_CooldownViewer
global C_CovenantCallings: W.Global_C_CovenantCallings
global C_CovenantPreview: W.Global_C_CovenantPreview
global C_CovenantSanctumUI: W.Global_C_CovenantSanctumUI
global C_Covenants: W.Global_C_Covenants
global C_CraftingOrders: W.Global_C_CraftingOrders
global C_CreatureInfo: W.Global_C_CreatureInfo
global C_CurrencyInfo: W.Global_C_CurrencyInfo
global C_Cursor: W.Global_C_Cursor
global C_CursorUtil: W.Global_C_CursorUtil
global C_CurveUtil: W.Global_C_CurveUtil
global C_DamageMeter: W.Global_C_DamageMeter
global C_DateAndTime: W.Global_C_DateAndTime
global C_DeathAlert: W.Global_C_DeathAlert
global C_DeathInfo: W.Global_C_DeathInfo
global C_DeathRecap: W.Global_C_DeathRecap
global C_Debug: W.Global_C_Debug
global C_DelvesUI: W.Global_C_DelvesUI
global C_Discord: W.Global_C_Discord
global C_DuelInfo: W.Global_C_DuelInfo
global C_DurationUtil: W.Global_C_DurationUtil
global C_DyeColor: W.Global_C_DyeColor
global C_EditMode: W.Global_C_EditMode
global C_EncodingUtil: W.Global_C_EncodingUtil
global C_EncounterEvents: W.Global_C_EncounterEvents
global C_EncounterInfo: W.Global_C_EncounterInfo
global C_EncounterJournal: W.Global_C_EncounterJournal
global C_EncounterTimeline: W.Global_C_EncounterTimeline
global C_EncounterWarnings: W.Global_C_EncounterWarnings
global C_EndOfMatchUI: W.Global_C_EndOfMatchUI
global C_EquipmentSet: W.Global_C_EquipmentSet
global C_EventScheduler: W.Global_C_EventScheduler
global C_EventToastManager: W.Global_C_EventToastManager
global C_EventUtils: W.Global_C_EventUtils
global C_ExpansionTrial: W.Global_C_ExpansionTrial
global C_ExternalEventURL: W.Global_C_ExternalEventURL
global C_FogOfWar: W.Global_C_FogOfWar
global C_FrameManager: W.Global_C_FrameManager
global C_FriendList: W.Global_C_FriendList
global C_FunctionContainers: W.Global_C_FunctionContainers
global C_GMTicketInfo: W.Global_C_GMTicketInfo
global C_GamePad: W.Global_C_GamePad
global C_GameRules: W.Global_C_GameRules
global C_Garrison: W.Global_C_Garrison
global C_GenericWidgetDisplay: W.Global_C_GenericWidgetDisplay
global C_Glue: W.Global_C_Glue
global C_GlyphInfo: W.Global_C_GlyphInfo
global C_GossipInfo: W.Global_C_GossipInfo
global C_GuildBank: W.Global_C_GuildBank
global C_GuildInfo: W.Global_C_GuildInfo
global C_Heirloom: W.Global_C_Heirloom
global C_HeirloomInfo: W.Global_C_HeirloomInfo
global C_HouseEditor: W.Global_C_HouseEditor
global C_HouseExterior: W.Global_C_HouseExterior
global C_Housing: W.Global_C_Housing
global C_HousingBasicMode: W.Global_C_HousingBasicMode
global C_HousingBlueprint: W.Global_C_HousingBlueprint
global C_HousingCatalog: W.Global_C_HousingCatalog
global C_HousingCleanupMode: W.Global_C_HousingCleanupMode
global C_HousingCustomizeMode: W.Global_C_HousingCustomizeMode
global C_HousingDecor: W.Global_C_HousingDecor
global C_HousingExpertMode: W.Global_C_HousingExpertMode
global C_HousingInspectMode: W.Global_C_HousingInspectMode
global C_HousingLayout: W.Global_C_HousingLayout
global C_HousingNeighborhood: W.Global_C_HousingNeighborhood
global C_HousingPhotoSharing: W.Global_C_HousingPhotoSharing
global C_ImmersiveInteraction: W.Global_C_ImmersiveInteraction
global C_IncomingSummon: W.Global_C_IncomingSummon
global C_InstanceEncounter: W.Global_C_InstanceEncounter
global C_InstanceLeaver: W.Global_C_InstanceLeaver
global C_InterfaceFileManifest: W.Global_C_InterfaceFileManifest
global C_Intl: W.Global_C_Intl
global C_InvasionInfo: W.Global_C_InvasionInfo
global C_IslandsInfo: W.Global_C_IslandsInfo
global C_IslandsQueue: W.Global_C_IslandsQueue
global C_Item: W.Global_C_Item
global C_ItemInteraction: W.Global_C_ItemInteraction
global C_ItemSocketInfo: W.Global_C_ItemSocketInfo
global C_ItemText: W.Global_C_ItemText
global C_ItemUpgrade: W.Global_C_ItemUpgrade
global C_KeyBindings: W.Global_C_KeyBindings
global C_LFGInfo: W.Global_C_LFGInfo
global C_LFGList: W.Global_C_LFGList
global C_LegendaryCrafting: W.Global_C_LegendaryCrafting
global C_LevelLink: W.Global_C_LevelLink
global C_LevelSquish: W.Global_C_LevelSquish
global C_LimitedInput: W.Global_C_LimitedInput
global C_LiveEvent: W.Global_C_LiveEvent
global C_LoadingScreen: W.Global_C_LoadingScreen
global C_LobbyMatchmakerInfo: W.Global_C_LobbyMatchmakerInfo
global C_Log: W.Global_C_Log
global C_Loot: W.Global_C_Loot
global C_LootHistory: W.Global_C_LootHistory
global C_LootJournal: W.Global_C_LootJournal
global C_LoreText: W.Global_C_LoreText
global C_LossOfControl: W.Global_C_LossOfControl
global C_MacOptions: W.Global_C_MacOptions
global C_Macro: W.Global_C_Macro
global C_Mail: W.Global_C_Mail
global C_MajorFactions: W.Global_C_MajorFactions
global C_Map: W.Global_C_Map
global C_MapExplorationInfo: W.Global_C_MapExplorationInfo
global C_MerchantFrame: W.Global_C_MerchantFrame
global C_Minimap: W.Global_C_Minimap
global C_ModelInfo: W.Global_C_ModelInfo
global C_ModifiedInstance: W.Global_C_ModifiedInstance
global C_MountJournal: W.Global_C_MountJournal
global C_MythicPlus: W.Global_C_MythicPlus
global C_NamePlate: W.Global_C_NamePlate
global C_NamePlateManager: W.Global_C_NamePlateManager
global C_Navigation: W.Global_C_Navigation
global C_NeighborhoodInitiative: W.Global_C_NeighborhoodInitiative
global C_NewItems: W.Global_C_NewItems
global C_PaperDollInfo: W.Global_C_PaperDollInfo
global C_PartyInfo: W.Global_C_PartyInfo
global C_PartyPose: W.Global_C_PartyPose
global C_PerksActivities: W.Global_C_PerksActivities
global C_PerksProgram: W.Global_C_PerksProgram
global C_PetBattles: W.Global_C_PetBattles
global C_PetInfo: W.Global_C_PetInfo
global C_PetJournal: W.Global_C_PetJournal
global C_PhotoSharing: W.Global_C_PhotoSharing
global C_Ping: W.Global_C_Ping
global C_Platform: W.Global_C_Platform
global C_PlayerChoice: W.Global_C_PlayerChoice
global C_PlayerInfo: W.Global_C_PlayerInfo
global C_PlayerInteractionManager: W.Global_C_PlayerInteractionManager
global C_PlayerMentorship: W.Global_C_PlayerMentorship
global C_Pony: W.Global_C_Pony
global C_ProfSpecs: W.Global_C_ProfSpecs
global C_PrototypeDialog: W.Global_C_PrototypeDialog
global C_PvP: W.Global_C_PvP
global C_QuestHub: W.Global_C_QuestHub
global C_QuestInfoSystem: W.Global_C_QuestInfoSystem
global C_QuestItemUse: W.Global_C_QuestItemUse
global C_QuestLine: W.Global_C_QuestLine
global C_QuestLog: W.Global_C_QuestLog
global C_QuestOffer: W.Global_C_QuestOffer
global C_QuestSession: W.Global_C_QuestSession
global C_RaidLocks: W.Global_C_RaidLocks
global C_RecentAllies: W.Global_C_RecentAllies
global C_RecruitAFriend: W.Global_C_RecruitAFriend
global C_Reincarnation: W.Global_C_Reincarnation
global C_RemixArtifactUI: W.Global_C_RemixArtifactUI
global C_ReportSystem: W.Global_C_ReportSystem
global C_Reputation: W.Global_C_Reputation
global C_ResearchInfo: W.Global_C_ResearchInfo
global C_RestrictedActions: W.Global_C_RestrictedActions
global C_ReturningPlayerUI: W.Global_C_ReturningPlayerUI
global C_Roleset: W.Global_C_Roleset
global C_Scenario: W.Global_C_Scenario
global C_ScenarioInfo: W.Global_C_ScenarioInfo
global C_ScrappingMachineUI: W.Global_C_ScrappingMachineUI
global C_ScriptWarnings: W.Global_C_ScriptWarnings
global C_ScriptedAnimations: W.Global_C_ScriptedAnimations
global C_SeasonInfo: W.Global_C_SeasonInfo
global C_Secrets: W.Global_C_Secrets
global C_SecureTransfer: W.Global_C_SecureTransfer
global C_SettingsUtil: W.Global_C_SettingsUtil
global C_SharedCharacterServices: W.Global_C_SharedCharacterServices
global C_SkillInfo: W.Global_C_SkillInfo
global C_SocialQueue: W.Global_C_SocialQueue
global C_SocialRestrictions: W.Global_C_SocialRestrictions
global C_SocialUI: W.Global_C_SocialUI
global C_Soulbinds: W.Global_C_Soulbinds
global C_Sound: W.Global_C_Sound
global C_SpecializationInfo: W.Global_C_SpecializationInfo
global C_SpectatingUI: W.Global_C_SpectatingUI
global C_Spell: W.Global_C_Spell
global C_SpellActivationOverlay: W.Global_C_SpellActivationOverlay
global C_SpellBook: W.Global_C_SpellBook
global C_SpellDiminish: W.Global_C_SpellDiminish
global C_SplashScreen: W.Global_C_SplashScreen
global C_StableInfo: W.Global_C_StableInfo
global C_StorePublic: W.Global_C_StorePublic
global C_StringUtil: W.Global_C_StringUtil
global C_SummonInfo: W.Global_C_SummonInfo
global C_SuperTrack: W.Global_C_SuperTrack
global C_System: W.Global_C_System
global C_SystemVisibilityManager: W.Global_C_SystemVisibilityManager
global C_TTSSettings: W.Global_C_TTSSettings
global C_TableUtil: W.Global_C_TableUtil
global C_TalkingHead: W.Global_C_TalkingHead
global C_TaskQuest: W.Global_C_TaskQuest
global C_TaxiMap: W.Global_C_TaxiMap
global C_Texture: W.Global_C_Texture
global C_Timer: W.Global_C_Timer
global C_TimerunningUI: W.Global_C_TimerunningUI
global C_TooltipComparison: W.Global_C_TooltipComparison
global C_TooltipInfo: W.Global_C_TooltipInfo
global C_ToyBox: W.Global_C_ToyBox
global C_ToyBoxInfo: W.Global_C_ToyBoxInfo
global C_TradeInfo: W.Global_C_TradeInfo
global C_TradeSkillUI: W.Global_C_TradeSkillUI
global C_Trainer: W.Global_C_Trainer
global C_TraitConfig: W.Global_C_TraitConfig
global C_Traits: W.Global_C_Traits
global C_Transmog: W.Global_C_Transmog
global C_TransmogCollection: W.Global_C_TransmogCollection
global C_TransmogOutfitInfo: W.Global_C_TransmogOutfitInfo
global C_TransmogSets: W.Global_C_TransmogSets
global C_Trophy: W.Global_C_Trophy
global C_Tutorial: W.Global_C_Tutorial
global C_UI: W.Global_C_UI
global C_UIActionHandler: W.Global_C_UIActionHandler
global C_UIColor: W.Global_C_UIColor
global C_UIFileAsset: W.Global_C_UIFileAsset
global C_UIWidgetManager: W.Global_C_UIWidgetManager
global C_UnitAuras: W.Global_C_UnitAuras
global C_UserFeedback: W.Global_C_UserFeedback
global C_Vehicle: W.Global_C_Vehicle
global C_VideoOptions: W.Global_C_VideoOptions
global C_VignetteInfo: W.Global_C_VignetteInfo
global C_VoiceChat: W.Global_C_VoiceChat
global C_WarbandScene: W.Global_C_WarbandScene
global C_Weather: W.Global_C_Weather
global C_WeeklyRewards: W.Global_C_WeeklyRewards
global C_Widget: W.Global_C_Widget
global C_WoWLabsMatchmaking: W.Global_C_WoWLabsMatchmaking
global C_WorldLootObject: W.Global_C_WorldLootObject
global C_WorldSafeLocsUIInternal: W.Global_C_WorldSafeLocsUIInternal
global C_WorldStateInfo: W.Global_C_WorldStateInfo
global C_WowEntitlementInfo: W.Global_C_WowEntitlementInfo
global C_WowLabsDataManager: W.Global_C_WowLabsDataManager
global C_WowSurvey: W.Global_C_WowSurvey
global C_WowTokenPublic: W.Global_C_WowTokenPublic
global C_WowTokenUI: W.Global_C_WowTokenUI
global C_XMLUtil: W.Global_C_XMLUtil
global C_ZoneAbility: W.Global_C_ZoneAbility
global CalculateStringEditDistance: function(string, string): number
global CallCompanion: function(string, number): unknown...
global CallbackRegistryMixin: W.Global_CallbackRegistryMixin
global CameraOrSelectOrMoveStart: function(): unknown...
global CameraOrSelectOrMoveStop: function(? boolean): unknown...
global CameraZoomIn: function(number): unknown...
global CameraZoomOut: function(number): unknown...
global CanAffordMerchantItem: function(unknown): unknown...
global CanAutoSetGamePadCursorControl: function(): unknown...
global CanBeRaidTarget: function(string): boolean
global CanCancelScene: function(): boolean
global CanChangePlayerDifficulty: function(): boolean | nil, boolean | nil
global CanComplainInboxItem: function(number): boolean | nil
global CanDualWield: function(): boolean
global CanEditGuildBankTabInfo: function(): unknown...
global CanEditGuildEvent: function(): unknown...
global CanEditGuildInfo: function(): unknown...
global CanEditGuildTabInfo: function(unknown): unknown...
global CanEditMOTD: function(): boolean
global CanEditPublicNote: function(): unknown...
global CanEjectPassengerFromSeat: function(number): boolean
global CanExitVehicle: function(): unknown...
global CanGamePadControlCursor: function(): unknown...
global CanGuildBankRepair: function(): unknown...
global CanGuildDemote: function(): boolean
global CanGuildInvite: function(): boolean
global CanGuildPromote: function(): boolean
global CanGuildRemove: function(): unknown...
global CanHearthAndResurrectFromArea: function(): unknown...
global CanInitiateWarGame: function(): unknown...
global CanInspect: function(string): boolean
global CanItemBeSocketedToArtifact: function(number): boolean
global CanJoinBattlefieldAsGroup: function(): boolean
global CanLootUnit: function(string): boolean | nil, boolean | nil
global CanMapChangeDifficulty: function(? number): boolean
global CanMerchantRepair: function(): boolean
global CanPartyLFGBackfill: function(): unknown...
global CanReplaceGuildMaster: function(): boolean
global CanResetTutorials: function(): unknown...
global CanScanResearchSite: function(): boolean
global CanShowAchievementUI: function(): boolean
global CanShowResetInstances: function(): boolean
global CanShowSetRoleButton: function(): boolean
global CanSignPetition: function(): unknown...
global CanSolveArtifact: function(): boolean
global CanSurrenderArena: function(): unknown...
global CanSwitchVehicleSeat: function(): boolean
global CanSwitchVehicleSeats: function(): boolean
global CanUpgradeToCurrentExpansion: function(): boolean
global CanViewGuildRecipes: function(number): boolean
global CanWithdrawGuildBankMoney: function(): unknown...
global CancelAreaSpiritHeal: function()
global CancelDuel: function(): unknown...
global CancelItemTempEnchantment: function(number): unknown...
global CancelLogout: function()
global CancelMasterLootRoll: function(unknown): unknown...
global CancelPendingEquip: function(number)
global CancelPetPossess: function(): unknown...
global CancelPreloadingMovie: function(number)
global CancelScene: function(): unknown...
global CancelShapeshiftForm: function(): unknown...
global CancelSpellByName: function(unknown): unknown...
global CancelTrade: function()
global CancelTradeAccept: function(): unknown...
global CancelUnitBuff: function(string, number, ? string): unknown...
global CannotBeResurrected: function(): unknown...
global CaseAccentInsensitiveParse: function(string): string
global CastPetAction: function(number, ? string): unknown...
global CastShapeshiftForm: function(number): unknown...
global CastSpell: function(number, string): unknown...
global CastSpellByID: function(unknown, unknown): unknown...
global CastSpellByName: function(string, ? string): unknown...
global CenterCamera: function(): unknown...
global ChangeChatColor: function(string, number, number, number): unknown...
global ChannelBan: function(string, string): unknown...
global ChannelInvite: function(string, string): unknown...
global ChannelKick: function(string, string): unknown...
global ChannelModerator: function(string, string): unknown...
global ChannelSetAllSilent: function(): unknown...
global ChannelSetPartyMemberSilent: function(unknown, unknown): unknown...
global ChannelToggleAnnouncements: function(string, string): unknown...
global ChannelUnban: function(string, string): unknown...
global ChannelUnmoderator: function(string, string): unknown...
global ChatBubbleFont: W.Global_ChatBubbleFont
global ChatFontNormal: W.Global_ChatFontNormal
global ChatFontSmall: W.Global_ChatFontSmall
global ChatTypeInfo: W.Global_ChatTypeInfo
global CheckInbox: function(): unknown...
global CheckInteractDistance: function(string, number): boolean
global CheckTalentMasterDist: function(): boolean
global CinematicFinished: function(number, ? boolean, ? boolean)
global CinematicStarted: function(number, number, ? boolean)
global ClassicExpansionAtLeast: function(number): boolean
global ClassicExpansionAtMost: function(number): boolean
global ClearAchievementComparisonUnit: function(): unknown...
global ClearAchievementSearchString: function(): unknown...
global ClearAllLFGDungeons: function(unknown): unknown...
global ClearAutoAcceptQuestSound: function(): unknown...
global ClearBattlemaster: function(): unknown...
global ClearClampedTextureRotation: function(W.Texture)
global ClearCursor: function()
global ClearCursorHoveredItem: function()
global ClearFailedPVPTalentIDs: function(): unknown...
global ClearFailedTalentIDs: function(): unknown...
global ClearFocus: function()
global ClearInspectPlayer: function(): unknown...
global ClearOutage: function()
global ClearOverrideBindings: function(W.Frame): unknown...
global ClearPartyAssignment: function(): unknown...
global ClearPendingBindConversionItem: function()
global ClearRaidMarker: function(? number)
global ClearSendMail: function(): unknown...
global ClearTarget: function(): boolean
global ClearTutorials: function(): unknown...
global ClickSendMailItemButton: function(? number, ? boolean): unknown...
global ClickTargetTradeButton: function(unknown): unknown...
global ClickTradeButton: function(unknown): unknown...
global ClickWorldMapActionButton: function(): unknown...
global CloseGuildBankFrame: function(): unknown...
global CloseGuildRegistrar: function(): unknown...
global CloseGuildRoster: function(): unknown...
global CloseItemText: function(): unknown...
global CloseLoot: function(? boolean): unknown...
global CloseMail: function(): unknown...
global CloseMerchant: function(): unknown...
global ClosePetition: function(): unknown...
global CloseQuest: function(): unknown...
global CloseResearch: function(): unknown...
global CloseTabardCreation: function(): unknown...
global CloseTaxiMap: function(): unknown...
global CloseTrade: function(): unknown...
global CloseTrainer: function(): unknown...
global ClosestGameObjectPosition: function(number): W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil
global ClosestUnitPosition: function(number): W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil
global CollapseGuildTradeSkillHeader: function(unknown): unknown...
global CollapseQuestHeader: function(number, boolean): unknown...
global CollapseWarGameHeader: function(unknown): unknown...
global ColorMixin: W.Global_ColorMixin
global CombatLogFont: W.Global_CombatLogFont
global CombatTextFont: W.Global_CombatTextFont
global CombatTextFontOutline: W.Global_CombatTextFontOutline
global CommentatorCCFont: W.Global_CommentatorCCFont
global CommentatorDampeningFont: W.Global_CommentatorDampeningFont
global CommentatorFontMedium: W.Global_CommentatorFontMedium
global CommentatorFontSmall: W.Global_CommentatorFontSmall
global CommentatorTeamNameFont: W.Global_CommentatorTeamNameFont
global CommentatorTeamScoreFont: W.Global_CommentatorTeamScoreFont
global CommentatorVictoryFanfare: W.Global_CommentatorVictoryFanfare
global CommentatorVictoryFanfareTeam: W.Global_CommentatorVictoryFanfareTeam
global CompleteLFGReadyCheck: function(unknown): unknown...
global CompleteLFGRoleCheck: function(unknown): unknown...
global CompleteQuest: function(): unknown...
global ConfirmAcceptQuest: function(): unknown...
global ConfirmBNRequestInviteFriend: function(number, ? boolean, ? boolean, ? boolean): unknown...
global ConfirmBindOnUse: function()
global ConfirmLootRoll: function(number, number): unknown...
global ConfirmLootSlot: function(number): unknown...
global ConfirmNoRefundOnUse: function()
global ConfirmOnUse: function()
global ConfirmReadyCheck: function(? number): unknown...
global ConfirmTalentWipe: function()
global ConsoleEcho: function(string, ? boolean, ? string): boolean
global ConsoleExec: function(string, ? boolean): boolean
global ConsoleFontNormal: W.Global_ConsoleFontNormal
global ConsoleFontSmall: W.Global_ConsoleFontSmall
global ConsoleGetAllCommands: function(): {W.ConsoleCommandInfo}
global ConsoleGetColorFromType: function(number): W.colorRGB
global ConsoleGetFontHeight: function(): number
global ConsoleIsActive: function(): boolean
global ConsolePrint: function(string...)
global ConsolePrintAllMatchingCommands: function(string)
global ConsoleSetFontHeight: function(number)
global Constants: W.Global_Constants
global ContainsIf: function({unknown: unknown}, unknown): boolean
global ContinuableContainer: W.Global_ContinuableContainer
global ConvertItemToBindToAccount: function()
global CopyTable: function({unknown: unknown}, ? boolean): {unknown: unknown}
global CopyToClipboard: function(string, ? boolean): number | nil
global CoreAbilityFont: W.Global_CoreAbilityFont
global CreateAbbreviateConfig: function({W.NumberAbbreviationBreakpoint}): W.AbbreviateConfig
global CreateAndInitFromMixin: function<T>(T, unknown...): T
global CreateAtlasMarkup: function(number | string, ? number, ? number, ? number, ? number, ? number, ? number, ? number): string
global CreateAtlasMarkupWithAtlasSize: function(number | string, ? number, ? number, ? number, ? number, ? number, unknown): string
global CreateColor: function(number, number, number, ? number): W.colorRGBA
global CreateColorFromBytes: function(number, number, number, number): W.colorRGBA
global CreateColorFromHexString: function(string): W.colorRGBA
global CreateFont: function(string): W.Font
global CreateFontFamily: function(string, {W.CreateFontFamilyMemberInfo}): W.Font
global CreateFrame: W.Global_CreateFrame
global CreateFrameWithOptions: function(W.CreateFrameOptions): W.Frame
global CreateFromMixins: function(unknown...): unknown
global CreateMacro: function(string, number | string, ? string, ? boolean): number
global CreateNewRaidProfile: function(unknown, unknown): unknown...
global CreateSecureDelegate: function(unknown, ? W.SecureDelegateOptions): unknown | nil
global CreateTextureMarkup: function(number | string, number, number, number, number, number, number, number, number, ? number, ? number): string
global CreateUnitHealPredictionCalculator: function(): W.UnitHealPredictionCalculator
global CreateVector2D: function(number, number): W.Vector2DMixin
global CreateVector3D: function(number, number, number): W.Vector3DMixin
global CreateWindow: function(? boolean, ? boolean): {unknown: unknown} | nil
global CursorHasItem: function(): boolean
global CursorHasMacro: function(): boolean
global CursorHasMoney: function(): boolean
global CursorHasSpell: function(): boolean
global DARKGRAY_COLOR: W.Global_DARKGRAY_COLOR
global DARKYELLOW_FONT_COLOR: W.Global_DARKYELLOW_FONT_COLOR
global DECREASE_STAT_COLOR: W.Global_DECREASE_STAT_COLOR
global DEFAULT_CHAT_CHANNEL_COLOR: W.Global_DEFAULT_CHAT_CHANNEL_COLOR
global DEFAULT_CHAT_FRAME: W.Global_DEFAULT_CHAT_FRAME
global DEFAULT_MATERIAL_TEXT_COLOR: W.Global_DEFAULT_MATERIAL_TEXT_COLOR
global DEFAULT_MATERIAL_TITLETEXT_COLOR: W.Global_DEFAULT_MATERIAL_TITLETEXT_COLOR
global DELVES_SCENARIO_NAME_COLOR: W.Global_DELVES_SCENARIO_NAME_COLOR
global DIFFICULT_DIFFICULTY_COLOR: W.Global_DIFFICULT_DIFFICULTY_COLOR
global DIM_GREEN_FONT_COLOR: W.Global_DIM_GREEN_FONT_COLOR
global DIM_RED_FONT_COLOR: W.Global_DIM_RED_FONT_COLOR
global DISABLED_FONT_COLOR: W.Global_DISABLED_FONT_COLOR
global DISABLED_REAGENT_COLOR: W.Global_DISABLED_REAGENT_COLOR
global DRAGONFLIGHT_BLACK_COLOR: W.Global_DRAGONFLIGHT_BLACK_COLOR
global DRAGONFLIGHT_BLUE_COLOR: W.Global_DRAGONFLIGHT_BLUE_COLOR
global DRAGONFLIGHT_BRONZE_COLOR: W.Global_DRAGONFLIGHT_BRONZE_COLOR
global DRAGONFLIGHT_GREEN_COLOR: W.Global_DRAGONFLIGHT_GREEN_COLOR
global DRAGONFLIGHT_RED_COLOR: W.Global_DRAGONFLIGHT_RED_COLOR
global DREAM_MAJOR_FACTION_COLOR: W.Global_DREAM_MAJOR_FACTION_COLOR
global DULL_RED_FONT_COLOR: W.Global_DULL_RED_FONT_COLOR
global DeclineArenaTeam: function()
global DeclineChannelInvite: function(string): unknown...
global DeclineGroup: function(): unknown...
global DeclineGuild: function()
global DeclineName: function(string, ? number, number): string...
global DeclineQuest: function(): unknown...
global DeclineResurrect: function()
global DeclineSpellConfirmationPrompt: function(number): unknown...
global DeleteCursorItem: function()
global DeleteGMTicket: function(): unknown...
global DeleteInboxItem: function(number): unknown...
global DeleteMacro: function(number | string): unknown...
global DeleteRaidProfile: function(unknown): unknown...
global DemoteAssistant: function(unknown): unknown...
global DepositGuildBankMoney: function(unknown): unknown...
global DescendStop: function(): unknown...
global DestinyFontHuge: W.Global_DestinyFontHuge
global DestinyFontLarge: W.Global_DestinyFontLarge
global DestinyFontMed: W.Global_DestinyFontMed
global DestroyTotem: function(number)
global DetectWowMouse: function(): unknown...
global DevTools_Dump: function(unknown, ? string)
global DialogButtonHighlightText: W.Global_DialogButtonHighlightText
global DialogButtonNormalText: W.Global_DialogButtonNormalText
global DismissCompanion: function(string): unknown...
global Dismount: function()
global DisplayChannelOwner: function(string): unknown...
global DoMasterLootRoll: function(unknown): unknown...
global DoReadyCheck: function(): unknown...
global DoesCurrentLocaleSellExpansionLevels: function(): boolean
global DoesItemContainSpec: function(number | string, number, ? number): boolean
global DoesTemplateExist: function(unknown): unknown...
global DropCursorMoney: function()
global DropItemOnUnit: function(string)
global DropdownButtonMixin: W.Global_DropdownButtonMixin
global DungeonAppearsInRandomLFD: function(unknown): unknown...
global EASY_DIFFICULTY_COLOR: W.Global_EASY_DIFFICULTY_COLOR
global EDIT_MODE_GRID_CENTER_LINE_COLOR: W.Global_EDIT_MODE_GRID_CENTER_LINE_COLOR
global EDIT_MODE_GRID_LINE_COLOR: W.Global_EDIT_MODE_GRID_LINE_COLOR
global EDIT_MODE_LAYOUT_LINK_COLOR: W.Global_EDIT_MODE_LAYOUT_LINK_COLOR
global EJ_ClearSearch: function(): unknown...
global EJ_EndSearch: function(): unknown...
global EJ_GetContentTuningID: function(): number
global EJ_GetCreatureInfo: function(number, ? number): number, string, string, number, number, number
global EJ_GetCurrentTier: function(): number
global EJ_GetDifficulty: function(): number
global EJ_GetEncounterInfo: function(number): string, string, number, number, string, number, number, number
global EJ_GetEncounterInfoByIndex: function(number, ? number): string, string, number, number, string, number, number, number
global EJ_GetInstanceByIndex: function(number, boolean): number, string, string, number, number, number, number, number, string, boolean, number, number, boolean
global EJ_GetInstanceForMap: function(number): number
global EJ_GetInstanceInfo: function(? number): string, string, number, number, number, number, number, string, boolean, number, number, boolean
global EJ_GetInvTypeSortOrder: function(number): number
global EJ_GetLootFilter: function(): number, number
global EJ_GetMapEncounter: function(number, number, ? boolean): number, number, number, string, string, number, number, string
global EJ_GetNumEncountersForLootByIndex: function(number): number
global EJ_GetNumLoot: function(): number
global EJ_GetNumSearchResults: function(): number
global EJ_GetNumTiers: function(): number
global EJ_GetSearchProgress: function(): number
global EJ_GetSearchResult: function(number): number, number, number, number, number, string
global EJ_GetSearchSize: function(): number
global EJ_GetSectionPath: function(number): number, number | nil, number | nil, number | nil
global EJ_GetTierInfo: function(number): string, string
global EJ_HandleLinkPath: function(number, number): number, number | nil, number | nil, number | nil
global EJ_InstanceIsRaid: function(): boolean
global EJ_IsLootListOutOfDate: function(): boolean
global EJ_IsSearchFinished: function(): boolean
global EJ_IsValidInstanceDifficulty: function(number): boolean
global EJ_ResetLootFilter: function(): unknown...
global EJ_SPELL_COLOR: W.Global_EJ_SPELL_COLOR
global EJ_SelectEncounter: function(number): unknown...
global EJ_SelectInstance: function(number): unknown...
global EJ_SelectTier: function(number): unknown...
global EJ_SetDifficulty: function(number): unknown...
global EJ_SetLootFilter: function(number, number): unknown...
global EJ_SetSearch: function(string): unknown...
global ENCHANT_COLOR: W.Global_ENCHANT_COLOR
global ENCOUNTER_JOURNAL_SCROLL_BAR_BACKGROUND_COLOR: W.Global_ENCOUNTER_JOURNAL_SCROLL_BAR_BACKGROUND_COLOR
global ENHANCED_CONDUIT_COLOR: W.Global_ENHANCED_CONDUIT_COLOR
global EPIC_PURPLE_COLOR: W.Global_EPIC_PURPLE_COLOR
global ERROR_COLOR: W.Global_ERROR_COLOR
global EXPANSION_COLOR_1: W.Global_EXPANSION_COLOR_1
global EXPANSION_COLOR_10: W.Global_EXPANSION_COLOR_10
global EXPANSION_COLOR_11: W.Global_EXPANSION_COLOR_11
global EXPANSION_COLOR_12: W.Global_EXPANSION_COLOR_12
global EXPANSION_COLOR_13: W.Global_EXPANSION_COLOR_13
global EXPANSION_COLOR_2: W.Global_EXPANSION_COLOR_2
global EXPANSION_COLOR_3: W.Global_EXPANSION_COLOR_3
global EXPANSION_COLOR_4: W.Global_EXPANSION_COLOR_4
global EXPANSION_COLOR_5: W.Global_EXPANSION_COLOR_5
global EXPANSION_COLOR_6: W.Global_EXPANSION_COLOR_6
global EXPANSION_COLOR_7: W.Global_EXPANSION_COLOR_7
global EXPANSION_COLOR_8: W.Global_EXPANSION_COLOR_8
global EXPANSION_COLOR_9: W.Global_EXPANSION_COLOR_9
global EXPEDITION_MAJOR_FACTION_COLOR: W.Global_EXPEDITION_MAJOR_FACTION_COLOR
global EditMacro: function(number | string, ? string, ? number | string, ? string): number
global EditModeLayoutDialog: W.Global_EditModeLayoutDialog
global EditModeManagerFrame: W.Global_EditModeManagerFrame
global EditModeUnsavedChangesDialog: W.Global_EditModeUnsavedChangesDialog
global EjectPassengerFromSeat: function(number)
global EndBoundTradeable: function(string)
global EndRefund: function(number)
global Enum: W.Global_Enum
global EnumerateFrames: function(? W.Frame): W.Frame | nil
global EnumerateServerChannels: function(): string...
global EquipCursorItem: function(number)
global EquipItemByName: function(number | string, ? number)
global EquipPendingItem: function(number)
global EquipmentManager_UnpackLocation: function(number)
global ErrorFont: W.Global_ErrorFont
global EventFrameMixin: W.Global_EventFrameMixin
global EventRegistry: W.Global_EventRegistry
global ExpandGuildTradeSkillHeader: function(unknown): unknown...
global ExpandQuestHeader: function(number, boolean): unknown...
global ExpandWarGameHeader: function(unknown): unknown...
global FACTION_AT_WAR_COLOR: W.Global_FACTION_AT_WAR_COLOR
global FACTION_GREEN_COLOR: W.Global_FACTION_GREEN_COLOR
global FACTION_ORANGE_COLOR: W.Global_FACTION_ORANGE_COLOR
global FACTION_RED_COLOR: W.Global_FACTION_RED_COLOR
global FACTION_YELLOW_COLOR: W.Global_FACTION_YELLOW_COLOR
global FAIR_DIFFICULTY_COLOR: W.Global_FAIR_DIFFICULTY_COLOR
global FLAME_MAJOR_FACTION_COLOR: W.Global_FLAME_MAJOR_FACTION_COLOR
global FRAMESTACK_FRAME_COLOR: W.Global_FRAMESTACK_FRAME_COLOR
global FRAMESTACK_HIDDEN_COLOR: W.Global_FRAMESTACK_HIDDEN_COLOR
global FRAMESTACK_REGION_COLOR: W.Global_FRAMESTACK_REGION_COLOR
global FRIENDLY_STATUS_COLOR: W.Global_FRIENDLY_STATUS_COLOR
global FRIENDS_BNET_BACKGROUND_COLOR: W.Global_FRIENDS_BNET_BACKGROUND_COLOR
global FRIENDS_BNET_NAME_COLOR: W.Global_FRIENDS_BNET_NAME_COLOR
global FRIENDS_BROADCAST_TIME_COLOR: W.Global_FRIENDS_BROADCAST_TIME_COLOR
global FRIENDS_GRAY_COLOR: W.Global_FRIENDS_GRAY_COLOR
global FRIENDS_OFFLINE_BACKGROUND_COLOR: W.Global_FRIENDS_OFFLINE_BACKGROUND_COLOR
global FRIENDS_OTHER_NAME_COLOR: W.Global_FRIENDS_OTHER_NAME_COLOR
global FRIENDS_WOW_BACKGROUND_COLOR: W.Global_FRIENDS_WOW_BACKGROUND_COLOR
global FRIENDS_WOW_NAME_COLOR: W.Global_FRIENDS_WOW_NAME_COLOR
global Fancy12Font: W.Global_Fancy12Font
global Fancy14Font: W.Global_Fancy14Font
global Fancy16Font: W.Global_Fancy16Font
global Fancy18Font: W.Global_Fancy18Font
global Fancy20Font: W.Global_Fancy20Font
global Fancy22Font: W.Global_Fancy22Font
global Fancy24Font: W.Global_Fancy24Font
global Fancy27Font: W.Global_Fancy27Font
global Fancy30Font: W.Global_Fancy30Font
global Fancy32Font: W.Global_Fancy32Font
global Fancy48Font: W.Global_Fancy48Font
global FindInTableIf: function({unknown: unknown}, unknown): string | number, unknown
global FindSpellBookSlotBySpellID: function(unknown, unknown): unknown...
global FlagTutorial: function(unknown): unknown...
global FlashClientIcon: function(? boolean)
global FlipCameraYaw: function(number): unknown...
global FlyoutHasSpell: function(number, number): boolean
global FocusFontSmall: W.Global_FocusFontSmall
global FocusUnit: function(? string)
global FollowUnit: function(? string, ? boolean)
global ForceLogout: function()
global ForceQuit: function()
global ForfeitDuel: function(): unknown...
global FormatFraction: function(number, number): string
global FormatLargeNumber: function(string | number): string
global FormatPercentage: function(number, ? boolean): string
global FormatShortDate: function(number, number, number): string
global FormatUnreadMailTooltip: function(W.GameTooltip, string, {string})
global FormatValueWithSign: function(number): string
global FriendsFont_11: W.Global_FriendsFont_11
global FriendsFont_Large: W.Global_FriendsFont_Large
global FriendsFont_Normal: W.Global_FriendsFont_Normal
global FriendsFont_Small: W.Global_FriendsFont_Small
global FriendsFont_UserText: W.Global_FriendsFont_UserText
global GLUE_DIALOG_FONT_COLOR: W.Global_GLUE_DIALOG_FONT_COLOR
global GLYPH_LINK_COLOR: W.Global_GLYPH_LINK_COLOR
global GMEuropaBugsEnabled: function(): unknown...
global GMEuropaComplaintsEnabled: function(): unknown...
global GMEuropaSuggestionsEnabled: function(): unknown...
global GMEuropaTicketsEnabled: function(): unknown...
global GMItemRestorationButtonEnabled: function(): unknown...
global GMQuickTicketSystemEnabled: function(): unknown...
global GMQuickTicketSystemThrottled: function(): unknown...
global GMReportLag: function(): unknown...
global GMRequestPlayerInfo: function(): unknown...
global GMResponseResolve: function(): unknown...
global GMSurveyAnswer: function(): unknown...
global GMSurveyAnswerSubmit: function(unknown, unknown, unknown): unknown...
global GMSurveyCommentSubmit: function(unknown): unknown...
global GMSurveyNumAnswers: function(): unknown...
global GMSurveyQuestion: function(number): unknown...
global GMSurveySubmit: function(): unknown...
global GOLD_FONT_COLOR: W.Global_GOLD_FONT_COLOR
global GRAY_FONT_COLOR: W.Global_GRAY_FONT_COLOR
global GREEN_FONT_COLOR: W.Global_GREEN_FONT_COLOR
global Game10Font_o1: W.Global_Game10Font_o1
global Game11Font: W.Global_Game11Font
global Game11Font_Shadow: W.Global_Game11Font_Shadow
global Game11Font_o1: W.Global_Game11Font_o1
global Game120Font: W.Global_Game120Font
global Game12Font: W.Global_Game12Font
global Game12Font_o1: W.Global_Game12Font_o1
global Game13Font: W.Global_Game13Font
global Game13FontShadow: W.Global_Game13FontShadow
global Game13Font_o1: W.Global_Game13Font_o1
global Game15Font: W.Global_Game15Font
global Game15Font_o1: W.Global_Game15Font_o1
global Game16Font: W.Global_Game16Font
global Game17Font_Shadow: W.Global_Game17Font_Shadow
global Game18Font: W.Global_Game18Font
global Game20Font: W.Global_Game20Font
global Game24Font: W.Global_Game24Font
global Game27Font: W.Global_Game27Font
global Game30Font: W.Global_Game30Font
global Game32Font: W.Global_Game32Font
global Game32Font_Shadow2: W.Global_Game32Font_Shadow2
global Game36Font: W.Global_Game36Font
global Game36Font_Shadow2: W.Global_Game36Font_Shadow2
global Game40Font: W.Global_Game40Font
global Game40Font_Shadow2: W.Global_Game40Font_Shadow2
global Game42Font: W.Global_Game42Font
global Game46Font: W.Global_Game46Font
global Game46Font_Shadow2: W.Global_Game46Font_Shadow2
global Game48Font: W.Global_Game48Font
global Game48FontShadow: W.Global_Game48FontShadow
global Game52Font_Shadow2: W.Global_Game52Font_Shadow2
global Game58Font_Shadow2: W.Global_Game58Font_Shadow2
global Game60Font: W.Global_Game60Font
global Game69Font_Shadow2: W.Global_Game69Font_Shadow2
global Game72Font: W.Global_Game72Font
global Game72Font_Shadow: W.Global_Game72Font_Shadow
global GameFont72Highlight: W.Global_GameFont72Highlight
global GameFont72HighlightShadow: W.Global_GameFont72HighlightShadow
global GameFont72Normal: W.Global_GameFont72Normal
global GameFont72NormalShadow: W.Global_GameFont72NormalShadow
global GameFontBlack: W.Global_GameFontBlack
global GameFontBlackMedium: W.Global_GameFontBlackMedium
global GameFontBlackSmall: W.Global_GameFontBlackSmall
global GameFontBlackSmall2: W.Global_GameFontBlackSmall2
global GameFontBlackTiny: W.Global_GameFontBlackTiny
global GameFontBlackTiny2: W.Global_GameFontBlackTiny2
global GameFontDarkGraySmall: W.Global_GameFontDarkGraySmall
global GameFontDisable: W.Global_GameFontDisable
global GameFontDisableHuge: W.Global_GameFontDisableHuge
global GameFontDisableLarge: W.Global_GameFontDisableLarge
global GameFontDisableLeft: W.Global_GameFontDisableLeft
global GameFontDisableMed2: W.Global_GameFontDisableMed2
global GameFontDisableMed3: W.Global_GameFontDisableMed3
global GameFontDisableOutline22: W.Global_GameFontDisableOutline22
global GameFontDisableSmall: W.Global_GameFontDisableSmall
global GameFontDisableSmall2: W.Global_GameFontDisableSmall2
global GameFontDisableSmallLeft: W.Global_GameFontDisableSmallLeft
global GameFontDisableTiny: W.Global_GameFontDisableTiny
global GameFontDisableTiny2: W.Global_GameFontDisableTiny2
global GameFontGreen: W.Global_GameFontGreen
global GameFontGreenLarge: W.Global_GameFontGreenLarge
global GameFontGreenSmall: W.Global_GameFontGreenSmall
global GameFontHighlight: W.Global_GameFontHighlight
global GameFontHighlightCenter: W.Global_GameFontHighlightCenter
global GameFontHighlightExtraSmall: W.Global_GameFontHighlightExtraSmall
global GameFontHighlightExtraSmallLeft: W.Global_GameFontHighlightExtraSmallLeft
global GameFontHighlightExtraSmallLeftTop: W.Global_GameFontHighlightExtraSmallLeftTop
global GameFontHighlightHuge: W.Global_GameFontHighlightHuge
global GameFontHighlightHuge2: W.Global_GameFontHighlightHuge2
global GameFontHighlightLarge: W.Global_GameFontHighlightLarge
global GameFontHighlightLarge2: W.Global_GameFontHighlightLarge2
global GameFontHighlightLeft: W.Global_GameFontHighlightLeft
global GameFontHighlightMed2: W.Global_GameFontHighlightMed2
global GameFontHighlightMedium: W.Global_GameFontHighlightMedium
global GameFontHighlightOutline: W.Global_GameFontHighlightOutline
global GameFontHighlightOutline22: W.Global_GameFontHighlightOutline22
global GameFontHighlightRight: W.Global_GameFontHighlightRight
global GameFontHighlightShadowHuge2: W.Global_GameFontHighlightShadowHuge2
global GameFontHighlightShadowOutline22: W.Global_GameFontHighlightShadowOutline22
global GameFontHighlightSmall: W.Global_GameFontHighlightSmall
global GameFontHighlightSmall2: W.Global_GameFontHighlightSmall2
global GameFontHighlightSmallLeft: W.Global_GameFontHighlightSmallLeft
global GameFontHighlightSmallLeftTop: W.Global_GameFontHighlightSmallLeftTop
global GameFontHighlightSmallOutline: W.Global_GameFontHighlightSmallOutline
global GameFontHighlightSmallRight: W.Global_GameFontHighlightSmallRight
global GameFontNormal: W.Global_GameFontNormal
global GameFontNormalCenter: W.Global_GameFontNormalCenter
global GameFontNormalGraySmall: W.Global_GameFontNormalGraySmall
global GameFontNormalHuge: W.Global_GameFontNormalHuge
global GameFontNormalHuge2: W.Global_GameFontNormalHuge2
global GameFontNormalHuge3: W.Global_GameFontNormalHuge3
global GameFontNormalHuge3Outline: W.Global_GameFontNormalHuge3Outline
global GameFontNormalHuge4: W.Global_GameFontNormalHuge4
global GameFontNormalHuge4Outline: W.Global_GameFontNormalHuge4Outline
global GameFontNormalHugeBlack: W.Global_GameFontNormalHugeBlack
global GameFontNormalHugeOutline: W.Global_GameFontNormalHugeOutline
global GameFontNormalLarge: W.Global_GameFontNormalLarge
global GameFontNormalLarge2: W.Global_GameFontNormalLarge2
global GameFontNormalLargeLeft: W.Global_GameFontNormalLargeLeft
global GameFontNormalLargeLeftTop: W.Global_GameFontNormalLargeLeftTop
global GameFontNormalLargeOutline: W.Global_GameFontNormalLargeOutline
global GameFontNormalLeft: W.Global_GameFontNormalLeft
global GameFontNormalLeftBottom: W.Global_GameFontNormalLeftBottom
global GameFontNormalLeftGreen: W.Global_GameFontNormalLeftGreen
global GameFontNormalLeftGrey: W.Global_GameFontNormalLeftGrey
global GameFontNormalLeftLightGreen: W.Global_GameFontNormalLeftLightGreen
global GameFontNormalLeftOrange: W.Global_GameFontNormalLeftOrange
global GameFontNormalLeftRed: W.Global_GameFontNormalLeftRed
global GameFontNormalLeftYellow: W.Global_GameFontNormalLeftYellow
global GameFontNormalMed1: W.Global_GameFontNormalMed1
global GameFontNormalMed2: W.Global_GameFontNormalMed2
global GameFontNormalMed2Outline: W.Global_GameFontNormalMed2Outline
global GameFontNormalMed3: W.Global_GameFontNormalMed3
global GameFontNormalMed3Outline: W.Global_GameFontNormalMed3Outline
global GameFontNormalOutline: W.Global_GameFontNormalOutline
global GameFontNormalOutline22: W.Global_GameFontNormalOutline22
global GameFontNormalRight: W.Global_GameFontNormalRight
global GameFontNormalShadowHuge2: W.Global_GameFontNormalShadowHuge2
global GameFontNormalShadowOutline22: W.Global_GameFontNormalShadowOutline22
global GameFontNormalSmall: W.Global_GameFontNormalSmall
global GameFontNormalSmall2: W.Global_GameFontNormalSmall2
global GameFontNormalSmallBattleNetBlueLeft: W.Global_GameFontNormalSmallBattleNetBlueLeft
global GameFontNormalSmallLeft: W.Global_GameFontNormalSmallLeft
global GameFontNormalTiny: W.Global_GameFontNormalTiny
global GameFontNormalTiny2: W.Global_GameFontNormalTiny2
global GameFontNormalWTF2: W.Global_GameFontNormalWTF2
global GameFontNormalWTF2Outline: W.Global_GameFontNormalWTF2Outline
global GameFontNormal_NoShadow: W.Global_GameFontNormal_NoShadow
global GameFontRed: W.Global_GameFontRed
global GameFontRedLarge: W.Global_GameFontRedLarge
global GameFontRedSmall: W.Global_GameFontRedSmall
global GameFontWhite: W.Global_GameFontWhite
global GameFontWhiteSmall: W.Global_GameFontWhiteSmall
global GameFontWhiteTiny: W.Global_GameFontWhiteTiny
global GameFontWhiteTiny2: W.Global_GameFontWhiteTiny2
global GameFont_Gigantic: W.Global_GameFont_Gigantic
global GameNormalNumberFont: W.Global_GameNormalNumberFont
global GameTooltip: W.Global_GameTooltip
global GameTooltipHeader: W.Global_GameTooltipHeader
global GameTooltipHeaderText: W.Global_GameTooltipHeaderText
global GameTooltipText: W.Global_GameTooltipText
global GameTooltipTextSmall: W.Global_GameTooltipTextSmall
global GetAccountExpansionLevel: function(): number
global GetAchievementCategory: function(number): number
global GetAchievementComparisonInfo: function(number): boolean, number, number, number
global GetAchievementCriteriaInfo: function(number, number, ? boolean): string, number, boolean, number, number, string, number, number, string, number, boolean, number, number
global GetAchievementCriteriaInfoByID: function(number, number): string, number, boolean, number, number, string, number, number, string, number, boolean, number, number
global GetAchievementGuildRep: function(number): boolean, boolean, number
global GetAchievementInfo: W.Global_GetAchievementInfo
global GetAchievementLink: function(number): string
global GetAchievementNumCriteria: function(number): number
global GetAchievementNumRewards: function(unknown): unknown...
global GetAchievementReward: function(unknown, unknown): unknown...
global GetAchievementSearchProgress: function(): unknown...
global GetAchievementSearchSize: function(): unknown...
global GetActionBarToggles: function(): boolean, boolean, boolean, boolean, boolean, boolean, boolean
global GetActionInfo: function(number): string, number | string, string
global GetActiveArtifactByRace: function(number, number): string, string, string, string, string, number, string
global GetActiveLevel: function(unknown): unknown...
global GetActiveLootRollIDs: function(): unknown...
global GetActiveQuestID: function(unknown): unknown...
global GetActiveSpecGroup: function(? boolean, ? boolean): number
global GetActiveTitle: function(unknown): unknown...
global GetAddOnCPUUsage: function(string | number): number
global GetAddOnMemoryUsage: function(string | number): number
global GetAllowLowLevelRaid: function(): boolean
global GetAllowRecentAlliesSeeLocation: function(): boolean
global GetAlternativeDefaultLanguage: function(): unknown...
global GetArchaeologyInfo: function(): string
global GetArchaeologyRaceInfo: function(number): string, number, number, number, number, number
global GetArchaeologyRaceInfoByID: function(number): string, number, number, number, number, number
global GetAreaSpiritHealerTime: function(): number
global GetAreaText: function(): string
global GetArenaOpponentSpec: function(number): number, number
global GetArtifactInfoByRace: function(number, number): string, string, number, string, string, number, string, number, number
global GetArtifactProgress: function(): number, number, number
global GetAttackPowerForStat: function(number, number): W.NumberLike
global GetAutoDeclineGuildInvites: function(): boolean
global GetAutoDeclineNeighborhoodInvites: function(): boolean
global GetAutoQuestPopUp: function(number): number, string
global GetAvailableBandwidth: function(): number
global GetAvailableLevel: function(unknown): unknown...
global GetAvailableLocaleInfo: function(? boolean): {W.LocaleInfo}
global GetAvailableLocales: function(? boolean): string...
global GetAvailableQuestInfo: function(number): boolean, number, boolean, boolean, number, boolean
global GetAvailableTitle: function(unknown): unknown...
global GetAverageItemLevel: function(): number, number, number
global GetAvoidance: function(): W.NumberLike
global GetBackgroundLoadingStatus: function(): number
global GetBaseDifficultyID: function(number): number
global GetBattlefieldArenaFaction: function(): number
global GetBattlefieldEstimatedWaitTime: function(): number
global GetBattlefieldInstanceExpiration: function(): number
global GetBattlefieldInstanceRunTime: function(): number
global GetBattlefieldMapIconScale: function(): unknown...
global GetBattlefieldPortExpiration: function(number): number
global GetBattlefieldStatus: function(number): string, string | nil, number, boolean, boolean, string | nil, string | nil, string | nil, boolean | nil, string | nil, string | nil, boolean
global GetBattlefieldTeamInfo: function(number): string, number, number, number
global GetBattlefieldTimeWaited: function(number): number
global GetBattlefieldWinner: function(): number | nil
global GetBattlegroundPoints: function(number): number, number
global GetBestFlexRaidChoice: function(): number
global GetBestRFChoice: function(): number
global GetBillingTimeRested: function(): number
global GetBindLocation: function(): string
global GetBinding: function(number, ? boolean): string, string, string | nil, string | nil, string | nil...
global GetBindingAction: function(string, ? boolean): string
global GetBindingKey: function(string): string...
global GetBindingText: function(? string, ? string, ? boolean): string
global GetBlockChance: function(): W.NumberLike
global GetBuildInfo: function(): string, string, string, number, string, string
global GetBuildOption: function(string): boolean | nil
global GetButtonMetatable: function(): {unknown: unknown}
global GetBuybackItemInfo: function(number): string | nil, number | nil, number, number, number, boolean, boolean | nil
global GetBuybackItemLink: function(unknown): unknown...
global GetCVar: function(string): string | nil
global GetCVarBitfield: function(string, number): boolean | nil
global GetCVarBool: function(string): boolean | nil
global GetCVarDefault: function(string): string | nil
global GetCallPetSpellInfo: function(unknown): unknown...
global GetCallstackHeight: function(): number
global GetCameraFOVDefaults: function(): number, number, number
global GetCameraZoom: function(): number
global GetCategoryAchievementPoints: function(unknown, unknown): unknown...
global GetCategoryInfo: function(number): string, number, number
global GetCategoryList: function(): {number}
global GetCategoryNumAchievements: function(number, ? boolean): number, number, number
global GetCemeteryPreference: function(): number
global GetChannelDisplayInfo: function(number): string, boolean, boolean | nil, number | nil, number | nil, boolean | nil, string, number
global GetChannelList: function(): number, string, boolean, unknown...
global GetChannelName: function(string | number): number, string, number, boolean
global GetChatTypeIndex: function(string): number
global GetChatWindowChannels: function(number): string, number, unknown...
global GetChatWindowInfo: function(number): string, number, number, number, number, number, boolean, boolean, number, boolean
global GetChatWindowMessages: function(number): string...
global GetChatWindowSavedDimensions: function(unknown): unknown...
global GetChatWindowSavedPosition: function(unknown): unknown...
global GetClassColor: function(W.Literal_8c18208578b0): number, number, number, string
global GetClassColorObj: function(string): W.colorRGB
global GetClassColoredTextForUnit: function(string, string): string
global GetClassInfo: function(number): string, string, number
global GetClassicExpansionLevel: function(): number
global GetClickFrame: function(string): {unknown: unknown} | nil
global GetClientDisplayExpansionLevel: function(): number
global GetCoinIcon: function(number): number
global GetCoinText: function(number, ? string): string
global GetCoinTextureString: function(number, ? number): string
global GetCollapsingStarCost: function(): number
global GetCombatRating: function(number): W.NumberLike | nil
global GetCombatRatingBonus: function(number): W.NumberLike | nil
global GetCombatRatingBonusForCombatRatingValue: function(number, number): number | nil
global GetComboPoints: function(string, string): W.NumberLike
global GetCompanionInfo: function(string, number): number, string, number, string, boolean, number
global GetComparisonAchievementPoints: function(): unknown...
global GetComparisonCategoryNumAchievements: function(unknown): unknown...
global GetComparisonStatistic: function(number): string
global GetCorpseRecoveryDelay: function(): number
global GetCorruption: function(): number
global GetCorruptionResistance: function(): number
global GetCritChance: function(): W.NumberLike
global GetCritChanceProvidesParryEffect: function(): boolean
global GetCriteriaSpell: function(): unknown...
global GetCurrenciesString: function({number | {unknown: unknown}}): string | nil
global GetCurrencyString: function(number, ? number, ? string, ? boolean): string
global GetCurrentArenaSeason: function(): number
global GetCurrentBindingSet: function(): number
global GetCurrentCinematicSummary: function(): string
global GetCurrentEnvironment: function(): unknown...
global GetCurrentEventID: function(): number | nil
global GetCurrentGlyphNameForSpell: function(unknown): unknown...
global GetCurrentGraphicsAPI: function(): string
global GetCurrentGraphicsSetting: function(): unknown...
global GetCurrentGuildBankTab: function(): unknown...
global GetCurrentKeyBoardFocus: function(): unknown...
global GetCurrentLevelFeatures: function(unknown): unknown...
global GetCurrentRegion: function(): number
global GetCurrentRegionName: function(): string
global GetCurrentScaledResolution: function(): unknown...
global GetCurrentTitle: function(): number
global GetCursorDelta: function(): number, number
global GetCursorInfo: function(): string | nil, W.OpaqueValue...
global GetCursorMoney: function(): number
global GetCursorPosition: function(): number, number
global GetDailyQuestsCompleted: function(): unknown...
global GetDefaultAbbreviationBreakpoints: function(? W.WowLocale): {W.NumberAbbreviationBreakpoint}
global GetDefaultLanguage: function(): string, number
global GetDefaultScale: function(): number
global GetDemotionRank: function(unknown): unknown...
global GetDetailedItemLevelInfo: function(number | string): number, boolean, number
global GetDifficultyInfo: function(number): string, string, boolean, boolean, boolean, boolean, number | nil, boolean, number | nil, number | nil, boolean
global GetDodgeChance: function(): W.NumberLike
global GetDodgeChanceFromAttribute: function(): W.NumberLike
global GetDownloadedPercentage: function(): number
global GetDungeonDifficultyID: function(): number
global GetDungeonForRandomSlot: function(unknown, unknown): unknown...
global GetEditBoxMetatable: function(): {unknown: unknown}
global GetEquipmentNameFromSpell: function(): unknown...
global GetErrorCallstackHeight: function(): number | nil
global GetEventCPUUsage: function(): number, number
global GetEventTime: function(number): number | nil, number | nil, string | nil, number | nil
global GetExpansionDisplayInfo: function(number, ? number): W.ExpansionDisplayInfo | nil
global GetExpansionForLevel: function(number): number | nil
global GetExpansionLevel: function(): number
global GetExpansionTrialInfo: function(): boolean, number | nil
global GetExpertise: function(): W.NumberLike, W.NumberLike, W.NumberLike
global GetExpertisePercent: function(): W.NumberLike, W.NumberLike, W.NumberLike
global GetFactionColor: function(string): W.colorRGBA
global GetFailedPVPTalentIDs: function(): unknown...
global GetFailedTalentIDs: function(): unknown...
global GetFileIDFromPath: function(string): number
global GetFileStreamingStatus: function(): number
global GetFilteredAchievementID: function(number): number
global GetFinalAtlasFromTextureKitIfExists: function(string, string | {unknown: unknown}): string | nil
global GetFinalNameFromTextureKit: function(string, string | {unknown: unknown}): string
global GetFlexRaidDungeonInfo: function(unknown): unknown...
global GetFlyoutID: function(unknown): unknown...
global GetFlyoutInfo: function(number): string, string, number, boolean
global GetFlyoutSlotInfo: function(number, number): number, number, boolean, string, number
global GetFollowerTypeIDFromSpell: function(): unknown...
global GetFontInfo: function(W.Font): W.FontScriptInfo | nil
global GetFontStringMetatable: function(): {unknown: unknown}
global GetFonts: function(): {string}
global GetFrameCPUUsage: function(W.Frame, ? boolean): number, number
global GetFrameMetatable: function(): {unknown: unknown}
global GetFramerate: function(): number
global GetFramesRegisteredForEvent: function(string): W.Frame...
global GetFunctionCPUUsage: function(): number, number
global GetGMStatus: function(): unknown...
global GetGMTicket: function(): unknown...
global GetGameMessageInfo: function(number): string | nil, number | nil, number | nil
global GetGameTime: function(): number, number
global GetGlobalEnvironment: function(): unknown...
global GetGraphicsAPIs: function(): string
global GetGraphicsCVarValueForQualityLevel: function(): unknown...
global GetGreetingText: function(): unknown...
global GetGroupMemberCounts: function(): unknown...
global GetGuildAchievementMemberInfo: function(number, number): string | nil
global GetGuildAchievementMembers: function(number): unknown...
global GetGuildAchievementNumMembers: function(number): number
global GetGuildBankBonusDepositMoney: function(): unknown...
global GetGuildBankItemInfo: function(number, number): number, number, boolean, boolean, number
global GetGuildBankItemLink: function(number, number): string
global GetGuildBankMoney: function(): number
global GetGuildBankMoneyTransaction: function(number): string, string, number, number, number, number, number
global GetGuildBankTabCost: function(): unknown...
global GetGuildBankTabInfo: function(number): string, string, boolean, boolean, number, number, boolean
global GetGuildBankTabPermissions: function(number): boolean, boolean, boolean, number
global GetGuildBankText: function(unknown): unknown...
global GetGuildBankTransaction: function(number, number): string, string, string, number, number, number, number, number, number, number
global GetGuildBankWithdrawGoldLimit: function(): number
global GetGuildBankWithdrawMoney: function(): number
global GetGuildCategoryList: function(): unknown...
global GetGuildChallengeInfo: function(unknown): unknown...
global GetGuildCharterCost: function(): unknown...
global GetGuildEventInfo: function(unknown): unknown...
global GetGuildFactionGroup: function(): unknown...
global GetGuildInfo: function(string): string, string, number, string | nil
global GetGuildLogoInfo: function(): unknown...
global GetGuildMemberRecipes: function(unknown, unknown): unknown...
global GetGuildNewsFilters: function(): unknown...
global GetGuildNewsMemberName: function(unknown, unknown): unknown...
global GetGuildNewsSort: function(): unknown...
global GetGuildPerkInfo: function(): unknown...
global GetGuildRecipeInfoPostQuery: function(): number, number, number
global GetGuildRecipeMember: function(number): string, boolean
global GetGuildRenameRequired: function(): unknown...
global GetGuildRewardInfo: function(unknown): unknown...
global GetGuildRosterInfo: function(number): string, string, number, number, string, string, string, string, boolean, number, string, number, number, boolean, boolean, number, string
global GetGuildRosterLargestAchievementPoints: function(): unknown...
global GetGuildRosterLastOnline: function(number): number, number, number, number
global GetGuildRosterSelection: function(): number
global GetGuildRosterShowOffline: function(): boolean
global GetGuildTabardFiles: function(): number, number, number, number, number, number
global GetGuildTradeSkillInfo: function(number): number, boolean, string, string, number, number, number, string, string, string, boolean, string, number, string, boolean, number
global GetHaste: function(): W.NumberLike
global GetHighlightedNumberDifferenceString: function(string, string): string
global GetHitModifier: function(): W.NumberLike
global GetHomePartyInfo: function({unknown: unknown}): {unknown: unknown}
global GetInboxHeaderInfo: function(number): string, string, string, string, number, number, number, number, boolean, boolean, boolean, boolean, boolean
global GetInboxInvoiceInfo: function(number): string | nil, string | nil, string | nil, number, number, number, number, number, number, number, number, boolean
global GetInboxItem: function(number, number): string, number, string, number, number, boolean
global GetInboxItemLink: function(number, number): string
global GetInboxNumItems: function(): number, number
global GetInboxText: function(number): string, string, string, boolean, boolean, boolean
global GetInspectArenaData: function(number): number, number, number, number, number
global GetInspectHonorData: function(): number, number, number, number, number, number
global GetInspectSpecialization: function(string): number
global GetInspectTalent: function(): unknown...
global GetInstanceBootTimeRemaining: function(): number
global GetInstanceInfo: function(): string, string, number, string, number, number, boolean | nil, number, number, number | nil, boolean
global GetInstanceLockTimeRemaining: function(): number, boolean, number, number
global GetInstanceLockTimeRemainingEncounter: function(number): string, string, boolean, boolean
global GetInventoryAlertStatus: function(string): number
global GetInventoryItemBroken: function(string, number): boolean
global GetInventoryItemCooldown: function(string, number): number, number, number
global GetInventoryItemCount: function(string, number): number
global GetInventoryItemDurability: function(number): number, number
global GetInventoryItemEquippedUnusable: function(unknown, unknown): unknown...
global GetInventoryItemID: function(string, number): number, number
global GetInventoryItemLink: function(string, number): string | nil
global GetInventoryItemQuality: function(string, number): number
global GetInventoryItemTexture: function(string, number): number
global GetInventoryItemsForSlot: function(number, {unknown: unknown}, ? boolean): {unknown: unknown}
global GetInventorySlotInfo: function(W.Literal_7a28f0f291f7): number, string, boolean
global GetInviteConfirmationInfo: function(unknown): number, string, string, boolean, boolean, number, number, number
global GetItemChildInfo: function(number | string, ? number): {number}
global GetItemClassInfo: function(number): string
global GetItemCooldown: function(number | string): number, number, boolean
global GetItemCount: function(number | string, ? boolean, ? boolean, ? boolean): number
global GetItemCreationContext: function(number | string): number, string
global GetItemFamily: function(number | string): number | nil
global GetItemGem: function(string, number): string, string
global GetItemIcon: function(number | string): number | nil
global GetItemInfo: function(number | string): string, string, number, number, number, string, string, number, string, number, number, number, number, number, number, number | nil, boolean
global GetItemInfoInstant: function(number | string): number, string, string, string, number, number, number
global GetItemInventorySlotInfo: function(number): string
global GetItemLevelColor: function(): number, number, number
global GetItemQualityColor: function(number): number, number, number, string
global GetItemSetInfo: function(number): string
global GetItemSpecInfo: function(number | string): {number}
global GetItemSpell: function(number | string): string, number
global GetItemSubClassInfo: function(number, number): string, boolean
global GetItemUniqueness: function(number | string): number, number
global GetJailersTowerLevel: function(): number
global GetJournalInfoForSpellConfirmation: function(unknown): unknown...
global GetLFDChoiceCollapseState: function(unknown): unknown...
global GetLFDChoiceEnabledState: function(unknown): unknown...
global GetLFDChoiceOrder: function(unknown): unknown...
global GetLFDLockInfo: function(unknown, unknown): unknown...
global GetLFDLockPlayerCount: function(): unknown...
global GetLFDRoleLockInfo: function(unknown, unknown): unknown...
global GetLFDRoleRestrictions: function(unknown): unknown...
global GetLFGBootProposal: function(): boolean, boolean, boolean, string, number, number, number, string
global GetLFGCategoryForID: function(unknown): unknown...
global GetLFGCompletionReward: function(): unknown...
global GetLFGCompletionRewardItem: function(unknown): unknown...
global GetLFGCompletionRewardItemLink: function(unknown): unknown...
global GetLFGDeserterExpiration: function(): number | nil
global GetLFGDungeonEncounterInfo: function(number, number): string, string, boolean, boolean
global GetLFGDungeonInfo: function(number): string, number, number, number, number, number, number, number, number, number, string, number, number, string, boolean, number, number, boolean, string, number, boolean, number
global GetLFGDungeonNumEncounters: function(number): number, number
global GetLFGDungeonRewardCapBarInfo: function(number): number, number, number, number, number, number, number, number, number, number
global GetLFGDungeonRewardCapInfo: function(unknown): unknown...
global GetLFGDungeonRewardInfo: function(unknown, unknown): unknown...
global GetLFGDungeonRewardLink: function(unknown, unknown): unknown...
global GetLFGDungeonRewards: function(unknown): unknown...
global GetLFGDungeonShortageRewardInfo: function(unknown, unknown, unknown): unknown...
global GetLFGDungeonShortageRewardLink: function(unknown, unknown, unknown): unknown...
global GetLFGInfoServer: function(unknown, unknown): unknown...
global GetLFGInviteRoleAvailability: function(unknown): unknown...
global GetLFGInviteRoleRestrictions: function(unknown): unknown...
global GetLFGProposal: function(): boolean, number, number, number, string, string, string, boolean, number, number, number, boolean, boolean, number
global GetLFGProposalEncounter: function(unknown): unknown...
global GetLFGProposalMember: function(unknown): unknown...
global GetLFGQueueStats: function(number, ? number): boolean, boolean, boolean, boolean, boolean, number, number, number, number, number, string, number, number, number, number, number, number, number
global GetLFGQueuedList: function(unknown, unknown): unknown...
global GetLFGRandomCooldownExpiration: function(): number | nil
global GetLFGRandomDungeonInfo: function(number): number, string, number, number, number, number, number, number, number, number, number, string, number, number, string, boolean, number, number, boolean, string, number, boolean
global GetLFGReadyCheckUpdate: function(): unknown...
global GetLFGReadyCheckUpdateBattlegroundInfo: function(): unknown...
global GetLFGRoleShortageRewards: function(number, number): boolean, boolean, boolean, boolean, number, number, number
global GetLFGRoleUpdate: function(): unknown...
global GetLFGRoleUpdateBattlegroundInfo: function(): string
global GetLFGRoleUpdateMember: function(unknown): unknown...
global GetLFGRoleUpdateSlot: function(number): number, number, number
global GetLFGRoles: function(): boolean, boolean, boolean, boolean
global GetLFGSuspendedPlayers: function(unknown): unknown...
global GetLFRChoiceOrder: function({unknown: unknown}): {unknown: unknown}
global GetLanguageByIndex: function(number): string, number
global GetLatestCompletedAchievements: function(): unknown...
global GetLatestCompletedComparisonAchievements: function(): unknown...
global GetLatestThreeSenders: function(): string | nil, string | nil, string | nil
global GetLatestUpdatedComparisonStats: function(): unknown...
global GetLatestUpdatedStats: function(): unknown...
global GetLegacyRaidDifficultyID: function(): number | nil
global GetLifesteal: function(): W.NumberLike
global GetLocalGameTime: function(): number, number
global GetLocale: function(): string
global GetLooseMacroIcons: function(): unknown...
global GetLooseMacroItemIcons: function(): unknown...
global GetLootInfo: function(): {{unknown: unknown}}
global GetLootRollItemInfo: function(number): string, string, number, number, boolean, boolean, boolean, boolean, number, number, number, number, boolean
global GetLootRollItemLink: function(number): string
global GetLootRollTimeLeft: function(number): number | nil
global GetLootSlotInfo: function(number): string, string, number, number, number, boolean, boolean, number, boolean, boolean
global GetLootSlotLink: function(number): string
global GetLootSlotType: function(number): number
global GetLootSourceInfo: function(number): string, number
global GetLootSpecialization: function(): number
global GetLootThreshold: function(): number
global GetMacroBody: function(number | string): string | nil
global GetMacroIcons: function(unknown): unknown...
global GetMacroIndexByName: function(string): number
global GetMacroInfo: function(number | string): string, number, string
global GetMacroItem: function(number | string): string, string
global GetMacroItemIcons: function(unknown): unknown...
global GetMacroSpell: function(number | string): number
global GetManaRegen: function(): W.NumberLike, W.NumberLike
global GetMasterLootCandidate: function(number, number): string
global GetMastery: function(): W.NumberLike
global GetMasteryEffect: function(): W.NumberLike, W.NumberLike
global GetMaxBattlefieldID: function(): number
global GetMaxLevelForExpansionLevel: function(number): number
global GetMaxLevelForLatestExpansion: function(): number
global GetMaxLevelForPlayerExpansion: function(): number
global GetMaxNumCUFProfiles: function(): unknown...
global GetMaxPlayerLevel: function(): number
global GetMaxRenderScale: function(): unknown...
global GetMaxRewardCurrencies: function(): unknown...
global GetMaxSpellStartRecoveryOffset: function(): number
global GetMaxTalentTier: function(): number
global GetMaximumExpansionLevel: function(): number
global GetMeleeHaste: function(): W.NumberLike
global GetMerchantCurrencies: function(): unknown...
global GetMerchantFilter: function(): unknown...
global GetMerchantItemCostInfo: function(number): number
global GetMerchantItemCostItem: function(number, number): string, number, string, string
global GetMerchantItemID: function(number): number
global GetMerchantItemInfo: function(number): string, number, number, number, number, boolean, boolean, boolean, number, number
global GetMerchantItemLink: function(number): string | nil
global GetMerchantItemMaxStack: function(number): number
global GetMerchantNumItems: function(): number
global GetMicroIconForRole: function(string): string
global GetMinRenderScale: function(): unknown...
global GetMinimapZoneText: function(): string
global GetMinimumExpansionLevel: function(): number
global GetMirrorTimerInfo: function(number): string | nil, number | nil, number | nil, number | nil, number | nil, string | nil, number | nil
global GetMirrorTimerProgress: function(string): number | nil
global GetModResilienceDamageReduction: function(): W.NumberLike
global GetModifiedClick: function(string): string
global GetModifiedClickAction: function(unknown): unknown...
global GetMoney: function(): number
global GetMoneyString: function(number, ? boolean): string
global GetMonitorAspectRatio: function(): unknown...
global GetMonitorCount: function(): unknown...
global GetMonitorName: function(): unknown...
global GetMouseButtonClicked: function(): string
global GetMouseButtonName: function(W.Literal_b7e0de2fb994): string
global GetMouseFoci: function(): {W.ScriptRegion}
global GetMovieDownloadProgress: function(number): boolean, number, number
global GetMultiCastTotemSpells: function(number): number, number, number, number, number, number, number
global GetNativeRealmID: function(): number
global GetNegativeCorruptionEffectInfo: function(): {W.CorruptionEffectInfo}
global GetNetIpTypes: function(): number...
global GetNetStats: function(): number, number, number...
global GetNextAchievement: function(number): number | nil, boolean | nil
global GetNextCompleatedTutorial: function(unknown): unknown...
global GetNextPendingInviteConfirmation: function(): unknown...
global GetNormalizedRealmName: function(): string
global GetNumActiveQuests: function(): number
global GetNumArchaeologyRaces: function(): number
global GetNumArenaOpponentSpecs: function(): unknown...
global GetNumArenaOpponents: function(): unknown...
global GetNumArtifactsByRace: function(number): number
global GetNumAutoQuestPopUps: function(): number
global GetNumAvailableQuests: function(): number
global GetNumBattlefieldFlagPositions: function(): unknown...
global GetNumBattlefieldScores: function(): number
global GetNumBattlefieldVehicles: function(unknown): unknown...
global GetNumBattlegroundTypes: function(): number
global GetNumBindings: function(): number
global GetNumBuybackItems: function(): number
global GetNumChannelMembers: function(): unknown...
global GetNumClasses: function(): number
global GetNumCompanions: function(string): number
global GetNumComparisonCompletedAchievements: function(number): number, number
global GetNumCompletedAchievements: function(): number, number
global GetNumDeclensionSets: function(string, ? number): number
global GetNumDisplayChannels: function(): number
global GetNumDungeonForRandomSlot: function(unknown): unknown...
global GetNumExpansions: function(): number
global GetNumFilteredAchievements: function(): number
global GetNumFlexRaidDungeons: function(): number
global GetNumFlyouts: function(): unknown...
global GetNumGroupChannels: function(): unknown...
global GetNumGroupMembers: function(? number): number
global GetNumGuildBankMoneyTransactions: function(): unknown...
global GetNumGuildBankTabs: function(): unknown...
global GetNumGuildBankTransactions: function(unknown): unknown...
global GetNumGuildChallenges: function(): unknown...
global GetNumGuildEvents: function(): unknown...
global GetNumGuildMembers: function(): number, number
global GetNumGuildNews: function(): unknown...
global GetNumGuildPerks: function(): unknown...
global GetNumGuildRewards: function(): unknown...
global GetNumGuildTradeSkill: function(): unknown...
global GetNumLanguages: function(): number
global GetNumLootItems: function(): number
global GetNumMacros: function(): number, number
global GetNumMembersInRank: function(unknown): unknown...
global GetNumModifiedClickActions: function(): unknown...
global GetNumPetitionNames: function(): number
global GetNumQuestChoices: function(): number
global GetNumQuestCurrencies: function(): unknown...
global GetNumQuestItemDrops: function(): unknown...
global GetNumQuestItems: function(): number
global GetNumQuestLeaderBoards: function(? number): number
global GetNumQuestLogChoices: function(number, ? boolean): number
global GetNumQuestLogRewardFactions: function(): unknown...
global GetNumQuestLogRewards: function(? number): number
global GetNumQuestLogTasks: function(): unknown...
global GetNumQuestRewards: function(): number
global GetNumRFDungeons: function(): number
global GetNumRaidProfiles: function(): unknown...
global GetNumRandomDungeons: function(): number
global GetNumRandomScenarios: function(): unknown...
global GetNumRoutes: function(unknown): unknown...
global GetNumSavedInstances: function(): number
global GetNumSavedWorldBosses: function(): number
global GetNumScenarios: function(): unknown...
global GetNumShapeshiftForms: function(): number
global GetNumSpecGroups: function(boolean): number
global GetNumSpecializations: function(? boolean, ? boolean): number
global GetNumSpecializationsForClassID: function(number): number
global GetNumSubgroupMembers: function(? number): number
global GetNumTitles: function(): number
global GetNumTotemSlots: function(): number
global GetNumTrainerServices: function(): number
global GetNumTreasurePickerItems: function(): unknown...
global GetNumUnspentPvpTalents: function(): unknown...
global GetNumUnspentTalents: function(): number
global GetNumWarGameTypes: function(): unknown...
global GetOSLocale: function(): string
global GetObjectiveText: function(): unknown...
global GetOptOutOfLoot: function(): boolean
global GetOverrideAPBySpellPower: function(): W.NumberLike
global GetOverrideSpellPowerByAP: function(): W.NumberLike
global GetPVPDesired: function(): boolean
global GetPVPGearStatRules: function(): boolean
global GetPVPLifetimeStats: function(): number, number
global GetPVPRoles: function(): boolean, boolean, boolean
global GetPVPSessionStats: function(): number, number
global GetPVPTimer: function(): number
global GetPVPYesterdayStats: function(): number, number
global GetParryChance: function(): W.NumberLike
global GetParryChanceFromAttribute: function(): W.NumberLike
global GetPartyAssignment: function(string, ? string, ? boolean): boolean
global GetPartyLFGBackfillInfo: function(): unknown...
global GetPartyLFGID: function(): unknown...
global GetPendingGlyphName: function(): unknown...
global GetPendingInviteConfirmations: function(): unknown...
global GetPersonalRatedInfo: function(number): number, number, number, number, number, number, number, number
global GetPetActionCooldown: function(number): number, number, number
global GetPetActionInfo: function(number): string, string, boolean, boolean, boolean, boolean, number, boolean, boolean
global GetPetActionSlotUsable: function(number): boolean
global GetPetActionsUsable: function(): unknown...
global GetPetExperience: function(): number, number
global GetPetFoodTypes: function(): string...
global GetPetIcon: function(): unknown...
global GetPetMeleeHaste: function(): W.NumberLike
global GetPetSpellBonusDamage: function(): W.NumberLike
global GetPetTimeRemaining: function(): unknown...
global GetPetitionInfo: function(): string, string, string, number, string, boolean, number
global GetPetitionNameInfo: function(unknown): unknown...
global GetPhysicalScreenSize: function(): number, number
global GetPlayerFacing: function(): number | nil
global GetPlayerInfoByGUID: function(W.TextLike): string | nil, string | nil, string | nil, string | nil, number | nil, string | nil, string | nil, number | nil
global GetPlayerTradeCurrency: function(): unknown...
global GetPlayerTradeMoney: function(): number
global GetPossessInfo: function(number): string, number, boolean
global GetPowerRegen: function(): W.NumberLike, W.NumberLike
global GetPowerRegenForPowerType: function(number): W.NumberLike, W.NumberLike
global GetPrevCompleatedTutorial: function(unknown): unknown...
global GetPreviousAchievement: function(number): number
global GetPreviousArenaSeason: function(): unknown...
global GetPrimarySpecialization: function(): unknown...
global GetProfessionInfo: function(number): string, number, number, number, number, number, number, number, number, number, string
global GetProfessions: function(): number, number, number, number, number
global GetProgressText: function(): string
global GetPromotionRank: function(unknown): unknown...
global GetProtocolTypes: function(): number...
global GetPvpPowerDamage: function(): W.NumberLike
global GetPvpPowerHealing: function(): W.NumberLike
global GetPvpTalentInfoByID: function(number, ? number, ? boolean, ? string): number, string, number, boolean, boolean, number, boolean, number, number, boolean, boolean
global GetPvpTalentInfoBySpecialization: function(): unknown...
global GetPvpTalentLink: function(): unknown...
global GetQuestBackgroundMaterial: function(): string | nil
global GetQuestCurrencyID: function(unknown, unknown): unknown...
global GetQuestExpansion: function(unknown): unknown...
global GetQuestFactionGroup: function(number): number
global GetQuestID: function(): number
global GetQuestItemInfo: function(string, number): string, number, number, number, boolean, number
global GetQuestItemInfoLootType: function(unknown, unknown): unknown...
global GetQuestItemLink: function(string, number): string
global GetQuestLink: function(number): string
global GetQuestLogChoiceInfo: function(number, ? number): string, number, number, number, boolean, number
global GetQuestLogChoiceInfoLootType: function(unknown): unknown...
global GetQuestLogCompletionText: function(): unknown...
global GetQuestLogCriteriaSpell: function(): unknown...
global GetQuestLogItemDrop: function(unknown): unknown...
global GetQuestLogItemLink: function(string, number, ? number): string
global GetQuestLogLeaderBoard: function(number, ? number): string, string, boolean
global GetQuestLogPortraitTurnIn: function(): unknown...
global GetQuestLogQuestText: function(? number): string, string
global GetQuestLogQuestType: function(): unknown...
global GetQuestLogRewardArtifactXP: function(): unknown...
global GetQuestLogRewardFactionInfo: function(unknown): unknown...
global GetQuestLogRewardHonor: function(unknown): unknown...
global GetQuestLogRewardInfo: function(number, ? number): string, string, number, number, boolean, number, number
global GetQuestLogRewardMoney: function(? number): number
global GetQuestLogRewardSkillPoints: function(): unknown...
global GetQuestLogRewardTitle: function(): unknown...
global GetQuestLogRewardXP: function(unknown): unknown...
global GetQuestLogSpecialItemCooldown: function(number): number, number, number
global GetQuestLogSpecialItemInfo: function(number): string | nil, number, number, boolean
global GetQuestLogTimeLeft: function(): number
global GetQuestMoneyToGet: function(): unknown...
global GetQuestObjectiveInfo: function(number, number, boolean): string, string, boolean, number, number
global GetQuestPOIBlobCount: function(unknown): unknown...
global GetQuestPOILeaderBoard: function(unknown): unknown...
global GetQuestPOIs: function(): unknown...
global GetQuestPortraitGiver: function(): unknown...
global GetQuestPortraitTurnIn: function(): unknown...
global GetQuestProgressBarPercent: function(number): number
global GetQuestResetTime: function(): number
global GetQuestReward: function(number): unknown...
global GetQuestSortIndex: function(number): number
global GetQuestText: function(): unknown...
global GetQuestUiMapID: function(): unknown...
global GetRFDungeonInfo: function(number): number, string, number, number, number, number, number, number, number, number, number, string, number, number, string, boolean, number, number, boolean, string, number, boolean, number
global GetRaidDifficultyID: function(): number | nil
global GetRaidProfileFlattenedOptions: function(unknown): unknown...
global GetRaidProfileName: function(unknown): unknown...
global GetRaidProfileOption: function(unknown, unknown): unknown...
global GetRaidProfileSavedPosition: function(unknown): unknown...
global GetRaidRosterInfo: function(number): string, number, number, number, string, string, string | nil, boolean, boolean, string, boolean, string
global GetRaidTargetIndex: function(string): W.NumberLike | nil
global GetRandomDungeonBestChoice: function(): unknown...
global GetRandomScenarioBestChoice: function(): unknown...
global GetRandomScenarioInfo: function(unknown): unknown...
global GetRangedCritChance: function(): W.NumberLike
global GetRangedHaste: function(): W.NumberLike
global GetRatedBattleGroundInfo: function(): unknown...
global GetReadyCheckStatus: function(unknown): unknown...
global GetReadyCheckTimeLeft: function(): unknown...
global GetRealZoneText: function(? number): string
global GetRealmID: function(): number
global GetRealmName: function(): string
global GetReleaseTimeRemaining: function(): number
global GetRepairAllCost: function(): number, boolean | nil
global GetResSicknessDuration: function(): string | nil
global GetRestState: function(): number | nil, string | nil, number | nil
global GetRestrictedAccountData: function(): number, number, number
global GetRewardArtifactXP: function(): unknown...
global GetRewardHonor: function(): unknown...
global GetRewardMoney: function(): unknown...
global GetRewardNumSkillUps: function(): unknown...
global GetRewardPackArtifactPower: function(unknown): unknown...
global GetRewardPackCurrencies: function(unknown): unknown...
global GetRewardPackItems: function(unknown): unknown...
global GetRewardPackMoney: function(unknown): unknown...
global GetRewardPackTitle: function(unknown): unknown...
global GetRewardPackTitleName: function(unknown): unknown...
global GetRewardSkillLineID: function(): unknown...
global GetRewardSkillPoints: function(): unknown...
global GetRewardText: function(): string
global GetRewardTitle: function(): unknown...
global GetRewardXP: function(): number
global GetRuneCooldown: function(number): number | nil, number | nil, boolean | nil
global GetRuneCount: function(number): number | nil
global GetRunningMacro: function(): unknown...
global GetRunningMacroButton: function(): unknown...
global GetSavedInstanceChatLink: function(number): string
global GetSavedInstanceEncounterInfo: function(number, number): string, number, boolean, boolean
global GetSavedInstanceInfo: function(number): string, number, number, number, boolean, boolean, number, boolean, number, string, number, number, boolean, number
global GetSavedWorldBossInfo: function(number): string, number, number
global GetScenariosChoiceOrder: function(? {unknown: unknown}): {unknown: unknown}
global GetSchoolString: function(number): string
global GetScreenDPIScale: function(): number
global GetScreenHeight: function(): number
global GetScreenWidth: function(): number
global GetScriptBucketThrottleLimits: function(): W.ScriptBucketThrottleLimits
global GetScriptCPUUsage: function(): number
global GetSecondsUntilParentalControlsKick: function(): number | nil
global GetSelectedArtifactInfo: function(): string, string, number, number, string, number, number, number
global GetSelectedDisplayChannel: function(): unknown...
global GetSelectedWarGameType: function(): unknown...
global GetSendMailCOD: function(): number
global GetSendMailItem: function(number): string, number, number, number, number
global GetSendMailItemLink: function(number): string
global GetSendMailMoney: function(): unknown...
global GetSendMailPrice: function(): number
global GetServerExpansionLevel: function(): number
global GetServerTime: function(): number
global GetSessionTime: function(): number
global GetShapeshiftForm: function(? boolean): number
global GetShapeshiftFormCooldown: function(number): number, number, number
global GetShapeshiftFormID: function(): number
global GetShapeshiftFormInfo: function(number): string, boolean, boolean, number
global GetSheathState: function(): number | nil
global GetShieldBlock: function(): W.NumberLike
global GetSoundEntryCount: function(number): number | nil
global GetSourceLocation: function(): string | nil
global GetSpecialization: function(? boolean, ? boolean, ? number): number
global GetSpecializationInfo: function(number, ? boolean, ? boolean, ? string, ? number, ? number): number | nil, string | nil, string | nil, number | nil, string | nil, number | nil, number | nil, string | nil, number | nil, boolean | nil
global GetSpecializationInfoByID: function(number, ? number): number, string, string, number, string, string, string
global GetSpecializationInfoForClassID: function(number, number, ? number): number, string, string, number, string, boolean, boolean, number | nil, number | nil
global GetSpecializationInfoForSpecID: function(number, ? number): number, string, string, number, string, boolean, boolean, number | nil, number | nil
global GetSpecializationMasterySpells: function(number, ? boolean, ? boolean): number, number
global GetSpecializationNameForSpecID: function(number, ? number): string | nil
global GetSpecializationRole: function(number, ? boolean, ? boolean): string
global GetSpecializationRoleByID: function(number): string
global GetSpecializationRoleEnum: function(): unknown...
global GetSpecializationRoleEnumByID: function(): unknown...
global GetSpecializationSpells: function(number, ? boolean, ? boolean): number, number, number...
global GetSpecializationSystem: function(): number
global GetSpecsForSpell: function(): unknown...
global GetSpeed: function(): W.NumberLike
global GetSpellBaseCooldown: function(number): number, number
global GetSpellBonusDamage: function(number): W.NumberLike | nil
global GetSpellBonusHealing: function(): W.NumberLike
global GetSpellConfirmationPromptsInfo: function(): unknown...
global GetSpellCooldown: W.Global_GetSpellCooldown
global GetSpellCritChance: function(): W.NumberLike
global GetSpellHitModifier: function(): W.NumberLike
global GetSpellPenetration: function(): W.NumberLike
global GetSpellQueueWindow: function(): number
global GetSpellsForCharacterUpgradeTier: function(unknown): unknown...
global GetStatistic: function(number, ? number): string, boolean, string
global GetStatisticsCategoryList: function(): {unknown: unknown}
global GetStringFromModifiers: function(number): string
global GetSturdiness: function(): W.NumberLike
global GetSubZoneText: function(): string
global GetSuggestedGroupSize: function(): unknown...
global GetTabardCreationCost: function(): unknown...
global GetTabardInfo: function(): unknown...
global GetTalentInfo: function(number, number, number, ? boolean, ? string): number, string, number, boolean, boolean, number, boolean, number, number, boolean, boolean
global GetTalentInfoByID: function(unknown, unknown, unknown, unknown): unknown...
global GetTalentInfoBySpecialization: function(unknown, unknown, unknown): unknown...
global GetTalentLink: function(unknown, unknown, unknown, unknown, unknown): unknown...
global GetTalentTierInfo: function(number, number, ? boolean, ? string): boolean, number, number
global GetTargetTradeCurrency: function(): unknown...
global GetTargetTradeMoney: function(): number
global GetTaskInfo: function(number): boolean, boolean, number
global GetTaskPOIs: function(): unknown...
global GetTasksTable: function(): unknown...
global GetTaxiBenchmarkMode: function(): boolean
global GetTaxiMapID: function(): number
global GetTexCoordsByGrid: function(number, number, number, number, number, number): number, number, number, number
global GetTexCoordsForRole: function(string): number, number, number, number
global GetText: function(string, number, unknown): string
global GetTextureInfo: function(W.Texture): string, string, number, number, number, number, number, number, number, number
global GetThreatStatusColor: function(number): number, number, number
global GetTickTime: function(): number
global GetTime: function(): number
global GetTimePreciseSec: function(): number
global GetTitleName: function(number): string, boolean
global GetTitleText: function(): string
global GetTotalAchievementPoints: function(? boolean): number
global GetTotemCannotDismiss: function(number): boolean | nil
global GetTotemDuration: function(number): W.DurationObject
global GetTotemInfo: function(number): boolean | W.OpaqueValue | nil, W.TextLike | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil
global GetTotemTimeLeft: function(number): W.NumberLike | nil
global GetTradePlayerItemInfo: function(number): string, number, number, number, string, boolean
global GetTradePlayerItemLink: function(number): string
global GetTradeTargetItemInfo: function(number): string, string, number, number, number, string
global GetTradeTargetItemLink: function(number): string
global GetTrainerGreetingText: function(): string
global GetTrainerSelectionIndex: function(): number
global GetTrainerServiceAbilityReq: function(number, number): string, boolean
global GetTrainerServiceCost: function(number): number, number, number
global GetTrainerServiceDescription: function(number): string
global GetTrainerServiceIcon: function(number): number
global GetTrainerServiceInfo: function(number): string, string, string, number
global GetTrainerServiceItemLink: function(number): string
global GetTrainerServiceLevelReq: function(number): number
global GetTrainerServiceNumAbilityReq: function(): unknown...
global GetTrainerServiceSkillLine: function(number): string
global GetTrainerServiceSkillReq: function(number): string | nil, number, boolean
global GetTrainerServiceStepIndex: function(): unknown...
global GetTrainerServiceTypeFilter: function(string): boolean
global GetTrainerTradeskillRankValues: function(): unknown...
global GetTreasurePickerItemInfo: function(unknown, unknown): unknown...
global GetTutorialsEnabled: function(): unknown...
global GetUICameraInfo: function(number): number, number, number, number, number, number, number, number, number, boolean
global GetUnitChargedPowerPoints: function(string): W.OpaqueValue | nil
global GetUnitEmpowerHoldAtMaxTime: function(string): W.NumberLike
global GetUnitEmpowerMinHoldTime: function(string): W.NumberLike
global GetUnitEmpowerStageDuration: function(string, number): W.NumberLike
global GetUnitHealthModifier: function(string): number
global GetUnitMaxHealthModifier: function(string): number
global GetUnitPowerBarInfo: function(string): W.UnitPowerBarInfo | nil
global GetUnitPowerBarInfoByID: function(number): W.UnitPowerBarInfo | nil
global GetUnitPowerBarStrings: function(string): string | nil, string | nil, string | nil
global GetUnitPowerBarStringsByID: function(number): string | nil, string | nil, string | nil
global GetUnitPowerBarTextureInfo: function(string, number, ? number): number, number, number, number, number
global GetUnitPowerBarTextureInfoByID: function(number, number): number, number, number, number, number
global GetUnitPowerModifier: function(string): number
global GetUnitSpeed: function(string): W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike
global GetUnitTotalModifiedMaxHealthPercent: function(string): number
global GetUpgradeExpansionLevel: function(): number
global GetVehicleUIIndicator: function(number): number | nil, number | nil
global GetVehicleUIIndicatorSeat: function(number, number): number | nil, number | nil, number | nil
global GetVersatilityBonus: function(number): W.NumberLike
global GetVideoCaps: function(): unknown...
global GetWarGameQueueStatus: function(): unknown...
global GetWarGameTypeInfo: function(unknown): unknown...
global GetWeaponEnchantInfo: function(): boolean, number, number, number, boolean, number, number, number, boolean, number, number, number
global GetWebTicket: function(): unknown...
global GetWorldElapsedTime: function(number): string, number, number
global GetWorldElapsedTimers: function(): unknown...
global GetWorldMapActionButtonSpellInfo: function(): unknown...
global GetWorldPVPQueueStatus: function(number): string, string | nil, number, number, number, number, boolean
global GetXPExhaustion: function(): number | nil
global GetZonePVPInfo: function(): string, boolean, string | nil
global GetZoneText: function(): string
global GiveMasterLoot: function(number, number): unknown...
global GroupHasOfflineMember: function(): unknown...
global GuildControlAddRank: function(unknown): unknown...
global GuildControlDelRank: function(number): unknown...
global GuildControlGetAllowedShifts: function(unknown): unknown...
global GuildControlGetNumRanks: function(): unknown...
global GuildControlGetRank: function(): unknown...
global GuildControlGetRankName: function(number): string
global GuildControlSaveRank: function(string): unknown...
global GuildControlSetRank: function(number): unknown...
global GuildControlSetRankFlag: function(number, boolean): unknown...
global GuildControlShiftRankDown: function(unknown): unknown...
global GuildControlShiftRankUp: function(unknown): unknown...
global GuildDemote: function(string)
global GuildDisband: function()
global GuildInfo: function(): unknown...
global GuildInvite: function(string)
global GuildLeave: function()
global GuildMasterAbsent: function(): unknown...
global GuildNewsSetSticky: function(unknown, unknown): unknown...
global GuildNewsSort: function(unknown): unknown...
global GuildPromote: function(string)
global GuildSetLeader: function(string)
global GuildSetMOTD: function(string)
global GuildUninvite: function(string)
global HEALTHBAR_HEAL_ABSORB_COLOR: W.Global_HEALTHBAR_HEAL_ABSORB_COLOR
global HEALTHBAR_MY_HEAL_PREDICTION_COLOR: W.Global_HEALTHBAR_MY_HEAL_PREDICTION_COLOR
global HEALTHBAR_OTHER_HEAL_PREDICTION_COLOR: W.Global_HEALTHBAR_OTHER_HEAL_PREDICTION_COLOR
global HEALTHBAR_TOTAL_ABSORB_COLOR: W.Global_HEALTHBAR_TOTAL_ABSORB_COLOR
global HEIRLOOM_BLUE_COLOR: W.Global_HEIRLOOM_BLUE_COLOR
global HIGHLIGHT_FONT_COLOR: W.Global_HIGHLIGHT_FONT_COLOR
global HIGHLIGHT_LIGHT_BLUE: W.Global_HIGHLIGHT_LIGHT_BLUE
global HOSTILE_STATUS_COLOR: W.Global_HOSTILE_STATUS_COLOR
global HandleAtlasMemberCommand: function(): unknown...
global HasAPEffectsSpellPower: function(): boolean
global HasArtifactEquipped: function(): unknown...
global HasAttachedGlyph: function(unknown): unknown...
global HasCompletedAnyAchievement: function(unknown): unknown...
global HasDualWieldPenalty: function(): boolean
global HasFullControl: function(): boolean
global HasIgnoreDualWieldWeapon: function(): boolean
global HasInboxItem: function(): unknown...
global HasKey: function(): boolean
global HasLFGRestrictions: function(): boolean
global HasLoadedCUFProfiles: function(): unknown...
global HasLootSpecializations: function(): boolean
global HasNewMail: function(): unknown...
global HasNoReleaseAura: function(): boolean, number, boolean
global HasPendingGlyphCast: function(): unknown...
global HasPetUI: function(): boolean, boolean
global HasSPEffectsAttackPower: function(): boolean
global HasSendMailItem: function(unknown): unknown...
global HasTimePassed: function(number, number): boolean
global HasWandEquipped: function(): boolean
global HaveQuestData: function(unknown): unknown...
global HaveQuestRewardData: function(unknown): unknown...
global HearthAndResurrectFromArea: function(): unknown...
global HideRepairCursor: function(): unknown...
global HideUIPanel: function(W.Frame, ? boolean)
global IMEHighlight: W.Global_IMEHighlight
global IMENormal: W.Global_IMENormal
global IMPOSSIBLE_DIFFICULTY_COLOR: W.Global_IMPOSSIBLE_DIFFICULTY_COLOR
global INACTIVE_COLOR: W.Global_INACTIVE_COLOR
global INCREASE_STAT_COLOR: W.Global_INCREASE_STAT_COLOR
global INFLUENCE_COLOR: W.Global_INFLUENCE_COLOR
global INSTANCE_LOCK_LINK_COLOR: W.Global_INSTANCE_LOCK_LINK_COLOR
global INVALID_EQUIPMENT_COLOR: W.Global_INVALID_EQUIPMENT_COLOR
global INVASION_DESCRIPTION_FONT_COLOR: W.Global_INVASION_DESCRIPTION_FONT_COLOR
global INVASION_FONT_COLOR: W.Global_INVASION_FONT_COLOR
global ITEM_ARTIFACT_COLOR: W.Global_ITEM_ARTIFACT_COLOR
global ITEM_EPIC_COLOR: W.Global_ITEM_EPIC_COLOR
global ITEM_GOOD_COLOR: W.Global_ITEM_GOOD_COLOR
global ITEM_LEGENDARY_COLOR: W.Global_ITEM_LEGENDARY_COLOR
global ITEM_POOR_COLOR: W.Global_ITEM_POOR_COLOR
global ITEM_SCALING_STAT_COLOR: W.Global_ITEM_SCALING_STAT_COLOR
global ITEM_STANDARD_COLOR: W.Global_ITEM_STANDARD_COLOR
global ITEM_SUPERIOR_COLOR: W.Global_ITEM_SUPERIOR_COLOR
global ITEM_WOW_TOKEN_COLOR: W.Global_ITEM_WOW_TOKEN_COLOR
global InCinematic: function(): boolean
global InCombatLockdown: function(): boolean
global InGuildParty: function(): boolean, number, number, number
global InRepairMode: function(): number
global InboxItemCanDelete: function(number): boolean
global InitiateRolePoll: function(): boolean
global InitiateTrade: function(string)
global InvoiceFont_Med: W.Global_InvoiceFont_Med
global InvoiceFont_Small: W.Global_InvoiceFont_Small
global InvoiceTextFontNormal: W.Global_InvoiceTextFontNormal
global InvoiceTextFontSmall: W.Global_InvoiceTextFontSmall
global Is64BitClient: function(): boolean
global IsAccountSecured: function(): boolean
global IsAchievementEligible: function(number): boolean
global IsActiveBattlefieldArena: function(): boolean, boolean
global IsActiveQuestLegendary: function(number): boolean
global IsActiveQuestTrivial: function(unknown): unknown...
global IsAdvancedFlyableArea: function(): boolean
global IsAllowedToUserTeleport: function(): boolean
global IsAltKeyDown: function(): boolean
global IsArenaSkirmish: function(): unknown...
global IsArtifactCompletionHistoryAvailable: function(): boolean
global IsArtifactPowerItem: function(number | string): boolean
global IsAvailableQuestTrivial: function(unknown): unknown...
global IsBNLogin: function(): unknown...
global IsBetaBuild: function(): boolean
global IsBindingForGamePad: function(unknown): unknown...
global IsBreadcrumbQuest: function(unknown): unknown...
global IsCastingGlyph: function(): unknown...
global IsCemeterySelectionAvailable: function(): boolean
global IsCharacterNewlyBoosted: function(): boolean
global IsChatAFK: function(): unknown...
global IsChatChannelRaid: function(): unknown...
global IsChatDND: function(): unknown...
global IsCollisionEnabled: function(): boolean
global IsConsumableItem: function(number | string): boolean
global IsControlKeyDown: function(): boolean
global IsCorruptedItem: function(number | string): boolean | nil
global IsCosmeticItem: function(number | string): boolean | nil
global IsCpuBound: function(): boolean | nil
global IsCurrentItem: function(number | string): boolean
global IsCurrentQuestFailed: function(): unknown...
global IsDebugBuild: function(): boolean
global IsDemonHunterAvailable: function(): boolean
global IsDisplayChannelModerator: function(): unknown...
global IsDisplayChannelOwner: function(): unknown...
global IsDressableItem: function(number | string): boolean
global IsDrivableArea: function(): boolean
global IsDualWielding: function(): boolean
global IsEquippableItem: function(number | string): boolean
global IsEquippedItem: function(number | string): boolean
global IsEquippedItemType: function(string): boolean
global IsEuropeanNumbers: function(): boolean
global IsEveryoneAssistant: function(): unknown...
global IsExpansionTrial: function(): boolean
global IsFalling: function(? string): boolean
global IsFishingLoot: function(): boolean
global IsFlyableArea: function(): boolean
global IsFlying: function(? string): boolean
global IsGMClient: function(): boolean
global IsGUIDInGroup: function(string, ? number): boolean
global IsGamePadCursorControlEnabled: function(): unknown...
global IsGamePadFreelookEnabled: function(): unknown...
global IsGraphicsCVarValueSupported: function(): unknown...
global IsGraphicsSettingValueSupported: function(): unknown...
global IsGuildLeader: function(): boolean
global IsGuildMember: function(string): boolean
global IsGuildRankAssignmentAllowed: function(unknown, unknown): unknown...
global IsHarmfulItem: function(number | string): boolean
global IsHelpfulItem: function(number | string): boolean
global IsInActiveWorldPVP: function(): unknown...
global IsInAuthenticatedRank: function(): unknown...
global IsInCinematicScene: function(): boolean
global IsInGlobalEnvironment: function(): unknown...
global IsInGroup: function(? number): boolean
global IsInGuild: function(): boolean
global IsInGuildGroup: function(): boolean
global IsInInstance: function(): boolean, string
global IsInJailersTower: function(): boolean
global IsInLFGDungeon: function(): boolean
global IsInRaid: function(? number): boolean
global IsInScenarioGroup: function(): unknown...
global IsIndoors: function(): boolean
global IsInsane: function(): boolean
global IsInventoryItemLocked: function(number): boolean
global IsInventoryItemProfessionBag: function(unknown, unknown): unknown...
global IsItemInRange: function(number | string, string): boolean | nil
global IsItemPreferredArmorType: function(W.ItemLocation): boolean
global IsJailersTowerLayerTimeLocked: function(number): string
global IsKeyDown: function(string, ? boolean): boolean | nil
global IsLFGComplete: function(): boolean
global IsLFGDungeonJoinable: function(number): boolean, boolean, boolean, number
global IsLeftAltKeyDown: function(): boolean
global IsLeftControlKeyDown: function(): boolean
global IsLeftMetaKeyDown: function(): boolean
global IsLeftShiftKeyDown: function(): boolean
global IsLegacyDifficulty: function(number): boolean | nil
global IsLinuxClient: function(): boolean
global IsLoggedIn: function(): boolean
global IsMacClient: function(): boolean
global IsMasterLooter: function(): unknown...
global IsMetaKeyDown: function(): boolean
global IsModifiedClick: function(? string): boolean
global IsModifierKeyDown: function(): boolean
global IsMounted: function(): boolean
global IsMouseButtonDown: function(? W.Literal_b7e0de2fb994): boolean
global IsMouselooking: function(): boolean
global IsMovieLocal: function(number): boolean
global IsMoviePlayable: function(number): boolean
global IsMovieReadable: function(number): boolean
global IsOnGroundFloorInJailersTower: function(): boolean
global IsOnTournamentRealm: function(): boolean
global IsOutOfBounds: function(): boolean
global IsOutdoors: function(): boolean
global IsOutlineModeSupported: function(): unknown...
global IsPVPTimerRunning: function(): boolean
global IsPartyLFG: function(): unknown...
global IsPartyWorldPVP: function(): unknown...
global IsPendingGlyphRemoval: function(): unknown...
global IsPetActive: function(): unknown...
global IsPetAttackAction: function(unknown): unknown...
global IsPetAttackActive: function(): boolean
global IsPlayerInGuildFromGUID: function(string): boolean
global IsPlayerInWorld: function(): boolean
global IsPlayerMoving: function(): boolean
global IsPlayerNeutral: function(): boolean
global IsPlayerSpell: function(number): boolean
global IsPublicBuild: function(): boolean
global IsPublicTestClient: function(): boolean
global IsQuestCompletable: function(): boolean
global IsQuestIDValidSpellTarget: function(unknown): unknown...
global IsQuestItemHidden: function(unknown): unknown...
global IsQuestLogSpecialItemInRange: function(unknown, unknown): unknown...
global IsQuestSequenced: function(unknown): unknown...
global IsRaidMarkerActive: function(number): boolean | W.OpaqueValue
global IsRaidMarkerSystemEnabled: function(): boolean
global IsRangedWeapon: function(): boolean
global IsResting: function(): boolean
global IsRestrictedAccount: function(): boolean
global IsRightAltKeyDown: function(): boolean
global IsRightControlKeyDown: function(): boolean
global IsRightMetaKeyDown: function(): boolean
global IsRightShiftKeyDown: function(): boolean
global IsSelectedSpellBookItem: function(unknown): unknown...
global IsServerControlledBackfill: function(): unknown...
global IsShiftKeyDown: function(): boolean
global IsSpecializationActivateSpell: function(): unknown...
global IsSpellClassOrSpec: W.Global_IsSpellClassOrSpec
global IsSpellKnown: function(number, boolean): boolean
global IsSpellKnownOrOverridesKnown: function(number, boolean): boolean
global IsSpellOverlayed: function(number): boolean
global IsSpellValidForPendingGlyph: function(unknown): unknown...
global IsStealthed: function(): boolean
global IsStoryQuest: function(unknown): unknown...
global IsSubZonePVPPOI: function(): boolean
global IsSubmerged: function(? string): boolean
global IsSwimming: function(? string): boolean
global IsTargetLoose: function(): boolean
global IsTestBuild: function(): boolean
global IsThreatWarningEnabled: function(): boolean
global IsTitleKnown: function(number): boolean
global IsTradeskillTrainer: function(): boolean
global IsTrialAccount: function(): boolean
global IsTutorialFlagged: function(unknown): unknown...
global IsUnitModelReadyForUI: function(string): boolean
global IsUsableItem: function(number | string): boolean, boolean
global IsUsingFixedTimeStep: function(): boolean
global IsUsingGamepad: function(): boolean
global IsUsingMouse: function(): boolean
global IsUsingVehicleControls: function(): unknown...
global IsVehicleAimAngleAdjustable: function(): boolean
global IsVehicleAimPowerAdjustable: function(): boolean
global IsVeteranTrialAccount: function(): boolean
global IsWargame: function(): boolean
global IsWindowsClient: function(): boolean
global IsXPUserDisabled: function(): boolean
global Item: W.Global_Item
global ItemAddedToArtifact: function(number): boolean
global ItemCanTargetGarrisonFollowerAbility: function(): unknown...
global ItemHasRange: function(number | string): boolean
global ItemLocation: W.Global_ItemLocation
global ItemLocationMixin: W.Global_ItemLocationMixin
global ItemMixin: W.Global_ItemMixin
global ItemTextFontNormal: W.Global_ItemTextFontNormal
global ItemTextGetCreator: function(): string | nil
global ItemTextGetItem: function(): string
global ItemTextGetMaterial: function(): string
global ItemTextGetPage: function(): number
global ItemTextGetText: function(): string
global ItemTextHasNextPage: function(): boolean
global ItemTextIsFullPage: function(): unknown...
global ItemTextNextPage: function(): unknown...
global ItemTextPrevPage: function(): unknown...
global ItemTransmogInfoMixin: W.Global_ItemTransmogInfoMixin
global JOURNAL_LINK_COLOR: W.Global_JOURNAL_LINK_COLOR
global JoinArena: function(): unknown...
global JoinChannelByName: function(string, ? string, ? number, boolean): number, string | nil
global JoinLFG: function(unknown): unknown...
global JoinPermanentChannel: function(string, ? string, ? number, ? boolean): number, string
global JoinRatedBattlefield: function(): unknown...
global JoinRatedSoloShuffle: function(): unknown...
global JoinSingleLFG: function(unknown, unknown): unknown...
global JoinSkirmish: function(number, ? boolean): unknown...
global JoinTemporaryChannel: function(string, ? string, ? number, ? boolean): number, string
global JumpOrAscendStart: function(): unknown...
global KYRIAN_BLUE_COLOR: W.Global_KYRIAN_BLUE_COLOR
global LEGENDARY_ORANGE_COLOR: W.Global_LEGENDARY_ORANGE_COLOR
global LFGTeleport: function(boolean): unknown...
global LIGHTBLUE_FONT_COLOR: W.Global_LIGHTBLUE_FONT_COLOR
global LIGHTERBLUE_FONT_COLOR: W.Global_LIGHTERBLUE_FONT_COLOR
global LIGHTERGRAY_FONT_COLOR: W.Global_LIGHTERGRAY_FONT_COLOR
global LIGHTGRAY_FONT_COLOR: W.Global_LIGHTGRAY_FONT_COLOR
global LIGHTYELLOW_FONT_COLOR: W.Global_LIGHTYELLOW_FONT_COLOR
global LINK_FONT_COLOR: W.Global_LINK_FONT_COLOR
global LOOT_LINK_COLOR: W.Global_LOOT_LINK_COLOR
global LORE_TEXT_BODY_COLOR: W.Global_LORE_TEXT_BODY_COLOR
global LaunchURL: function(string)
global LearnPvpTalent: function(number, number): unknown...
global LearnPvpTalents: function(): unknown...
global LearnTalent: function(number): boolean
global LearnTalents: function(): unknown...
global LeaveBattlefield: function(): unknown...
global LeaveChannelByLocalID: function(unknown): unknown...
global LeaveChannelByName: function(string): unknown...
global LeaveLFG: function(unknown): unknown...
global LeaveSingleLFG: function(unknown, unknown): unknown...
global ListChannelByName: function(number | string): unknown...
global ListChannels: function(): unknown...
global LoadBindings: function(number): unknown...
global LoadURLIndex: function(number, ? number)
global LocalizedClassList: function(? boolean): unknown
global LoggingChat: function(? boolean): boolean | nil
global LoggingCombat: function(? boolean): boolean | nil
global Logout: function()
global LootMoneyNotify: function(unknown, unknown): unknown...
global LootSlot: function(number): unknown...
global LootSlotHasItem: function(number): boolean
global MARBLE_MATERIAL_TEXT_COLOR: W.Global_MARBLE_MATERIAL_TEXT_COLOR
global MARBLE_MATERIAL_TITLETEXT_COLOR: W.Global_MARBLE_MATERIAL_TITLETEXT_COLOR
global MailFont_Large: W.Global_MailFont_Large
global MailTextFontNormal: W.Global_MailTextFontNormal
global MakeModifiers: function(): number
global MapCanvasDataProviderMixin: W.Global_MapCanvasDataProviderMixin
global MapCanvasPinMixin: W.Global_MapCanvasPinMixin
global Menu: W.Global_Menu
global MenuUtil: W.Global_MenuUtil
global Minimap: W.Global_Minimap
global MinutesToSeconds: function(number): number
global MinutesToTime: function(number, ? boolean): string
global MissionCombatTextFontOutline: W.Global_MissionCombatTextFontOutline
global Mixin: function(unknown, unknown...): unknown
global MouseOverrideCinematicDisable: function(? boolean)
global MouselookStart: function(): unknown...
global MouselookStop: function(): unknown...
global MoveAndSteerStart: function(): unknown...
global MoveAndSteerStop: function(): unknown...
global MoveBackwardStart: function(number): unknown...
global MoveBackwardStop: function(): unknown...
global MoveForwardStart: function(number): unknown...
global MoveForwardStop: function(): unknown...
global MoveViewDownStart: function(number): unknown...
global MoveViewDownStop: function(): unknown...
global MoveViewInStart: function(number): unknown...
global MoveViewInStop: function(): unknown...
global MoveViewLeftStart: function(number): unknown...
global MoveViewLeftStop: function(): unknown...
global MoveViewOutStart: function(number): unknown...
global MoveViewOutStop: function(): unknown...
global MoveViewRightStart: function(number): unknown...
global MoveViewRightStop: function(): unknown...
global MoveViewUpStart: function(number): unknown...
global MoveViewUpStop: function(): unknown...
global MovieSubtitleFont: W.Global_MovieSubtitleFont
global MultiSampleAntiAliasingSupported: function(): unknown...
global MuteSoundFile: function(number | string): unknown...
global NECROLORD_GREEN_COLOR: W.Global_NECROLORD_GREEN_COLOR
global NEUTRAL_STATUS_COLOR: W.Global_NEUTRAL_STATUS_COLOR
global NEW_FEATURE_SHADOW_COLOR: W.Global_NEW_FEATURE_SHADOW_COLOR
global NIFFEN_MAJOR_FACTION_COLOR: W.Global_NIFFEN_MAJOR_FACTION_COLOR
global NIGHT_FAE_BLUE_COLOR: W.Global_NIGHT_FAE_BLUE_COLOR
global NORMAL_FONT_COLOR: W.Global_NORMAL_FONT_COLOR
global NOT_ON_THREAT_COLOR: W.Global_NOT_ON_THREAT_COLOR
global NO_THREAT_COLOR: W.Global_NO_THREAT_COLOR
global NeutralPlayerSelectFaction: function(number): unknown...
global NewSubSpellFont: W.Global_NewSubSpellFont
global NextView: function(): unknown...
global NoPlayTime: function(): boolean | nil
global NotWhileDeadError: function()
global NotifyInspect: function(string)
global NumTaxiNodes: function(): number
global Number11Font: W.Global_Number11Font
global Number11FontWhite: W.Global_Number11FontWhite
global Number12Font: W.Global_Number12Font
global Number12Font_o1: W.Global_Number12Font_o1
global Number13Font: W.Global_Number13Font
global Number13FontGray: W.Global_Number13FontGray
global Number13FontRed: W.Global_Number13FontRed
global Number13FontWhite: W.Global_Number13FontWhite
global Number13FontYellow: W.Global_Number13FontYellow
global Number14FontGray: W.Global_Number14FontGray
global Number14FontGreen: W.Global_Number14FontGreen
global Number14FontRed: W.Global_Number14FontRed
global Number14FontWhite: W.Global_Number14FontWhite
global Number15Font: W.Global_Number15Font
global Number15FontWhite: W.Global_Number15FontWhite
global Number16Font: W.Global_Number16Font
global Number18Font: W.Global_Number18Font
global Number18FontWhite: W.Global_Number18FontWhite
global NumberFontNormal: W.Global_NumberFontNormal
global NumberFontNormalGray: W.Global_NumberFontNormalGray
global NumberFontNormalHuge: W.Global_NumberFontNormalHuge
global NumberFontNormalLarge: W.Global_NumberFontNormalLarge
global NumberFontNormalLargeRight: W.Global_NumberFontNormalLargeRight
global NumberFontNormalLargeRightGray: W.Global_NumberFontNormalLargeRightGray
global NumberFontNormalLargeRightRed: W.Global_NumberFontNormalLargeRightRed
global NumberFontNormalLargeRightYellow: W.Global_NumberFontNormalLargeRightYellow
global NumberFontNormalLargeYellow: W.Global_NumberFontNormalLargeYellow
global NumberFontNormalRight: W.Global_NumberFontNormalRight
global NumberFontNormalRightGray: W.Global_NumberFontNormalRightGray
global NumberFontNormalRightGreen: W.Global_NumberFontNormalRightGreen
global NumberFontNormalRightRed: W.Global_NumberFontNormalRightRed
global NumberFontNormalRightYellow: W.Global_NumberFontNormalRightYellow
global NumberFontNormalSmall: W.Global_NumberFontNormalSmall
global NumberFontNormalSmallGray: W.Global_NumberFontNormalSmallGray
global NumberFontNormalYellow: W.Global_NumberFontNormalYellow
global NumberFontSmallBattleNetBlueLeft: W.Global_NumberFontSmallBattleNetBlueLeft
global NumberFontSmallWhiteLeft: W.Global_NumberFontSmallWhiteLeft
global NumberFontSmallYellowLeft: W.Global_NumberFontSmallYellowLeft
global NumberFont_GameNormal: W.Global_NumberFont_GameNormal
global NumberFont_Normal_Med: W.Global_NumberFont_Normal_Med
global NumberFont_OutlineThick_Mono_Small: W.Global_NumberFont_OutlineThick_Mono_Small
global NumberFont_Outline_Huge: W.Global_NumberFont_Outline_Huge
global NumberFont_Outline_Large: W.Global_NumberFont_Outline_Large
global NumberFont_Outline_Med: W.Global_NumberFont_Outline_Med
global NumberFont_Shadow_Large: W.Global_NumberFont_Shadow_Large
global NumberFont_Shadow_Med: W.Global_NumberFont_Shadow_Med
global NumberFont_Shadow_Small: W.Global_NumberFont_Shadow_Small
global NumberFont_Shadow_Tiny: W.Global_NumberFont_Shadow_Tiny
global NumberFont_Small: W.Global_NumberFont_Small
global OBJECTIVE_TRACKER_BLOCK_HEADER_COLOR: W.Global_OBJECTIVE_TRACKER_BLOCK_HEADER_COLOR
global ORANGE_FONT_COLOR: W.Global_ORANGE_FONT_COLOR
global ORANGE_THREAT_COLOR: W.Global_ORANGE_THREAT_COLOR
global ObjectiveFont: W.Global_ObjectiveFont
global OfferPetition: function(): unknown...
global OpenTrainer: function(): unknown...
global OpeningCinematic: function()
global OptionsFontHighlight: W.Global_OptionsFontHighlight
global OptionsFontHighlightSmall: W.Global_OptionsFontHighlightSmall
global OptionsFontLarge: W.Global_OptionsFontLarge
global OptionsFontLeft: W.Global_OptionsFontLeft
global OptionsFontSmall: W.Global_OptionsFontSmall
global OptionsFontSmallLeft: W.Global_OptionsFontSmallLeft
global OutageDetected: function()
global PANEL_BACKGROUND_COLOR: W.Global_PANEL_BACKGROUND_COLOR
global PAPER_FRAME_COLLAPSED_COLOR: W.Global_PAPER_FRAME_COLLAPSED_COLOR
global PAPER_FRAME_DARK_COLOR: W.Global_PAPER_FRAME_DARK_COLOR
global PAPER_FRAME_EXPANDED_COLOR: W.Global_PAPER_FRAME_EXPANDED_COLOR
global PAPER_FRAME_TEXT_COLOR: W.Global_PAPER_FRAME_TEXT_COLOR
global PAPER_FRAME_TITLE_COLOR: W.Global_PAPER_FRAME_TITLE_COLOR
global PARCHMENTLARGE_MATERIAL_TEXT_COLOR: W.Global_PARCHMENTLARGE_MATERIAL_TEXT_COLOR
global PARCHMENTLARGE_MATERIAL_TITLETEXT_COLOR: W.Global_PARCHMENTLARGE_MATERIAL_TITLETEXT_COLOR
global PARCHMENT_MATERIAL_TEXT_COLOR: W.Global_PARCHMENT_MATERIAL_TEXT_COLOR
global PARCHMENT_MATERIAL_TITLETEXT_COLOR: W.Global_PARCHMENT_MATERIAL_TITLETEXT_COLOR
global PASSIVE_SPELL_FONT_COLOR: W.Global_PASSIVE_SPELL_FONT_COLOR
global PLAYER_FACTION_COLOR_ALLIANCE: W.Global_PLAYER_FACTION_COLOR_ALLIANCE
global PLAYER_FACTION_COLOR_HORDE: W.Global_PLAYER_FACTION_COLOR_HORDE
global PLAYER_NAMEPLATE_ALTERNATE_HEALTH_COLOR: W.Global_PLAYER_NAMEPLATE_ALTERNATE_HEALTH_COLOR
global PLUNDERSTORM_WANTED_COLOR: W.Global_PLUNDERSTORM_WANTED_COLOR
global POWERBAR_PREDICTION_COLOR_ENERGY: W.Global_POWERBAR_PREDICTION_COLOR_ENERGY
global POWERBAR_PREDICTION_COLOR_FOCUS: W.Global_POWERBAR_PREDICTION_COLOR_FOCUS
global POWERBAR_PREDICTION_COLOR_FURY: W.Global_POWERBAR_PREDICTION_COLOR_FURY
global POWERBAR_PREDICTION_COLOR_INSANITY: W.Global_POWERBAR_PREDICTION_COLOR_INSANITY
global POWERBAR_PREDICTION_COLOR_LUNAR_POWER: W.Global_POWERBAR_PREDICTION_COLOR_LUNAR_POWER
global POWERBAR_PREDICTION_COLOR_MAELSTROM: W.Global_POWERBAR_PREDICTION_COLOR_MAELSTROM
global POWERBAR_PREDICTION_COLOR_MANA: W.Global_POWERBAR_PREDICTION_COLOR_MANA
global POWERBAR_PREDICTION_COLOR_PAIN: W.Global_POWERBAR_PREDICTION_COLOR_PAIN
global POWERBAR_PREDICTION_COLOR_RAGE: W.Global_POWERBAR_PREDICTION_COLOR_RAGE
global POWERBAR_PREDICTION_COLOR_RUNIC_POWER: W.Global_POWERBAR_PREDICTION_COLOR_RUNIC_POWER
global PROFESSIONS_JEWELCRAFTING_GEMSTONES_AMBER: W.Global_PROFESSIONS_JEWELCRAFTING_GEMSTONES_AMBER
global PROFESSIONS_JEWELCRAFTING_GEMSTONES_CITRINE: W.Global_PROFESSIONS_JEWELCRAFTING_GEMSTONES_CITRINE
global PROFESSIONS_JEWELCRAFTING_GEMSTONES_EMERALD: W.Global_PROFESSIONS_JEWELCRAFTING_GEMSTONES_EMERALD
global PROFESSIONS_JEWELCRAFTING_GEMSTONES_ONYX: W.Global_PROFESSIONS_JEWELCRAFTING_GEMSTONES_ONYX
global PROFESSIONS_JEWELCRAFTING_GEMSTONES_RUBY: W.Global_PROFESSIONS_JEWELCRAFTING_GEMSTONES_RUBY
global PROFESSIONS_JEWELCRAFTING_GEMSTONES_SAPPHIRE: W.Global_PROFESSIONS_JEWELCRAFTING_GEMSTONES_SAPPHIRE
global PROFESSION_RECIPE_COLOR: W.Global_PROFESSION_RECIPE_COLOR
global PROFESSION_SPEC_COLOR_ALCHEMY: W.Global_PROFESSION_SPEC_COLOR_ALCHEMY
global PROFESSION_SPEC_COLOR_ARCHAEOLOGY: W.Global_PROFESSION_SPEC_COLOR_ARCHAEOLOGY
global PROFESSION_SPEC_COLOR_BLACKSMITHING: W.Global_PROFESSION_SPEC_COLOR_BLACKSMITHING
global PROFESSION_SPEC_COLOR_COOKING: W.Global_PROFESSION_SPEC_COLOR_COOKING
global PROFESSION_SPEC_COLOR_ENCHANTING: W.Global_PROFESSION_SPEC_COLOR_ENCHANTING
global PROFESSION_SPEC_COLOR_ENGINEERING: W.Global_PROFESSION_SPEC_COLOR_ENGINEERING
global PROFESSION_SPEC_COLOR_FIRST_AID: W.Global_PROFESSION_SPEC_COLOR_FIRST_AID
global PROFESSION_SPEC_COLOR_FISHING: W.Global_PROFESSION_SPEC_COLOR_FISHING
global PROFESSION_SPEC_COLOR_HERBALISM: W.Global_PROFESSION_SPEC_COLOR_HERBALISM
global PROFESSION_SPEC_COLOR_INSCRIPTION: W.Global_PROFESSION_SPEC_COLOR_INSCRIPTION
global PROFESSION_SPEC_COLOR_JEWELCRAFTING: W.Global_PROFESSION_SPEC_COLOR_JEWELCRAFTING
global PROFESSION_SPEC_COLOR_LEATHERWORKING: W.Global_PROFESSION_SPEC_COLOR_LEATHERWORKING
global PROFESSION_SPEC_COLOR_MINING: W.Global_PROFESSION_SPEC_COLOR_MINING
global PROFESSION_SPEC_COLOR_SKINNING: W.Global_PROFESSION_SPEC_COLOR_SKINNING
global PROFESSION_SPEC_COLOR_TAILORING: W.Global_PROFESSION_SPEC_COLOR_TAILORING
global PROGENITOR_MATERIAL_TEXT_COLOR: W.Global_PROGENITOR_MATERIAL_TEXT_COLOR
global PROGENITOR_MATERIAL_TITLETEXT_COLOR: W.Global_PROGENITOR_MATERIAL_TITLETEXT_COLOR
global PURE_BLUE_COLOR: W.Global_PURE_BLUE_COLOR
global PURE_GREEN_COLOR: W.Global_PURE_GREEN_COLOR
global PURE_RED_COLOR: W.Global_PURE_RED_COLOR
global PVPInfoTextFont: W.Global_PVPInfoTextFont
global PVP_SCOREBOARD_ALLIANCE_ALT_CELL_COLOR: W.Global_PVP_SCOREBOARD_ALLIANCE_ALT_CELL_COLOR
global PVP_SCOREBOARD_ALLIANCE_ALT_ROW_COLOR: W.Global_PVP_SCOREBOARD_ALLIANCE_ALT_ROW_COLOR
global PVP_SCOREBOARD_ALLIANCE_CELL_COLOR: W.Global_PVP_SCOREBOARD_ALLIANCE_CELL_COLOR
global PVP_SCOREBOARD_ALLIANCE_ROW_COLOR: W.Global_PVP_SCOREBOARD_ALLIANCE_ROW_COLOR
global PVP_SCOREBOARD_HORDE_ALT_CELL_COLOR: W.Global_PVP_SCOREBOARD_HORDE_ALT_CELL_COLOR
global PVP_SCOREBOARD_HORDE_ALT_ROW_COLOR: W.Global_PVP_SCOREBOARD_HORDE_ALT_ROW_COLOR
global PVP_SCOREBOARD_HORDE_CELL_COLOR: W.Global_PVP_SCOREBOARD_HORDE_CELL_COLOR
global PVP_SCOREBOARD_HORDE_ROW_COLOR: W.Global_PVP_SCOREBOARD_HORDE_ROW_COLOR
global PartialPlayTime: function(): boolean | nil
global PartyLFGStartBackfill: function(): unknown...
global PetAggressiveMode: function(): unknown...
global PetAttack: function(): unknown...
global PetCanBeAbandoned: function(): boolean
global PetCanBeDismissed: function(): unknown...
global PetDefensiveAssistMode: function(): unknown...
global PetDefensiveMode: function(): unknown...
global PetDismiss: function(): unknown...
global PetFollow: function(): unknown...
global PetHasActionBar: function(): boolean
global PetHasSpellbook: function(): unknown...
global PetMoveTo: function(unknown): unknown...
global PetPassiveMode: function(): unknown...
global PetStopAttack: function(): unknown...
global PetUsesPetFrame: function(): unknown...
global PetWait: function(): unknown...
global PickupAction: function(number): unknown...
global PickupBagFromSlot: function(number): unknown...
global PickupCompanion: function(string, number): unknown...
global PickupGuildBankItem: function(unknown, unknown): unknown...
global PickupGuildBankMoney: function(unknown): unknown...
global PickupInventoryItem: function(number): unknown...
global PickupItem: function(number | string)
global PickupMacro: function(number | string): unknown...
global PickupMerchantItem: function(number): unknown...
global PickupPetAction: function(number): unknown...
global PickupPetSpell: function(number): unknown...
global PickupPlayerMoney: function(number)
global PickupPvpTalent: function(): unknown...
global PickupTalent: function(unknown): unknown...
global PitchDownStart: function(): unknown...
global PitchDownStop: function(): unknown...
global PitchUpStart: function(): unknown...
global PitchUpStop: function(): unknown...
global PlaceAction: function(number): unknown...
global PlaceRaidMarker: function(number, ? string)
global PlayAutoAcceptQuestSound: function(): unknown...
global PlayMusic: function(number | string): boolean
global PlaySound: function(number, ? string, ? boolean, ? boolean): boolean, number
global PlaySoundFile: function(number | string, ? string): boolean, number
global PlayVocalErrorSoundID: function(number)
global PlayerCanTeleport: function(): boolean
global PlayerEffectiveAttackPower: function(): W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil
global PlayerGetTimerunningSeasonID: function(): number | nil
global PlayerHasToy: function(number): boolean
global PlayerIsInCombat: function(): boolean
global PlayerIsPVPInactive: function(string): boolean
global PlayerIsSpellTarget: function(string): boolean | W.OpaqueValue
global PlayerIsTimerunning: function(): boolean
global PlayerLocation: W.Global_PlayerLocation
global PlayerLocationMixin: W.Global_PlayerLocationMixin
global PlayerVehicleHasComboPoints: function(): boolean
global PortGraveyard: function()
global PreloadMovie: function(number)
global PrevView: function(): unknown...
global PriceFont: W.Global_PriceFont
global PriceFontGray: W.Global_PriceFontGray
global PriceFontGreen: W.Global_PriceFontGreen
global PriceFontRed: W.Global_PriceFontRed
global PriceFontWhite: W.Global_PriceFontWhite
global PriceFontYellow: W.Global_PriceFontYellow
global ProcessExceptionClient: function(string): unknown...
global ProcessQuestLogRewardFactions: function(): unknown...
global PromoteToAssistant: function(unknown): unknown...
global PromoteToLeader: W.Global_PromoteToLeader
global PutItemInBackpack: function(): unknown...
global PutItemInBag: function(? number): boolean | nil
global QUEST_ACCOUNT_COMPLETED_NOTICE_FONT_COLOR: W.Global_QUEST_ACCOUNT_COMPLETED_NOTICE_FONT_COLOR
global QUEST_ACTIVE_TRIVIAL_GRAY: W.Global_QUEST_ACTIVE_TRIVIAL_GRAY
global QUEST_INELIGIBLE_GRAY: W.Global_QUEST_INELIGIBLE_GRAY
global QUEST_OBJECTIVE_COMPLETED_FONT_COLOR: W.Global_QUEST_OBJECTIVE_COMPLETED_FONT_COLOR
global QUEST_OBJECTIVE_COMPLETED_FONT_COLOR_DARK_BACKGROUND: W.Global_QUEST_OBJECTIVE_COMPLETED_FONT_COLOR_DARK_BACKGROUND
global QUEST_OBJECTIVE_DISABLED_FONT_COLOR: W.Global_QUEST_OBJECTIVE_DISABLED_FONT_COLOR
global QUEST_OBJECTIVE_DISABLED_HIGHLIGHT_FONT_COLOR: W.Global_QUEST_OBJECTIVE_DISABLED_HIGHLIGHT_FONT_COLOR
global QUEST_OBJECTIVE_FONT_COLOR: W.Global_QUEST_OBJECTIVE_FONT_COLOR
global QUEST_OBJECTIVE_HIGHLIGHT_FONT_COLOR: W.Global_QUEST_OBJECTIVE_HIGHLIGHT_FONT_COLOR
global QUEST_REWARD_CONTEXT_FONT_COLOR: W.Global_QUEST_REWARD_CONTEXT_FONT_COLOR
global QueryGuildBankLog: function(unknown): unknown...
global QueryGuildBankTab: function(unknown): unknown...
global QueryGuildBankText: function(): unknown...
global QueryGuildEventLog: function(): unknown...
global QueryGuildNews: function(): unknown...
global QueryGuildRecipes: function(): unknown...
global QuestChooseRewardError: function(): unknown...
global QuestDifficulty_Difficult: W.Global_QuestDifficulty_Difficult
global QuestDifficulty_Header: W.Global_QuestDifficulty_Header
global QuestDifficulty_Impossible: W.Global_QuestDifficulty_Impossible
global QuestDifficulty_Standard: W.Global_QuestDifficulty_Standard
global QuestDifficulty_Trivial: W.Global_QuestDifficulty_Trivial
global QuestDifficulty_VeryDifficult: W.Global_QuestDifficulty_VeryDifficult
global QuestFlagsPVP: function(): unknown...
global QuestFont: W.Global_QuestFont
global QuestFontHighlightHuge: W.Global_QuestFontHighlightHuge
global QuestFontLeft: W.Global_QuestFontLeft
global QuestFontNormalHuge: W.Global_QuestFontNormalHuge
global QuestFontNormalLarge: W.Global_QuestFontNormalLarge
global QuestFontNormalSmall: W.Global_QuestFontNormalSmall
global QuestFont_30: W.Global_QuestFont_30
global QuestFont_39: W.Global_QuestFont_39
global QuestFont_Enormous: W.Global_QuestFont_Enormous
global QuestFont_Huge: W.Global_QuestFont_Huge
global QuestFont_Large: W.Global_QuestFont_Large
global QuestFont_Outline_Huge: W.Global_QuestFont_Outline_Huge
global QuestFont_Shadow_Enormous: W.Global_QuestFont_Shadow_Enormous
global QuestFont_Shadow_Huge: W.Global_QuestFont_Shadow_Huge
global QuestFont_Shadow_Small: W.Global_QuestFont_Shadow_Small
global QuestFont_Shadow_Super_Huge: W.Global_QuestFont_Shadow_Super_Huge
global QuestFont_Super_Huge: W.Global_QuestFont_Super_Huge
global QuestFont_Super_Huge_Outline: W.Global_QuestFont_Super_Huge_Outline
global QuestGetAutoAccept: function(): boolean
global QuestGetAutoLaunched: function(): unknown...
global QuestHasPOIInfo: function(unknown): unknown...
global QuestIsDaily: function(): boolean
global QuestIsFromAdventureMap: function(): unknown...
global QuestIsFromAreaTrigger: function(): boolean
global QuestIsWeekly: function(): boolean
global QuestLogPushQuest: function(): unknown...
global QuestLogRewardHasTreasurePicker: function(): unknown...
global QuestLogShouldShowPortrait: function(): unknown...
global QuestMapRewardsFont: W.Global_QuestMapRewardsFont
global QuestMapUpdateAllQuests: function(): unknown...
global QuestPOIUpdateIcons: function(): unknown...
global QuestTitleFont: W.Global_QuestTitleFont
global QuestTitleFontBlackShadow: W.Global_QuestTitleFontBlackShadow
global Quit: function()
global RAF_VERSION_THREE_COLOR: W.Global_RAF_VERSION_THREE_COLOR
global RAF_VERSION_TWO_COLOR: W.Global_RAF_VERSION_TWO_COLOR
global RARE_BLUE_COLOR: W.Global_RARE_BLUE_COLOR
global RARE_MISSION_COLOR: W.Global_RARE_MISSION_COLOR
global RED_FONT_COLOR: W.Global_RED_FONT_COLOR
global RED_THREAT_COLOR: W.Global_RED_THREAT_COLOR
global REFUND_TOOLTIP_COLOR: W.Global_REFUND_TOOLTIP_COLOR
global RPE_FONT_COLOR: W.Global_RPE_FONT_COLOR
global RUNEFORGE_LEGEDARY_SPEC_COLOR: W.Global_RUNEFORGE_LEGEDARY_SPEC_COLOR
global RaidNotice_AddMessage: function(W.Frame, string, W.ChatTypeColor, ? number)
global RaidProfileExists: function(unknown): unknown...
global RaidProfileHasUnsavedChanges: function(): unknown...
global RaidWarningFrame: W.Global_RaidWarningFrame
global RandomRoll: function(number, number)
global RedockChatWindows: function(): unknown...
global RefreshLFGList: function(): unknown...
global RegisterCVar: function(string, ? string)
global RegisterEventCallback: function(string, function())
global RegisterStateDriver: function(W.Frame, string, string)
global RegisterStaticConstants: function(unknown): unknown...
global RegisterUnitEventCallback: function(string, function(), string)
global RejectProposal: function(): unknown...
global ReleaseAction: function(): unknown...
global RemoveAutoQuestPopUp: function(): unknown...
global RemoveChatWindowChannel: function(number, string): unknown...
global RemoveChatWindowMessages: function(number, string): unknown...
global RemoveItemFromArtifact: function(): boolean
global RemovePvpTalent: function(): unknown...
global RemoveRaidTargets: function()
global RemoveTalent: function(unknown): unknown...
global RenamePetition: function(string): unknown...
global RepairAllItems: function(? boolean): unknown...
global ReplaceEnchant: function()
global ReplaceGenderTokens: function(string, string): string
global ReplaceGuildMaster: function(): unknown...
global ReplaceTradeEnchant: function()
global ReplaceTradeskillEnchant: function()
global RepopMe: function()
global ReportBug: function(string)
global ReportInfoMixin: W.Global_ReportInfoMixin
global ReportPlayerIsPVPAFK: function(string)
global ReportSuggestion: function(string)
global ReputationDetailFont: W.Global_ReputationDetailFont
global RequestBattlefieldScoreData: function(): unknown...
global RequestBattlegroundInstanceInfo: function(number): unknown...
global RequestBottomLeftActionBar: function(): unknown...
global RequestGuildChallengeInfo: function(): unknown...
global RequestGuildPartyState: function(): unknown...
global RequestGuildRewards: function(): unknown...
global RequestLFDPartyLockInfo: function(): unknown...
global RequestLFDPlayerLockInfo: function(): unknown...
global RequestPVPOptionsEnabled: function(): unknown...
global RequestPVPRewards: function(): unknown...
global RequestRaidInfo: function(): unknown...
global RequestRandomBattlegroundInstanceInfo: function(): unknown...
global RequestRatedInfo: function(): unknown...
global RequestTimePlayed: function()
global RequeueSkirmish: function(): unknown...
global ResetCPUUsage: function()
global ResetChatColors: function(): unknown...
global ResetChatWindows: function(): unknown...
global ResetCursor: function()
global ResetInstances: function()
global ResetSetMerchantFilter: function(): unknown...
global ResetTestCvars: function()
global ResetTutorials: function(): unknown...
global ResetView: function(unknown): unknown...
global ResistancePercent: function(number, number): number
global RespondInstanceLock: function(boolean)
global RespondMailLockSendItem: function(unknown, unknown): unknown...
global RespondToInviteConfirmation: function(unknown, unknown): unknown...
global RestartGx: function()
global RestoreRaidProfileFromCopy: function(): unknown...
global ResurrectGetOfferer: function(): string
global ResurrectHasSickness: function(): boolean
global ResurrectHasTimer: function(): boolean
global RetrieveCorpse: function()
global ReturnInboxItem: function(unknown): unknown...
global RollOnLoot: function(number, ? number): unknown...
global RunBinding: function(string, ? string): unknown...
global RunMacro: function(number | string): unknown...
global RunNextFrame: function(W.Callback_78161532ce55)
global RunScript: function(string)
global SCENARIO_STAGE_COLOR: W.Global_SCENARIO_STAGE_COLOR
global SCENARIO_SUBTITLE_COLOR: W.Global_SCENARIO_SUBTITLE_COLOR
global SEPIA_COLOR: W.Global_SEPIA_COLOR
global SET_ITEM_COLOR: W.Global_SET_ITEM_COLOR
global SILVER_MATERIAL_TEXT_COLOR: W.Global_SILVER_MATERIAL_TEXT_COLOR
global SILVER_MATERIAL_TITLETEXT_COLOR: W.Global_SILVER_MATERIAL_TITLETEXT_COLOR
global SOULBIND_CONDUIT_ENHANCED_COLOR: W.Global_SOULBIND_CONDUIT_ENHANCED_COLOR
global SOUNDKIT: W.Global_SOUNDKIT
global SPELLBOOK_FONT_COLOR: W.Global_SPELLBOOK_FONT_COLOR
global SPELLBOOK_UNLEARNED_TINT_COLOR: W.Global_SPELLBOOK_UNLEARNED_TINT_COLOR
global SPELL_LINK_COLOR: W.Global_SPELL_LINK_COLOR
global SPELL_SUBTEXT_COLOR: W.Global_SPELL_SUBTEXT_COLOR
global STONE_MATERIAL_TEXT_COLOR: W.Global_STONE_MATERIAL_TEXT_COLOR
global STONE_MATERIAL_TITLETEXT_COLOR: W.Global_STONE_MATERIAL_TITLETEXT_COLOR
global STORM_MAJOR_FACTION_COLOR: W.Global_STORM_MAJOR_FACTION_COLOR
global SUBSCRIPTION_INTERSTITIAL_COLOR: W.Global_SUBSCRIPTION_INTERSTITIAL_COLOR
global SafePack: function(unknown...): {unknown: unknown}
global SafeUnpack: function({unknown: unknown}): unknown...
global SaveBindings: function(number): unknown...
global SaveRaidProfileCopy: function(unknown): unknown...
global SaveView: function(number): unknown...
global Screenshot: function()
global SearchLFGGetEncounterResults: function(unknown, unknown): unknown...
global SearchLFGGetJoinedID: function(): unknown...
global SearchLFGGetNumResults: function(): number, number
global SearchLFGGetPartyResults: function(number, number): unknown...
global SearchLFGGetResults: function(number): unknown...
global SearchLFGJoin: function(number, number): unknown...
global SearchLFGLeave: function(): unknown...
global SearchLFGSort: function(unknown): unknown...
global SecondsToClock: function(number, ? boolean): string
global SecondsToMinutes: function(number): number
global SecondsToTime: function(number, ? boolean, ? boolean, ? number, ? boolean): string
global SecondsToTimeAbbrev: function(number): string, number
global SecureCmdOptionParse: function(string): string, string
global SecureHandlerSetFrameRef: function(W.Frame, string, W.Frame)
global SecureHandlerWrapScript: function(W.Frame, string, W.Frame, string, ? string)
global SelectActiveQuest: function(): unknown...
global SelectAvailableQuest: function(): unknown...
global SelectTrainerService: function(number): unknown...
global SelectedRealmName: function(): string
global SellCursorItem: function()
global SendChatMessage: function(string, ? W.Literal_62300f6ce66d, ? number, ? string)
global SendMail: function(string, string, ? string): unknown...
global SendSubscriptionInterstitialResponse: function(number)
global SendSystemMessage: function(string): unknown...
global SetAchievementComparisonUnit: function(string): boolean
global SetAchievementSearchString: function(string): boolean
global SetActionBarToggles: function(boolean, boolean, boolean, boolean, boolean, boolean, boolean, string): unknown...
global SetAllowDangerousScripts: function(? boolean)
global SetAllowLowLevelRaid: function(? boolean)
global SetAllowRecentAlliesSeeLocation: function(boolean)
global SetAutoDeclineGuildInvites: function(? boolean)
global SetAutoDeclineNeighborhoodInvites: function(? boolean)
global SetBarSlotFromIntro: function(unknown): unknown...
global SetBattlefieldScoreFaction: function(number): unknown...
global SetBinding: function(string, ? string, ? number): boolean
global SetBindingClick: function(string, string, string): boolean
global SetBindingItem: function(string, string): boolean
global SetBindingMacro: function(string, number | string): boolean
global SetBindingSpell: function(string, string): boolean
global SetCVar: function(string, ? boolean | string | number): boolean
global SetCVarBitfield: function(string, number, boolean): boolean
global SetCemeteryPreference: function(number)
global SetChannelOwner: function(string, string): unknown...
global SetChannelPassword: function(string, string): unknown...
global SetChatColorNameByClass: function(unknown, unknown): unknown...
global SetChatWindowAlpha: function(unknown, unknown): unknown...
global SetChatWindowColor: function(unknown, unknown, unknown, unknown): unknown...
global SetChatWindowDocked: function(unknown, unknown): unknown...
global SetChatWindowLocked: function(unknown, unknown): unknown...
global SetChatWindowName: function(unknown, unknown): unknown...
global SetChatWindowSavedDimensions: function(unknown, unknown, unknown): unknown...
global SetChatWindowSavedPosition: function(unknown, unknown, unknown, unknown): unknown...
global SetChatWindowShown: function(unknown, unknown): unknown...
global SetChatWindowSize: function(unknown, unknown): unknown...
global SetChatWindowUninteractable: function(unknown, unknown): unknown...
global SetClampedTextureRotation: function(W.Texture, number)
global SetConsoleKey: function(string)
global SetCurrentGraphicsSetting: function(): unknown...
global SetCurrentGuildBankTab: function(unknown): unknown...
global SetCurrentTitle: function(number)
global SetCursor: function(? string): boolean
global SetCursorByMode: function(number): boolean
global SetCursorHoveredItem: function(W.ItemLocation)
global SetCursorHoveredItemTradeItem: function(boolean)
global SetCursorPosition: function(number, number)
global SetCursorVirtualItem: function(number | string, number)
global SetDungeonDifficultyID: function(number)
global SetErrorCallstackHeight: function(? number)
global SetEuropeanNumbers: function(boolean)
global SetEveryoneIsAssistant: function(): unknown...
global SetFocusedAchievement: function(? number): unknown...
global SetGamePadCursorControl: function(): unknown...
global SetGamePadFreeLook: function(): unknown...
global SetGuildBankTabInfo: function(number, string, number): unknown...
global SetGuildBankTabItemWithdraw: function(): unknown...
global SetGuildBankTabPermissions: function(number, number, boolean): unknown...
global SetGuildBankText: function(number, string): unknown...
global SetGuildBankWithdrawGoldLimit: function(number): unknown...
global SetGuildMemberRank: function(unknown, unknown): unknown...
global SetGuildNewsFilter: function(unknown, unknown): unknown...
global SetGuildRosterSelection: function(number): unknown...
global SetGuildRosterShowOffline: function(boolean): unknown...
global SetGuildTradeSkillCategoryFilter: function(unknown): unknown...
global SetGuildTradeSkillItemNameFilter: function(unknown): unknown...
global SetInWorldUIVisibility: function(boolean)
global SetLFGBootVote: function(unknown): unknown...
global SetLFGComment: function(string): unknown...
global SetLFGDungeon: function(unknown, unknown): unknown...
global SetLFGDungeonEnabled: function(unknown, unknown): unknown...
global SetLFGHeaderCollapsed: function(unknown, unknown): unknown...
global SetLFGRoles: function(unknown, unknown, unknown, unknown): unknown...
global SetLegacyRaidDifficultyID: function(number, ? boolean)
global SetLootPortrait: function(): unknown...
global SetLootSpecialization: function(number)
global SetLootThreshold: function(number): unknown...
global SetMacroItem: function(unknown, unknown, unknown): unknown...
global SetMacroSpell: function(number | string, string, string): unknown...
global SetMerchantFilter: function(): unknown...
global SetModifiedClick: function(string, string): unknown...
global SetMouselookOverrideBinding: function(unknown, unknown): unknown...
global SetMoveEnabled: function(): unknown...
global SetMultiCastSpell: function(number, number): unknown...
global SetOptOutOfLoot: function(boolean): unknown...
global SetOverrideBinding: function(W.Frame, boolean, string, string): unknown...
global SetOverrideBindingClick: function(W.Frame, boolean, string, string, ? string): unknown...
global SetOverrideBindingItem: function(W.Frame, boolean, string, string): unknown...
global SetOverrideBindingMacro: function(W.Frame, boolean, string, string): unknown...
global SetOverrideBindingSpell: function(W.Frame, boolean, string, string): unknown...
global SetPOIIconOverlapDistance: function(unknown): unknown...
global SetPOIIconOverlapPushDistance: function(unknown): unknown...
global SetPVP: function(? boolean)
global SetPVPRoles: function(boolean, boolean, boolean): unknown...
global SetPartyAssignment: function(unknown, unknown): unknown...
global SetPortraitTexture: function(W.Texture, string, ? boolean)
global SetPortraitTextureFromCreatureDisplayID: function(W.Texture, number)
global SetPortraitToTexture: function(W.Texture, number | string)
global SetRaidDifficultyID: function(number, ? boolean)
global SetRaidProfileOption: function(unknown, unknown, unknown): unknown...
global SetRaidProfileSavedPosition: function(unknown, unknown, unknown, unknown, unknown, unknown, unknown, unknown): unknown...
global SetRaidSubgroup: function(number, number): unknown...
global SetRaidTarget: function(string, number)
global SetSavedInstanceExtend: function(unknown, unknown): unknown...
global SetSelectedArtifact: function(number, ? number): unknown...
global SetSelectedDisplayChannel: function(unknown): unknown...
global SetSelectedScreenResolutionIndex: function(unknown): unknown...
global SetSelectedWarGameType: function(unknown): unknown...
global SetSendMailCOD: function(unknown): unknown...
global SetSendMailMoney: function(unknown): unknown...
global SetSendMailShowing: function(boolean): unknown...
global SetSpellbookPetAction: function(unknown, unknown): unknown...
global SetTableSecurityOption: function(unknown, number)
global SetTaxiBenchmarkMode: function(? boolean)
global SetTaxiMap: function(string): unknown...
global SetTradeCurrency: function(unknown, unknown): unknown...
global SetTrainerServiceTypeFilter: function(string, boolean, boolean): unknown...
global SetTurnEnabled: function(): unknown...
global SetUIVisibility: function(boolean)
global SetUnitCursorTexture: function(W.Texture, string, ? number, ? boolean): boolean
global SetView: function(number): unknown...
global Settings: W.Global_Settings
global SetupAtlasesOnRegions: function(W.Frame, {unknown: unknown}, ? boolean)
global SetupFullscreenScale: function(W.Frame): unknown...
global SetupTextureKitOnFrame: W.Global_SetupTextureKitOnFrame
global SetupTextureKitOnFrames: function(string | {unknown: unknown}, {unknown: unknown}, boolean, ? boolean)
global SetupTextureKitOnRegions: function(string | {unknown: unknown}, W.Frame, {unknown: unknown}, boolean, ? boolean)
global SetupTextureKitsFromRegionInfo: function(string | {unknown: unknown}, W.Frame, {unknown: unknown})
global ShouldShowExpansionUpgradeBanner: function(): boolean
global ShouldShowIslandsWeeklyPOI: function(): boolean
global ShouldShowSpecialSplashScreen: function(): boolean
global ShowBossFrameWhenUninteractable: function(? string): boolean
global ShowBuybackSellCursor: function(unknown): unknown...
global ShowCloak: function(boolean)
global ShowHelm: function(boolean)
global ShowQuestComplete: function(number): unknown...
global ShowQuestOffer: function(unknown): unknown...
global ShowRepairCursor: function(): unknown...
global ShowUIPanel: function(W.Frame, ? number)
global ShowingCloak: function(): boolean
global ShowingHelm: function(): boolean
global SignPetition: function(): unknown...
global SimulateMouseClick: function(W.Literal_b7e0de2fb994)
global SimulateMouseDown: function(W.Literal_b7e0de2fb994)
global SimulateMouseUp: function(W.Literal_b7e0de2fb994)
global SimulateMouseWheel: function(number)
global SitStandOrDescendStart: function()
global SlashCmdList: W.Global_SlashCmdList
global SocketInventoryItem: function(unknown): unknown...
global SocketItemToArtifact: function(): boolean
global SolveArtifact: function(): unknown...
global SortBGList: function(): boolean
global SortBattlefieldScoreData: function(unknown): unknown...
global SortGuildRoster: function(string): unknown...
global SortGuildTradeSkill: function(unknown): unknown...
global SortQuestSortTypes: function(): unknown...
global SortQuests: function(): unknown...
global Sound_ChatSystem_GetInputDriverNameByIndex: function(unknown): unknown...
global Sound_ChatSystem_GetNumInputDrivers: function(): unknown...
global Sound_ChatSystem_GetNumOutputDrivers: function(): unknown...
global Sound_ChatSystem_GetOutputDriverNameByIndex: function(unknown): unknown...
global Sound_GameSystem_GetInputDriverNameByIndex: function(unknown): unknown...
global Sound_GameSystem_GetNumInputDrivers: function(): unknown...
global Sound_GameSystem_GetNumOutputDrivers: function(): unknown...
global Sound_GameSystem_GetOutputDriverNameByIndex: function(unknown): unknown...
global Sound_GameSystem_RestartSoundSystem: function(): unknown...
global Spell: W.Global_Spell
global SpellCanTargetGarrisonFollower: function(): unknown...
global SpellCanTargetGarrisonFollowerAbility: function(): unknown...
global SpellCanTargetGarrisonMission: function(): unknown...
global SpellCanTargetItem: function(): unknown...
global SpellCanTargetItemID: function(): unknown...
global SpellCanTargetQuest: function(): unknown...
global SpellCanTargetUnit: function(string): boolean
global SpellCancelQueuedSpell: function(): unknown...
global SpellFont_Small: W.Global_SpellFont_Small
global SpellIsTargeting: function(): boolean
global SpellMixin: W.Global_SpellMixin
global SpellStopCasting: function(): boolean
global SpellStopTargeting: function(): unknown...
global SpellTargetItem: function(unknown): unknown...
global SpellTargetUnit: function(string): unknown...
global SplashFrameCanBeShown: function(): boolean
global SplashHeaderFont: W.Global_SplashHeaderFont
global SplitGuildBankItem: function(unknown, unknown, unknown): unknown...
global SplitTextIntoHeaderAndNonHeader: function(string): string | nil, string | nil
global SplitTextIntoLines: function(string, string): {string}
global StartAttack: function(? string, ? boolean)
global StartAutoRun: function(): unknown...
global StartDuel: W.Global_StartDuel
global StartSoloShuffleWarGameByName: function(): unknown...
global StartSpectatorSoloShuffleWarGame: function(): unknown...
global StartSpectatorWarGame: function(unknown, unknown, unknown, unknown, unknown): unknown...
global StartWarGame: function(unknown, unknown, unknown): unknown...
global StartWarGameByName: function(unknown): unknown...
global StopAttack: function()
global StopAutoRun: function(): unknown...
global StopCinematic: function()
global StopMacro: function(): unknown...
global StopMusic: function(): unknown...
global StopSound: function(number, ? number): unknown...
global StoreSecureReference: function(unknown, unknown): unknown...
global StrafeLeftStart: function(): unknown...
global StrafeLeftStop: function(): unknown...
global StrafeRightStart: function(): unknown...
global StrafeRightStop: function(): unknown...
global Stuck: function()
global SubSpellFont: W.Global_SubSpellFont
global SubZoneTextFont: W.Global_SubZoneTextFont
global SubmitRequiredGuildRename: function(): unknown...
global SummonRandomCritter: function(): unknown...
global SupportsClipCursor: function(): boolean
global SurrenderArena: function(): unknown...
global SwapRaidSubgroup: function(number, number): unknown...
global SwapToGlobalEnvironment: function(): unknown...
global SwitchAchievementSearchTab: function(unknown): unknown...
global System15Font: W.Global_System15Font
global SystemFont22_Outline: W.Global_SystemFont22_Outline
global SystemFont22_Shadow_Outline: W.Global_SystemFont22_Shadow_Outline
global SystemFont_Huge1: W.Global_SystemFont_Huge1
global SystemFont_Huge1_Outline: W.Global_SystemFont_Huge1_Outline
global SystemFont_Huge2: W.Global_SystemFont_Huge2
global SystemFont_Huge4: W.Global_SystemFont_Huge4
global SystemFont_InverseShadow_Small: W.Global_SystemFont_InverseShadow_Small
global SystemFont_Large: W.Global_SystemFont_Large
global SystemFont_LargeNamePlate: W.Global_SystemFont_LargeNamePlate
global SystemFont_LargeNamePlateFixed: W.Global_SystemFont_LargeNamePlateFixed
global SystemFont_Med1: W.Global_SystemFont_Med1
global SystemFont_Med2: W.Global_SystemFont_Med2
global SystemFont_Med3: W.Global_SystemFont_Med3
global SystemFont_NamePlate: W.Global_SystemFont_NamePlate
global SystemFont_NamePlateCastBar: W.Global_SystemFont_NamePlateCastBar
global SystemFont_NamePlateFixed: W.Global_SystemFont_NamePlateFixed
global SystemFont_Outline: W.Global_SystemFont_Outline
global SystemFont_OutlineThick_Huge2: W.Global_SystemFont_OutlineThick_Huge2
global SystemFont_OutlineThick_Huge4: W.Global_SystemFont_OutlineThick_Huge4
global SystemFont_OutlineThick_WTF: W.Global_SystemFont_OutlineThick_WTF
global SystemFont_Outline_Small: W.Global_SystemFont_Outline_Small
global SystemFont_Outline_WTF2: W.Global_SystemFont_Outline_WTF2
global SystemFont_Shadow_Huge1: W.Global_SystemFont_Shadow_Huge1
global SystemFont_Shadow_Huge2: W.Global_SystemFont_Shadow_Huge2
global SystemFont_Shadow_Huge2_Outline: W.Global_SystemFont_Shadow_Huge2_Outline
global SystemFont_Shadow_Huge3: W.Global_SystemFont_Shadow_Huge3
global SystemFont_Shadow_Huge4: W.Global_SystemFont_Shadow_Huge4
global SystemFont_Shadow_Huge4_Outline: W.Global_SystemFont_Shadow_Huge4_Outline
global SystemFont_Shadow_Large: W.Global_SystemFont_Shadow_Large
global SystemFont_Shadow_Large2: W.Global_SystemFont_Shadow_Large2
global SystemFont_Shadow_Large_Outline: W.Global_SystemFont_Shadow_Large_Outline
global SystemFont_Shadow_Med1: W.Global_SystemFont_Shadow_Med1
global SystemFont_Shadow_Med1_Outline: W.Global_SystemFont_Shadow_Med1_Outline
global SystemFont_Shadow_Med2: W.Global_SystemFont_Shadow_Med2
global SystemFont_Shadow_Med2_Outline: W.Global_SystemFont_Shadow_Med2_Outline
global SystemFont_Shadow_Med3: W.Global_SystemFont_Shadow_Med3
global SystemFont_Shadow_Med3_Outline: W.Global_SystemFont_Shadow_Med3_Outline
global SystemFont_Shadow_Outline_Huge3: W.Global_SystemFont_Shadow_Outline_Huge3
global SystemFont_Shadow_Small: W.Global_SystemFont_Shadow_Small
global SystemFont_Shadow_Small2: W.Global_SystemFont_Shadow_Small2
global SystemFont_Small: W.Global_SystemFont_Small
global SystemFont_Small2: W.Global_SystemFont_Small2
global SystemFont_Tiny: W.Global_SystemFont_Tiny
global SystemFont_Tiny2: W.Global_SystemFont_Tiny2
global SystemFont_WTF2: W.Global_SystemFont_WTF2
global SystemFont_World: W.Global_SystemFont_World
global SystemFont_World_ThickOutline: W.Global_SystemFont_World_ThickOutline
global System_IME: W.Global_System_IME
global TALENT_LINK_COLOR: W.Global_TALENT_LINK_COLOR
global TOOLTIP_DEFAULT_BACKGROUND_COLOR: W.Global_TOOLTIP_DEFAULT_BACKGROUND_COLOR
global TOOLTIP_DEFAULT_COLOR: W.Global_TOOLTIP_DEFAULT_COLOR
global TOY_LINK_COLOR: W.Global_TOY_LINK_COLOR
global TRADESKILL_EXPERIENCE_COLOR: W.Global_TRADESKILL_EXPERIENCE_COLOR
global TRANSMOGRIFY_COLOR: W.Global_TRANSMOGRIFY_COLOR
global TRANSMOGRIFY_FONT_COLOR: W.Global_TRANSMOGRIFY_FONT_COLOR
global TRIVIAL_DIFFICULTY_COLOR: W.Global_TRIVIAL_DIFFICULTY_COLOR
global TUSKARR_MAJOR_FACTION_COLOR: W.Global_TUSKARR_MAJOR_FACTION_COLOR
global TUTORIAL_BLUE_FONT_COLOR: W.Global_TUTORIAL_BLUE_FONT_COLOR
global TUTORIAL_FONT_COLOR: W.Global_TUTORIAL_FONT_COLOR
global TakeInboxItem: function(number, number): unknown...
global TakeInboxMoney: function(number): unknown...
global TakeInboxTextItem: function(unknown): unknown...
global TakeTaxiNode: function(number): unknown...
global TargetDirectionEnemy: function(number, ? number)
global TargetDirectionFinished: function()
global TargetDirectionFriend: function(number, ? number)
global TargetLastEnemy: function()
global TargetLastFriend: function()
global TargetLastTarget: function()
global TargetNearest: function(? boolean)
global TargetNearestEnemy: function(? boolean)
global TargetNearestEnemyPlayer: function(? boolean)
global TargetNearestFriend: function(? boolean)
global TargetNearestFriendPlayer: function(? boolean)
global TargetNearestPartyMember: function(? boolean)
global TargetNearestRaidMember: function(? boolean)
global TargetPriorityHighlightEnd: function()
global TargetPriorityHighlightStart: function(? boolean)
global TargetSpellReplacesBonusTree: function(): boolean
global TargetToggle: function()
global TargetTotem: function(number)
global TargetUnit: function(? string, ? boolean)
global TaxiGetDestX: function(? number, ? number): number
global TaxiGetDestY: function(? number, ? number): number
global TaxiGetNodeSlot: function(): unknown...
global TaxiGetSrcX: function(? number, ? number): number
global TaxiGetSrcY: function(? number, ? number): number
global TaxiIsDirectFlight: function(): unknown...
global TaxiNodeCost: function(number): number
global TaxiNodeGetType: function(number): string
global TaxiNodeName: function(number): string
global TaxiNodePosition: function(number): number, number
global TaxiRequestEarlyLanding: function(): unknown...
global TextStatusBarText: W.Global_TextStatusBarText
global TextStatusBarTextLarge: W.Global_TextStatusBarTextLarge
global TimeoutResurrect: function()
global ToggleAnimKitDisplay: function()
global ToggleAutoRun: function(): unknown...
global ToggleCollision: function()
global ToggleCollisionDisplay: function()
global ToggleDebugAIDisplay: function()
global ToggleGravity: function()
global TogglePVP: function()
global TogglePetAutocast: function(unknown): unknown...
global TogglePlayerBounds: function()
global TogglePortals: function()
global ToggleRun: function(): unknown...
global ToggleSelfHighlight: function(): boolean
global ToggleSheath: function()
global ToggleTris: function()
global ToggleWindowed: function(): unknown...
global Tooltip_Med: W.Global_Tooltip_Med
global Tooltip_Small: W.Global_Tooltip_Small
global TransmogLocationMixin: W.Global_TransmogLocationMixin
global TransmogPendingInfoMixin: W.Global_TransmogPendingInfoMixin
global TriggerTutorial: function(unknown): unknown...
global TurnInGuildCharter: function(): unknown...
global TurnLeftStart: function(number): unknown...
global TurnLeftStop: function(): unknown...
global TurnOrActionStart: function(): unknown...
global TurnOrActionStop: function(): unknown...
global TurnRightStart: function(number): unknown...
global TurnRightStop: function(): unknown...
global UIFrameFade: function(W.Frame, {unknown: unknown})
global UIFrameFadeIn: function(W.Frame, number, number, number)
global UIFrameFadeOut: function(W.Frame, number, number, number)
global UIParent: W.Global_UIParent
global UISpecialFrames: W.Global_UISpecialFrames
global UI_TOOLS_DISABLED_COLOR: W.Global_UI_TOOLS_DISABLED_COLOR
global UI_TOOLS_LINE_COLOR: W.Global_UI_TOOLS_LINE_COLOR
global UI_TOOLS_PRESSED_BUTTON_COLOR: W.Global_UI_TOOLS_PRESSED_BUTTON_COLOR
global UNCOMMON_GREEN_COLOR: W.Global_UNCOMMON_GREEN_COLOR
global USED_IN_TRADESKILL_COLOR: W.Global_USED_IN_TRADESKILL_COLOR
global UiMapPoint: W.Global_UiMapPoint
global UninviteUnit: function(string, ? string): unknown...
global UnitAffectingCombat: function(string): boolean
global UnitAlliedRaceInfo: function(string): boolean, boolean
global UnitArmor: function(string): W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike
global UnitAttackPower: function(string): W.NumberLike, W.NumberLike, W.NumberLike
global UnitAttackSpeed: function(string): W.NumberLike, W.NumberLike | nil
global UnitBattlePetLevel: function(string): number | nil
global UnitBattlePetSpeciesID: function(string): number | nil
global UnitBattlePetType: function(string): number | nil
global UnitCanAssist: function(string, string, ? boolean, ? boolean): boolean
global UnitCanAttack: function(string, string): boolean
global UnitCanCooperate: function(string, string): boolean
global UnitCanPetBattle: function(string, string): boolean
global UnitCastingDuration: function(string): W.OpaqueValue | nil
global UnitCastingInfo: function(string): W.TextLike, W.TextLike, W.NumberLike, W.NumberLike, W.NumberLike, boolean, W.TextLike, boolean | W.OpaqueValue | nil, W.NumberLike, number | nil, number
global UnitChannelDuration: function(string): W.DurationObject | nil
global UnitChannelInfo: function(string): W.TextLike, W.TextLike, W.NumberLike, W.NumberLike, W.NumberLike, boolean, boolean | W.OpaqueValue | nil, W.NumberLike, boolean, number, number | nil
global UnitChromieTimeID: function(string): number
global UnitClass: function(string): W.TextLike | nil, W.TextLike | nil, W.NumberLike | nil
global UnitClassBase: function(string): W.TextLike | nil, W.NumberLike | nil
global UnitClassFromGUID: function(W.TextLike): W.TextLike | nil, string | nil, number | nil
global UnitClassification: function(string): string
global UnitControllingVehicle: function(string): boolean
global UnitCreatureFamily: function(string): W.TextLike, W.NumberLike
global UnitCreatureID: function(string): W.NumberLike | nil
global UnitCreatureType: function(string): W.TextLike, W.NumberLike
global UnitDamage: function(string): W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike
global UnitDetailedThreatSituation: function(string, string): boolean | W.OpaqueValue | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil
global UnitDistanceSquared: function(string): number, boolean
global UnitEffectiveLevel: function(string): number
global UnitEmpoweredChannelDuration: function(string, ? boolean): W.DurationObject | nil
global UnitEmpoweredStageDurations: function(string): {W.DurationObject} | nil
global UnitEmpoweredStagePercentages: function(string, ? boolean): {number} | nil
global UnitExists: function(? string): boolean
global UnitFactionGroup: function(string, ? boolean): string, string
global UnitFullName: function(string): W.TextLike, W.TextLike
global UnitGUID: function(string): W.TextLike | nil
global UnitGetAvailableRoles: function(string): boolean | W.OpaqueValue | nil, boolean | W.OpaqueValue | nil, boolean | W.OpaqueValue | nil
global UnitGetDetailedHealPrediction: function(string, ? string, W.UnitHealPredictionCalculator)
global UnitGetIncomingHeals: function(string, ? string): W.NumberLike | nil
global UnitGetTotalAbsorbs: function(string): W.NumberLike
global UnitGetTotalHealAbsorbs: function(string): W.NumberLike
global UnitGroupRolesAssigned: function(? string): W.TextLike
global UnitGroupRolesAssignedEnum: function(? string): W.NumberLike
global UnitHPPerStamina: function(string): number
global UnitHasIncomingResurrection: function(string): boolean
global UnitHasLFGDeserter: function(string): boolean
global UnitHasLFGRandomCooldown: function(string): boolean
global UnitHasPowerType: function(string, number): boolean
global UnitHasRelicSlot: function(string): boolean
global UnitHasVehiclePlayerFrameUI: function(? string): boolean
global UnitHasVehicleUI: function(? string): boolean
global UnitHealth: function(string, ? boolean): W.NumberLike
global UnitHealthMax: function(string): W.NumberLike
global UnitHealthMissing: function(string, ? boolean): W.NumberLike
global UnitHealthPercent: W.Global_UnitHealthPercent
global UnitHonor: function(string): number
global UnitHonorLevel: function(string): W.NumberLike
global UnitHonorMax: function(string): number
global UnitInAnyGroup: function(? string, ? number): boolean
global UnitInBattleground: function(? string, ? number): number | nil
global UnitInOtherParty: function(string): boolean
global UnitInParty: function(? string, ? number): boolean
global UnitInPartyIsAI: function(? string): boolean
global UnitInPartyShard: function(string): boolean
global UnitInRaid: function(? string, ? number): W.NumberLike | nil
global UnitInRange: function(string): boolean | W.OpaqueValue, boolean | W.OpaqueValue
global UnitInSubgroup: function(? string, ? number): boolean
global UnitInVehicle: function(string): boolean
global UnitInVehicleControlSeat: function(? string): boolean
global UnitInVehicleHidesPetFrame: function(? string): boolean
global UnitIsAFK: function(string): boolean | W.OpaqueValue
global UnitIsBattlePet: function(string): boolean | nil
global UnitIsBattlePetCompanion: function(string): boolean
global UnitIsBossMob: function(string): boolean
global UnitIsCharmed: function(? string): boolean | W.OpaqueValue
global UnitIsConnected: function(string): boolean
global UnitIsControlling: function(string): boolean
global UnitIsCorpse: function(? string): boolean
global UnitIsDND: function(string): boolean | W.OpaqueValue
global UnitIsDead: function(string): boolean
global UnitIsDeadOrGhost: function(string): boolean
global UnitIsEnemy: function(string, string): boolean
global UnitIsFeignDeath: function(string): boolean
global UnitIsFriend: function(string, string): boolean
global UnitIsGameObject: function(? string): boolean
global UnitIsGhost: function(string): boolean
global UnitIsGroupAssistant: function(string): boolean | W.OpaqueValue
global UnitIsGroupLeader: function(string, ? number): boolean | W.OpaqueValue
global UnitIsHumanPlayer: function(? string, ? number): boolean
global UnitIsInMyGuild: function(string): boolean
global UnitIsInteractable: function(? string): boolean
global UnitIsLieutenant: function(string): boolean
global UnitIsMercenary: function(string): boolean
global UnitIsMinion: function(string): boolean
global UnitIsNPCAsPlayer: function(? string): boolean
global UnitIsOtherPlayersBattlePet: function(? string): boolean
global UnitIsOtherPlayersPet: function(? string): boolean
global UnitIsOwnerOrControllerOfUnit: function(string, string): boolean | W.OpaqueValue
global UnitIsPVP: function(string): boolean | W.OpaqueValue
global UnitIsPVPFreeForAll: function(string): boolean
global UnitIsPVPSanctuary: function(? string): boolean
global UnitIsPlayer: function(? string, ? number): boolean
global UnitIsPlayerControlledOrGroupMember: function(string): boolean
global UnitIsPossessed: function(? string): boolean | W.OpaqueValue
global UnitIsQuestBoss: function(string): boolean
global UnitIsRaidOfficer: function(? string): boolean | W.OpaqueValue
global UnitIsSameServer: function(string): boolean
global UnitIsTapDenied: function(string): boolean
global UnitIsTrivial: function(string): boolean
global UnitIsUnconscious: function(string): boolean
global UnitIsUnit: function(string, string): boolean | W.OpaqueValue | nil
global UnitIsVisible: function(? string): boolean
global UnitIsWildBattlePet: function(string): boolean
global UnitLeadsAnyGroup: function(string): boolean | W.OpaqueValue
global UnitLevel: function(string): number
global UnitName: function(string): W.TextLike, W.TextLike
global UnitNameFromGUID: function(W.TextLike): W.TextLike, W.TextLike
global UnitNameUnmodified: function(string): W.TextLike, W.TextLike
global UnitNameplateShowsWidgetsOnly: function(string): boolean
global UnitNumPowerBarTimers: function(string): number
global UnitOnTaxi: function(string): boolean
global UnitOwnerGUID: function(string): W.TextLike
global UnitPVPName: function(string): W.TextLike
global UnitPartialPower: function(string, ? number, ? boolean): W.NumberLike
global UnitPercentHealthFromGUID: function(W.TextLike): W.NumberLike | nil
global UnitPhaseReason: function(string): W.NumberLike | nil
global UnitPlayerControlled: function(? string): boolean
global UnitPlayerOrPetInParty: function(? string, ? number): boolean
global UnitPlayerOrPetInRaid: function(? string, ? number): boolean
global UnitPosition: function(string): number, number, number, number
global UnitPower: function(string, ? number, ? boolean): W.NumberLike
global UnitPowerBarID: function(string): number
global UnitPowerBarTimerInfo: function(string, ? number): W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil, W.NumberLike | nil
global UnitPowerDisplayMod: function(number): number
global UnitPowerMax: function(string, ? number, ? boolean): W.NumberLike
global UnitPowerMissing: function(string, ? number, ? boolean): W.NumberLike
global UnitPowerPercent: W.Global_UnitPowerPercent
global UnitPowerType: function(string, ? number): number | nil, string | nil, number | nil, number | nil, number | nil
global UnitPvpClassification: function(string): number | nil
global UnitQuestTrivialLevelRange: function(string): number
global UnitQuestTrivialLevelRangeScaling: function(string): number
global UnitRace: function(string): W.TextLike | nil, W.TextLike | nil, W.NumberLike | nil
global UnitRangedAttackPower: function(string): W.NumberLike, W.NumberLike, W.NumberLike
global UnitRangedDamage: function(string): W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike
global UnitReaction: function(string, string): number | nil
global UnitRealmRelationship: function(string): number | nil
global UnitSelectionColor: function(string, ? boolean): number, number, number, number
global UnitSelectionType: function(string, ? boolean): number
global UnitSetRole: function(string, ? string): boolean | nil
global UnitSetRoleEnum: function(string, ? number): boolean | nil
global UnitSex: function(string): W.NumberLike | nil
global UnitSexBase: function(string): W.NumberLike | nil
global UnitShouldDisplayName: function(string): boolean
global UnitShouldDisplaySpellTargetName: function(string): boolean
global UnitSpellHaste: function(string): W.NumberLike
global UnitSpellTargetClass: function(string): W.TextLike
global UnitSpellTargetName: function(string): W.TextLike
global UnitStagger: function(string): W.NumberLike
global UnitStat: function(string, number): W.NumberLike, W.NumberLike, W.NumberLike, W.NumberLike
global UnitSwitchToVehicleSeat: function(string, number)
global UnitTargetsVehicleInRaidUI: function(? string): boolean
global UnitThreatLeadSituation: function(string, string): W.NumberLike | nil
global UnitThreatPercentageOfLead: function(string, string): W.NumberLike | nil
global UnitThreatSituation: function(string, ? string): W.NumberLike | nil
global UnitTokenFromGUID: function(W.TextLike): W.TextLike | nil
global UnitTreatAsPlayerForDisplay: function(string): boolean
global UnitTrialBankedLevels: function(string): number, number, number
global UnitTrialXP: function(string): number
global UnitUsingVehicle: function(string): boolean
global UnitVehicleSeatCount: function(string): number
global UnitVehicleSeatInfo: function(string, number): string | nil, string | nil, string | nil, boolean | nil, boolean | nil
global UnitVehicleSkin: function(? string): number
global UnitWeaponAttackPower: function(string): W.NumberLike, W.NumberLike, W.NumberLike
global UnitWidgetSet: function(string): number | nil
global UnitXP: function(string): number
global UnitXPMax: function(string): number
global UnlearnSpecialization: function(unknown, unknown): unknown...
global UnmuteSoundFile: function(number | string): unknown...
global UnregisterEventCallback: function(string, function())
global UnregisterUnitEventCallback: function(string, function(), string)
global UpdateAddOnCPUUsage: function()
global UpdateAddOnMemoryUsage: function()
global UpdateInventoryAlertStatus: function(): unknown...
global UpdateWarGamesList: function(): unknown...
global UpdateWindow: function()
global UseAction: function(number, ? number, ? number): unknown...
global UseInventoryItem: function(number): unknown...
global UseItemByName: function(number | string, ? string)
global UseQuestLogSpecialItem: function(unknown): unknown...
global UseToy: function(number): unknown...
global UseToyByName: function(string): unknown...
global UseWorldMapActionButtonSpellOnQuest: function(): unknown...
global VALDRAKKEN_MAJOR_FACTION_COLOR: W.Global_VALDRAKKEN_MAJOR_FACTION_COLOR
global VENTHYR_RED_COLOR: W.Global_VENTHYR_RED_COLOR
global VERY_DARK_GRAY_COLOR: W.Global_VERY_DARK_GRAY_COLOR
global VERY_LIGHT_GRAY_COLOR: W.Global_VERY_LIGHT_GRAY_COLOR
global Vector2DMixin: W.Global_Vector2DMixin
global Vector3DMixin: W.Global_Vector3DMixin
global VehicleAimDecrement: function(): unknown...
global VehicleAimDownStart: function(): unknown...
global VehicleAimDownStop: function(): unknown...
global VehicleAimGetNormPower: function(): unknown...
global VehicleAimIncrement: function(): unknown...
global VehicleAimRequestAngle: function(): unknown...
global VehicleAimSetNormPower: function(): unknown...
global VehicleAimUpStart: function(): unknown...
global VehicleAimUpStop: function(): unknown...
global VehicleExit: function(): unknown...
global VehicleMenuBarStatusBarText: W.Global_VehicleMenuBarStatusBarText
global VehicleNextSeat: function(): unknown...
global VehiclePrevSeat: function(): unknown...
global ViewGuildRecipes: function(unknown): unknown...
global WARBOARD_OPTION_TEXT_COLOR: W.Global_WARBOARD_OPTION_TEXT_COLOR
global WARNING_FONT_COLOR: W.Global_WARNING_FONT_COLOR
global WEB_MAJOR_FACTION_COLOR: W.Global_WEB_MAJOR_FACTION_COLOR
global WHITE_FONT_COLOR: W.Global_WHITE_FONT_COLOR
global WOW_PROJECT_BURNING_CRUSADE_CLASSIC: W.Global_WOW_PROJECT_BURNING_CRUSADE_CLASSIC
global WOW_PROJECT_CATACLYSM_CLASSIC: W.Global_WOW_PROJECT_CATACLYSM_CLASSIC
global WOW_PROJECT_CLASSIC: W.Global_WOW_PROJECT_CLASSIC
global WOW_PROJECT_ID: W.Global_WOW_PROJECT_ID
global WOW_PROJECT_MAINLINE: W.Global_WOW_PROJECT_MAINLINE
global WOW_PROJECT_MISTS_CLASSIC: W.Global_WOW_PROJECT_MISTS_CLASSIC
global WOW_PROJECT_WRATH_CLASSIC: W.Global_WOW_PROJECT_WRATH_CLASSIC
global WarGameRespond: function(unknown): unknown...
global WhiteNormalNumberFont: W.Global_WhiteNormalNumberFont
global WithdrawGuildBankMoney: function(unknown): unknown...
global WorldLootObjectExists: function(? string): boolean
global WorldMapFrame: W.Global_WorldMapFrame
global WorldMapTextFont: W.Global_WorldMapTextFont
global WrapTextInColorCode: function(string, string): string
global YELLOW_FONT_COLOR: W.Global_YELLOW_FONT_COLOR
global YELLOW_THREAT_COLOR: W.Global_YELLOW_THREAT_COLOR
global ZoneTextFont: W.Global_ZoneTextFont
global addframetext: function(string): unknown...
global canaccessallvalues: function(unknown...): boolean
global canaccesssecrets: function(): boolean
global canaccesstable: function(unknown): boolean
global canaccessvalue: function(unknown): boolean
global debuglocals: function(number): string | nil
global debugprofilestart: function()
global debugprofilestop: function(): number
global debugstack: W.Global_debugstack
global dropsecretaccess: function()
global dumpobject: function(? unknown): unknown | nil
global forceinsecure: function(): unknown...
global geterrorhandler: function(): unknown
global getglobal: function(string): unknown
global hasanysecretvalues: function(unknown...): boolean
global hooksecurefunc: W.Global_hooksecurefunc
global issecrettable: function(unknown): boolean
global issecretvalue: function(unknown): boolean
global issecure: function(): boolean
global issecurevalue: function(unknown): boolean, string | nil
global issecurevariable: W.Global_issecurevariable
global mapvalues: function(unknown, unknown...): unknown...
global message: function(string, boolean)
global pcallwithenv: function(string, {unknown: unknown}): boolean
global ripairs: function({unknown: unknown}): unknown, {unknown: unknown}, number
global scrub: function(unknown...): unknown...
global scrubsecretvalues: function(unknown...): unknown...
global secretunwrap: function(unknown...): unknown | nil...
global secretwrap: function(unknown...): unknown...
global securecall: W.Global_securecall
global securecallfunction: function<A, R>(W.Callback_49bbf91b6c33<A, R>, A...): R...
global securecallmethod: function(unknown, string, unknown...): unknown...
global securecopy: function(? unknown, ? W.SecureCopyOptions): unknown
global secureexecuterange: function({unknown: unknown}, unknown, unknown...): unknown...
global seterrorhandler: function(unknown): unknown...
global setglobal: function(string, unknown)
global setsecurehookforbidden: function(): unknown...
global settablesecurity: function(unknown, number)
global tAppendAll: function({unknown: unknown}, {unknown: unknown})
global tCompare: function({unknown: unknown}, {unknown: unknown}, ? number): boolean
global tContains: function({unknown: unknown}, unknown): boolean
global tDeleteItem: function({unknown: unknown}, unknown)
global tFilter: function({unknown: unknown}, unknown, boolean): {unknown: unknown}
global tIndexOf: function({unknown: unknown}, unknown): number | nil
global tInvert: function({unknown: unknown}): {unknown: unknown}
global tUnorderedRemove: function({unknown: unknown}, number)
-- Additional globals exist only in Blizzard's secure environment.
global C_CombatLogSecure: W.Global_C_CombatLogSecure
global C_PingSecure: W.Global_C_PingSecure
global GetForbiddenObjectTable: W.Global_GetForbiddenObjectTable
