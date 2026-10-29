-- Generated public native environment; addon fragments add only owned globals.
local type W = require("wow")
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
global AbandonSkill: function(number): W.Data...
global AbbreviateLargeNumbers: function(W.MaybeSecret<number>, ? W.NumberAbbrevOptions): string
global AbbreviateNumbers: function(W.MaybeSecret<number>, ? W.NumberAbbrevOptions): string
global AcceptAreaSpiritHeal: function()
global AcceptArenaTeam: function()
global AcceptBattlefieldPort: function(number, boolean): W.Data...
global AcceptDuel: function(): W.Data...
global AcceptGroup: function(): W.Data...
global AcceptGuild: function()
global AcceptProposal: function(): W.Data...
global AcceptQuest: function(): W.Data...
global AcceptResurrect: function()
global AcceptSpellConfirmationPrompt: function(number): W.Data...
global AcceptTrade: function(): W.Data...
global AccumulateIf: function(any, any): number
global AchievementCriteriaFont: W.Global_AchievementCriteriaFont
global AchievementDateFont: W.Global_AchievementDateFont
global AchievementDescriptionFont: W.Global_AchievementDescriptionFont
global AchievementFont_Small: W.Global_AchievementFont_Small
global AchievementPointsFont: W.Global_AchievementPointsFont
global AchievementPointsFontSmall: W.Global_AchievementPointsFontSmall
global AcknowledgeAutoAcceptQuest: function(): W.Data...
global AcknowledgeSurvey: function(any): W.Data...
global ActionBindsItem: function()
global ActionButton_HideOverlayGlow: function(any)
global ActionButton_ShowOverlayGlow: function(any)
global AddAutoQuestPopUp: function(number, string): W.Data...
global AddChatWindowChannel: function(number, string): W.Data...
global AddChatWindowMessages: function(number, string): W.Data...
global AddSourceLocationExclude: function(string)
global Ambiguate: function(W.MaybeSecret<string>, string): string
global AntiAliasingSupported: function(): W.Data...
global ArchaeologyGetIconInfo: function(number): W.Data...
global ArchaeologyMapUpdateAll: function(number): number
global ArcheologyGetVisibleBlobID: function(number): number
global AreClassRolesSoftSuggestions: function(): boolean
global AreColorsEqual: function(W.colorRGBA, W.colorRGBA): boolean
global AreDangerousScriptsAllowed: function(): boolean
global AreTalentsLocked: function(): W.Data...
global AreVector2DEqual: function(W.Vector2DMixin, W.Vector2DMixin): boolean
global AreVector3DEqual: function(W.Vector3DMixin, W.Vector3DMixin): boolean
global ArenaTeamDisband: function(number)
global ArenaTeamInviteByName: function(number, string)
global ArenaTeamLeave: function(number)
global ArenaTeamSetLeaderByName: function(number, string)
global ArenaTeamUninviteByName: function(number, string)
global ArtifactAppearanceSetHighlightFont: W.Global_ArtifactAppearanceSetHighlightFont
global ArtifactAppearanceSetNormalFont: W.Global_ArtifactAppearanceSetNormalFont
global AscendStop: function(): W.Data...
global AssistUnit: function(? string, ? boolean)
global AttachGlyphToSpell: function(any): W.Data...
global AttackTarget: function()
global AuraUtil: W.Global_AuraUtil
global AutoChooseCurrentGraphicsSetting: function(): W.Data...
global AutoEquipCursorItem: function()
global AutoLootMailItem: function(any): W.Data...
global AutoStoreGuildBankItem: function(number, number): W.Data...
global BANK_BAG_BACKGROUND_COLOR: W.Global_BANK_BAG_BACKGROUND_COLOR
global BATTLENET_FONT_COLOR: W.Global_BATTLENET_FONT_COLOR
global BATTLEPET_ABILITY_LINK_COLOR: W.Global_BATTLEPET_ABILITY_LINK_COLOR
global BIND_TRADE_TOOLTIP_COLOR: W.Global_BIND_TRADE_TOOLTIP_COLOR
global BLACK_FONT_COLOR: W.Global_BLACK_FONT_COLOR
global BLUE_FONT_COLOR: W.Global_BLUE_FONT_COLOR
global BNAcceptFriendInvite: function(any): W.Data...
global BNCheckBattleTagInviteToGuildMember: function(any): W.Data...
global BNCheckBattleTagInviteToUnit: function(any): W.Data...
global BNConnected: function(): boolean
global BNDeclineFriendInvite: function(any): W.Data...
global BNFeaturesEnabled: function(): W.Data...
global BNFeaturesEnabledAndConnected: function(): W.Data...
global BNGetBlockedInfo: function(any): W.Data...
global BNGetDisplayName: function(any): W.Data...
global BNGetFOFInfo: function(boolean, boolean, number): number, string, boolean
global BNGetFriendIndex: function(number): number
global BNGetFriendInviteInfo: function(number): number, number, boolean, string, number
global BNGetInfo: function(): number, string, number, string, boolean, boolean, boolean
global BNGetNumBlocked: function(): W.Data...
global BNGetNumFOF: function(any, any, any): W.Data...
global BNGetNumFriendInvites: function(): W.Data...
global BNGetNumFriends: function(): number, number, number, number
global BNGetSelectedBlock: function(): W.Data...
global BNGetSelectedFriend: function(): W.Data...
global BNInviteFriend: function(any): W.Data...
global BNIsBlocked: function(any): W.Data...
global BNIsFriend: function(any): W.Data...
global BNIsSelf: function(any): W.Data...
global BNRemoveFriend: function(any): W.Data...
global BNRequestFOFInfo: function(any): W.Data...
global BNRequestInviteFriend: function(any, any, any, any): W.Data...
global BNSendFriendInvite: function(any, any): W.Data...
global BNSendFriendInviteByID: function(any, any): W.Data...
global BNSendVerifiedBattleTagInvite: function(): W.Data...
global BNSetBlocked: function(any, any): W.Data...
global BNSetFriendFavoriteFlag: function(number, boolean): W.Data...
global BNSetFriendNote: function(number, string): W.Data...
global BNSetSelectedBlock: function(any): W.Data...
global BNSetSelectedFriend: function(any): W.Data...
global BNSummonFriendByIndex: function(any): W.Data...
global BNTokenFindName: function(any): W.Data...
global BRIGHTBLUE_FONT_COLOR: W.Global_BRIGHTBLUE_FONT_COLOR
global BRONZE_MATERIAL_TEXT_COLOR: W.Global_BRONZE_MATERIAL_TEXT_COLOR
global BRONZE_MATERIAL_TITLETEXT_COLOR: W.Global_BRONZE_MATERIAL_TITLETEXT_COLOR
global BattlefieldMgrEntryInviteResponse: function(any, any): W.Data...
global BattlefieldMgrExitRequest: function(any): W.Data...
global BattlefieldMgrQueueInviteResponse: function(any, any): W.Data...
global BattlefieldMgrQueueRequest: function(): W.Data...
global BeginTrade: function()
global BindEnchant: function()
global BossEmoteNormalHuge: W.Global_BossEmoteNormalHuge
global BreakUpLargeNumbers: function(W.MaybeSecret<number>, ? boolean): string
global BuyGuildBankTab: function(): W.Data...
global BuyGuildCharter: function(string): W.Data...
global BuyMerchantItem: function(number, ? number): W.Data...
global BuyTrainerService: function(number): W.Data...
global BuybackItem: function(number): W.Data...
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
global CallCompanion: function(string, number): W.Data...
global CallbackRegistryMixin: W.Global_CallbackRegistryMixin
global CameraOrSelectOrMoveStart: function(): W.Data...
global CameraOrSelectOrMoveStop: function(? boolean): W.Data...
global CameraZoomIn: function(number): W.Data...
global CameraZoomOut: function(number): W.Data...
global CanAffordMerchantItem: function(any): W.Data...
global CanAutoSetGamePadCursorControl: function(): W.Data...
global CanBeRaidTarget: function(string): boolean
global CanCancelScene: function(): boolean
global CanChangePlayerDifficulty: function(): boolean, boolean
global CanComplainInboxItem: function(number): boolean
global CanDualWield: function(): boolean
global CanEditGuildBankTabInfo: function(): W.Data...
global CanEditGuildEvent: function(): W.Data...
global CanEditGuildInfo: function(): W.Data...
global CanEditGuildTabInfo: function(any): W.Data...
global CanEditMOTD: function(): boolean
global CanEditPublicNote: function(): W.Data...
global CanEjectPassengerFromSeat: function(number): boolean
global CanExitVehicle: function(): W.Data...
global CanGamePadControlCursor: function(): W.Data...
global CanGuildBankRepair: function(): W.Data...
global CanGuildDemote: function(): boolean
global CanGuildInvite: function(): boolean
global CanGuildPromote: function(): boolean
global CanGuildRemove: function(): W.Data...
global CanHearthAndResurrectFromArea: function(): W.Data...
global CanInitiateWarGame: function(): W.Data...
global CanInspect: function(string): boolean
global CanItemBeSocketedToArtifact: function(number): boolean
global CanJoinBattlefieldAsGroup: function(): boolean
global CanLootUnit: function(string): boolean, boolean
global CanMapChangeDifficulty: function(? number): boolean
global CanMerchantRepair: function(): boolean
global CanPartyLFGBackfill: function(): W.Data...
global CanReplaceGuildMaster: function(): boolean
global CanResetTutorials: function(): W.Data...
global CanScanResearchSite: function(): boolean
global CanShowAchievementUI: function(): boolean
global CanShowResetInstances: function(): boolean
global CanShowSetRoleButton: function(): boolean
global CanSignPetition: function(): W.Data...
global CanSolveArtifact: function(): boolean
global CanSurrenderArena: function(): W.Data...
global CanSwitchVehicleSeat: function(): boolean
global CanSwitchVehicleSeats: function(): boolean
global CanUpgradeToCurrentExpansion: function(): boolean
global CanViewGuildRecipes: function(number): boolean
global CanWithdrawGuildBankMoney: function(): W.Data...
global CancelAreaSpiritHeal: function()
global CancelDuel: function(): W.Data...
global CancelItemTempEnchantment: function(number): W.Data...
global CancelLogout: function()
global CancelMasterLootRoll: function(any): W.Data...
global CancelPendingEquip: function(number)
global CancelPetPossess: function(): W.Data...
global CancelPreloadingMovie: function(number)
global CancelScene: function(): W.Data...
global CancelShapeshiftForm: function(): W.Data...
global CancelSpellByName: function(any): W.Data...
global CancelTrade: function()
global CancelTradeAccept: function(): W.Data...
global CancelUnitBuff: function(string, number, ? string): W.Data...
global CannotBeResurrected: function(): W.Data...
global CaseAccentInsensitiveParse: function(string): string
global CastPetAction: function(number, ? string): W.Data...
global CastShapeshiftForm: function(number): W.Data...
global CastSpell: function(number, string): W.Data...
global CastSpellByID: function(any, any): W.Data...
global CastSpellByName: function(string, ? string): W.Data...
global CenterCamera: function(): W.Data...
global ChangeChatColor: function(string, number, number, number): W.Data...
global ChannelBan: function(string, string): W.Data...
global ChannelInvite: function(string, string): W.Data...
global ChannelKick: function(string, string): W.Data...
global ChannelModerator: function(string, string): W.Data...
global ChannelSetAllSilent: function(): W.Data...
global ChannelSetPartyMemberSilent: function(any, any): W.Data...
global ChannelToggleAnnouncements: function(string, string): W.Data...
global ChannelUnban: function(string, string): W.Data...
global ChannelUnmoderator: function(string, string): W.Data...
global ChatBubbleFont: W.Global_ChatBubbleFont
global ChatFontNormal: W.Global_ChatFontNormal
global ChatFontSmall: W.Global_ChatFontSmall
global ChatFrame1: W.Global_ChatFrame1
global ChatFrame10: W.Global_ChatFrame10
global ChatFrame2: W.Global_ChatFrame2
global ChatFrame3: W.Global_ChatFrame3
global ChatFrame4: W.Global_ChatFrame4
global ChatFrame5: W.Global_ChatFrame5
global ChatFrame6: W.Global_ChatFrame6
global ChatFrame7: W.Global_ChatFrame7
global ChatFrame8: W.Global_ChatFrame8
global ChatFrame9: W.Global_ChatFrame9
global ChatTypeInfo: W.Global_ChatTypeInfo
global CheckInbox: function(): W.Data...
global CheckInteractDistance: function(string, number): boolean
global CheckTalentMasterDist: function(): boolean
global CinematicFinished: function(number, ? boolean, ? boolean)
global CinematicStarted: function(number, number, ? boolean)
global ClassicExpansionAtLeast: function(number): boolean
global ClassicExpansionAtMost: function(number): boolean
global ClearAchievementComparisonUnit: function(): W.Data...
global ClearAchievementSearchString: function(): W.Data...
global ClearAllLFGDungeons: function(any): W.Data...
global ClearAutoAcceptQuestSound: function(): W.Data...
global ClearBattlemaster: function(): W.Data...
global ClearClampedTextureRotation: function(W.Texture)
global ClearCursor: function()
global ClearCursorHoveredItem: function()
global ClearFailedPVPTalentIDs: function(): W.Data...
global ClearFailedTalentIDs: function(): W.Data...
global ClearFocus: function()
global ClearInspectPlayer: function(): W.Data...
global ClearOutage: function()
global ClearOverrideBindings: function(W.Frame): W.Data...
global ClearPartyAssignment: function(): W.Data...
global ClearPendingBindConversionItem: function()
global ClearRaidMarker: function(? number)
global ClearSendMail: function(): W.Data...
global ClearTarget: function(): boolean
global ClearTutorials: function(): W.Data...
global ClickSendMailItemButton: function(? number, ? boolean): W.Data...
global ClickTargetTradeButton: function(any): W.Data...
global ClickTradeButton: function(any): W.Data...
global ClickWorldMapActionButton: function(): W.Data...
global CloseGuildBankFrame: function(): W.Data...
global CloseGuildRegistrar: function(): W.Data...
global CloseGuildRoster: function(): W.Data...
global CloseItemText: function(): W.Data...
global CloseLoot: function(? boolean): W.Data...
global CloseMail: function(): W.Data...
global CloseMerchant: function(): W.Data...
global ClosePetition: function(): W.Data...
global CloseQuest: function(): W.Data...
global CloseResearch: function(): W.Data...
global CloseTabardCreation: function(): W.Data...
global CloseTaxiMap: function(): W.Data...
global CloseTrade: function(): W.Data...
global CloseTrainer: function(): W.Data...
global ClosestGameObjectPosition: function(number): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global ClosestUnitPosition: function(number): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global CollapseGuildTradeSkillHeader: function(any): W.Data...
global CollapseQuestHeader: function(number, boolean): W.Data...
global CollapseWarGameHeader: function(any): W.Data...
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
global CompleteLFGReadyCheck: function(any): W.Data...
global CompleteLFGRoleCheck: function(any): W.Data...
global CompleteQuest: function(): W.Data...
global ConfirmAcceptQuest: function(): W.Data...
global ConfirmBNRequestInviteFriend: function(number, ? boolean, ? boolean, ? boolean): W.Data...
global ConfirmBindOnUse: function()
global ConfirmLootRoll: function(number, number): W.Data...
global ConfirmLootSlot: function(number): W.Data...
global ConfirmNoRefundOnUse: function()
global ConfirmOnUse: function()
global ConfirmReadyCheck: function(? number): W.Data...
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
global ContainsIf: function(any, any): boolean
global ContinuableContainer: W.Global_ContinuableContainer
global ConvertItemToBindToAccount: function()
global CopyTable: function<K, V>({K: V}, ? boolean): {K: V}
global CopyToClipboard: function(string, ? boolean): number
global CoreAbilityFont: W.Global_CoreAbilityFont
global CreateAbbreviateConfig: function({W.NumberAbbreviationBreakpoint}): W.AbbreviateConfig
global CreateAndInitFromMixin: function<T>(T, any...): T
global CreateAtlasMarkup: function(number | string, ? number, ? number, ? number, ? number, ? number, ? number, ? number): string
global CreateAtlasMarkupWithAtlasSize: function(number | string, ? number, ? number, ? number, ? number, ? number, any): string
global CreateColor: function(number, number, number, ? number): W.colorRGBA
global CreateColorFromBytes: function(number, number, number, number): W.colorRGBA
global CreateColorFromHexString: function(string): W.colorRGBA
global CreateFont: function(string): W.Font
global CreateFontFamily: function(string, {W.CreateFontFamilyMemberInfo}): W.Font
global CreateFrame: W.Global_CreateFrame
global CreateFrameWithOptions: function(W.CreateFrameOptions): W.Frame
global CreateFromMixins: function(any...): W.Data
global CreateMacro: function(string, number | string, ? string, ? boolean): number
global CreateNewRaidProfile: function(any, any): W.Data...
global CreateSecureDelegate: function(any, ? W.SecureDelegateOptions): W.Data
global CreateTextureMarkup: function(number | string, number, number, number, number, number, number, number, number, ? number, ? number): string
global CreateUnitHealPredictionCalculator: function(): W.UnitHealPredictionCalculator
global CreateVector2D: function(number, number): W.Vector2DMixin
global CreateVector3D: function(number, number, number): W.Vector3DMixin
global CreateWindow: function(? boolean, ? boolean): W.Data
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
global DeclineChannelInvite: function(string): W.Data...
global DeclineGroup: function(): W.Data...
global DeclineGuild: function()
global DeclineName: function(string, ? number, number): string...
global DeclineQuest: function(): W.Data...
global DeclineResurrect: function()
global DeclineSpellConfirmationPrompt: function(number): W.Data...
global DeleteCursorItem: function()
global DeleteGMTicket: function(): W.Data...
global DeleteInboxItem: function(number): W.Data...
global DeleteMacro: function(number | string): W.Data...
global DeleteRaidProfile: function(any): W.Data...
global DemoteAssistant: function(any): W.Data...
global DepositGuildBankMoney: function(any): W.Data...
global DescendStop: function(): W.Data...
global DestinyFontHuge: W.Global_DestinyFontHuge
global DestinyFontLarge: W.Global_DestinyFontLarge
global DestinyFontMed: W.Global_DestinyFontMed
global DestroyTotem: function(number)
global DetectWowMouse: function(): W.Data...
global DevTools_Dump: function(any, ? string)
global DialogButtonHighlightText: W.Global_DialogButtonHighlightText
global DialogButtonNormalText: W.Global_DialogButtonNormalText
global DismissCompanion: function(string): W.Data...
global Dismount: function()
global DisplayChannelOwner: function(string): W.Data...
global DoMasterLootRoll: function(any): W.Data...
global DoReadyCheck: function(): W.Data...
global DoesCurrentLocaleSellExpansionLevels: function(): boolean
global DoesItemContainSpec: function(number | string, number, ? number): boolean
global DoesTemplateExist: function(any): W.Data...
global DropCursorMoney: function()
global DropItemOnUnit: function(string)
global DropdownButtonMixin: W.Global_DropdownButtonMixin
global DungeonAppearsInRandomLFD: function(any): W.Data...
global EASY_DIFFICULTY_COLOR: W.Global_EASY_DIFFICULTY_COLOR
global EDIT_MODE_GRID_CENTER_LINE_COLOR: W.Global_EDIT_MODE_GRID_CENTER_LINE_COLOR
global EDIT_MODE_GRID_LINE_COLOR: W.Global_EDIT_MODE_GRID_LINE_COLOR
global EDIT_MODE_LAYOUT_LINK_COLOR: W.Global_EDIT_MODE_LAYOUT_LINK_COLOR
global EJ_ClearSearch: function(): W.Data...
global EJ_EndSearch: function(): W.Data...
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
global EJ_GetSectionPath: function(number): number, number, number, number
global EJ_GetTierInfo: function(number): string, string
global EJ_HandleLinkPath: function(number, number): number, number, number, number
global EJ_InstanceIsRaid: function(): boolean
global EJ_IsLootListOutOfDate: function(): boolean
global EJ_IsSearchFinished: function(): boolean
global EJ_IsValidInstanceDifficulty: function(number): boolean
global EJ_ResetLootFilter: function(): W.Data...
global EJ_SPELL_COLOR: W.Global_EJ_SPELL_COLOR
global EJ_SelectEncounter: function(number): W.Data...
global EJ_SelectInstance: function(number): W.Data...
global EJ_SelectTier: function(number): W.Data...
global EJ_SetDifficulty: function(number): W.Data...
global EJ_SetLootFilter: function(number, number): W.Data...
global EJ_SetSearch: function(string): W.Data...
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
global EnumerateFrames: function(? W.Frame): W.Frame
global EnumerateServerChannels: function(): string...
global EquipCursorItem: function(number)
global EquipItemByName: function(number | string, ? number)
global EquipPendingItem: function(number)
global EquipmentManager_UnpackLocation: function(number)
global ErrorFont: W.Global_ErrorFont
global EventFrameMixin: W.Global_EventFrameMixin
global EventRegistry: W.Global_EventRegistry
global ExpandGuildTradeSkillHeader: function(any): W.Data...
global ExpandQuestHeader: function(number, boolean): W.Data...
global ExpandWarGameHeader: function(any): W.Data...
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
global FindInTableIf: function(any, any): string | number, W.Data
global FindSpellBookSlotBySpellID: function(any, any): W.Data...
global FlagTutorial: function(any): W.Data...
global FlashClientIcon: function(? boolean)
global FlipCameraYaw: function(number): W.Data...
global FlyoutHasSpell: function(number, number): boolean
global FocusFontSmall: W.Global_FocusFontSmall
global FocusUnit: function(? string)
global FollowUnit: function(? string, ? boolean)
global ForceLogout: function()
global ForceQuit: function()
global ForfeitDuel: function(): W.Data...
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
global GMEuropaBugsEnabled: function(): W.Data...
global GMEuropaComplaintsEnabled: function(): W.Data...
global GMEuropaSuggestionsEnabled: function(): W.Data...
global GMEuropaTicketsEnabled: function(): W.Data...
global GMItemRestorationButtonEnabled: function(): W.Data...
global GMQuickTicketSystemEnabled: function(): W.Data...
global GMQuickTicketSystemThrottled: function(): W.Data...
global GMReportLag: function(): W.Data...
global GMRequestPlayerInfo: function(): W.Data...
global GMResponseResolve: function(): W.Data...
global GMSurveyAnswer: function(): W.Data...
global GMSurveyAnswerSubmit: function(any, any, any): W.Data...
global GMSurveyCommentSubmit: function(any): W.Data...
global GMSurveyNumAnswers: function(): W.Data...
global GMSurveyQuestion: function(number): W.Data...
global GMSurveySubmit: function(): W.Data...
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
global GetAchievementNumRewards: function(any): W.Data...
global GetAchievementReward: function(any, any): W.Data...
global GetAchievementSearchProgress: function(): W.Data...
global GetAchievementSearchSize: function(): W.Data...
global GetActionBarToggles: function(): boolean, boolean, boolean, boolean, boolean, boolean, boolean
global GetActionInfo: function(number): string, number | string, string
global GetActiveArtifactByRace: function(number, number): string, string, string, string, string, number, string
global GetActiveLevel: function(any): W.Data...
global GetActiveLootRollIDs: function(): W.Data...
global GetActiveQuestID: function(any): W.Data...
global GetActiveSpecGroup: function(? boolean, ? boolean): number
global GetActiveTitle: function(any): W.Data...
global GetAddOnCPUUsage: function(string | number): number
global GetAddOnMemoryUsage: function(string | number): number
global GetAllowLowLevelRaid: function(): boolean
global GetAllowRecentAlliesSeeLocation: function(): boolean
global GetAlternativeDefaultLanguage: function(): W.Data...
global GetArchaeologyInfo: function(): string
global GetArchaeologyRaceInfo: function(number): string, number, number, number, number, number
global GetArchaeologyRaceInfoByID: function(number): string, number, number, number, number, number
global GetAreaSpiritHealerTime: function(): number
global GetAreaText: function(): string
global GetArenaOpponentSpec: function(number): number, number
global GetArtifactInfoByRace: function(number, number): string, string, number, string, string, number, string, number, number
global GetArtifactProgress: function(): number, number, number
global GetAttackPowerForStat: function(number, number): W.MaybeSecret<number>
global GetAutoDeclineGuildInvites: function(): boolean
global GetAutoDeclineNeighborhoodInvites: function(): boolean
global GetAutoQuestPopUp: function(number): number, string
global GetAvailableBandwidth: function(): number
global GetAvailableLevel: function(any): W.Data...
global GetAvailableLocaleInfo: function(? boolean): {W.LocaleInfo}
global GetAvailableLocales: function(? boolean): string...
global GetAvailableQuestInfo: function(number): boolean, number, boolean, boolean, number, boolean
global GetAvailableTitle: function(any): W.Data...
global GetAverageItemLevel: function(): number, number, number
global GetAvoidance: function(): W.MaybeSecret<number>
global GetBackgroundLoadingStatus: function(): number
global GetBaseDifficultyID: function(number): number
global GetBattlefieldArenaFaction: function(): number
global GetBattlefieldEstimatedWaitTime: function(): number
global GetBattlefieldInstanceExpiration: function(): number
global GetBattlefieldInstanceRunTime: function(): number
global GetBattlefieldMapIconScale: function(): W.Data...
global GetBattlefieldPortExpiration: function(number): number
global GetBattlefieldStatus: function(number): string, string, number, boolean, boolean, string, string, string, boolean, string, string, boolean
global GetBattlefieldTeamInfo: function(number): string, number, number, number
global GetBattlefieldTimeWaited: function(number): number
global GetBattlefieldWinner: function(): number
global GetBattlegroundPoints: function(number): number, number
global GetBestFlexRaidChoice: function(): number
global GetBestRFChoice: function(): number
global GetBillingTimeRested: function(): number
global GetBindLocation: function(): string
global GetBinding: function(number, ? boolean): string, string, string, string, string...
global GetBindingAction: function(string, ? boolean): string
global GetBindingKey: function(string): string...
global GetBindingText: function(? string, ? string, ? boolean): string
global GetBlockChance: function(): W.MaybeSecret<number>
global GetBuildInfo: function(): string, string, string, number, string, string
global GetBuildOption: function(string): boolean
global GetButtonMetatable: function(): W.Data
global GetBuybackItemInfo: function(number): string, number, number, number, number, boolean, boolean
global GetBuybackItemLink: function(any): W.Data...
global GetCVar: function(string): string
global GetCVarBitfield: function(string, number): boolean
global GetCVarBool: function(string): boolean
global GetCVarDefault: function(string): string
global GetCallPetSpellInfo: function(any): W.Data...
global GetCallstackHeight: function(): number
global GetCameraFOVDefaults: function(): number, number, number
global GetCameraZoom: function(): number
global GetCategoryAchievementPoints: function(any, any): W.Data...
global GetCategoryInfo: function(number): string, number, number
global GetCategoryList: function(): {number}
global GetCategoryNumAchievements: function(number, ? boolean): number, number, number
global GetCemeteryPreference: function(): number
global GetChannelDisplayInfo: function(number): string, boolean, boolean, number, number, boolean, string, number
global GetChannelList: function(): number, string, boolean, W.Data...
global GetChannelName: function(string | number): number, string, number, boolean
global GetChatTypeIndex: function(string): number
global GetChatWindowChannels: function(number): string, number, W.Data...
global GetChatWindowInfo: function(number): string, number, number, number, number, number, boolean, boolean, number, boolean
global GetChatWindowMessages: function(number): string...
global GetChatWindowSavedDimensions: function(any): W.Data...
global GetChatWindowSavedPosition: function(any): W.Data...
global GetClassColor: function(W.Literal_8c18208578b0): number, number, number, string
global GetClassColorObj: function(string): W.colorRGB
global GetClassColoredTextForUnit: function(string, string): string
global GetClassInfo: function(number): string, string, number
global GetClassicExpansionLevel: function(): number
global GetClickFrame: function(string): W.Data
global GetClientDisplayExpansionLevel: function(): number
global GetCoinIcon: function(number): number
global GetCoinText: function(number, ? string): string
global GetCoinTextureString: function(number, ? number): string
global GetCollapsingStarCost: function(): number
global GetCombatRating: function(number): W.MaybeSecret<number>
global GetCombatRatingBonus: function(number): W.MaybeSecret<number>
global GetCombatRatingBonusForCombatRatingValue: function(number, number): number
global GetComboPoints: function(string, string): W.MaybeSecret<number>
global GetCompanionInfo: function(string, number): number, string, number, string, boolean, number
global GetComparisonAchievementPoints: function(): W.Data...
global GetComparisonCategoryNumAchievements: function(any): W.Data...
global GetComparisonStatistic: function(number): string
global GetCorpseRecoveryDelay: function(): number
global GetCorruption: function(): number
global GetCorruptionResistance: function(): number
global GetCritChance: function(): W.MaybeSecret<number>
global GetCritChanceProvidesParryEffect: function(): boolean
global GetCriteriaSpell: function(): W.Data...
global GetCurrenciesString: function({any}): string
global GetCurrencyString: function(number, ? number, ? string, ? boolean): string
global GetCurrentArenaSeason: function(): number
global GetCurrentBindingSet: function(): number
global GetCurrentCinematicSummary: function(): string
global GetCurrentEnvironment: function(): W.Data...
global GetCurrentEventID: function(): number
global GetCurrentGlyphNameForSpell: function(any): W.Data...
global GetCurrentGraphicsAPI: function(): string
global GetCurrentGraphicsSetting: function(): W.Data...
global GetCurrentGuildBankTab: function(): W.Data...
global GetCurrentKeyBoardFocus: function(): W.Data...
global GetCurrentLevelFeatures: function(any): W.Data...
global GetCurrentRegion: function(): number
global GetCurrentRegionName: function(): string
global GetCurrentScaledResolution: function(): W.Data...
global GetCurrentTitle: function(): number
global GetCursorDelta: function(): number, number
global GetCursorInfo: function(): string, W.SecretData...
global GetCursorMoney: function(): number
global GetCursorPosition: function(): number, number
global GetDailyQuestsCompleted: function(): W.Data...
global GetDefaultAbbreviationBreakpoints: function(? W.WowLocale): {W.NumberAbbreviationBreakpoint}
global GetDefaultLanguage: function(): string, number
global GetDefaultScale: function(): number
global GetDemotionRank: function(any): W.Data...
global GetDetailedItemLevelInfo: function(number | string): number, boolean, number
global GetDifficultyInfo: function(number): string, string, boolean, boolean, boolean, boolean, number, boolean, number, number, boolean
global GetDodgeChance: function(): W.MaybeSecret<number>
global GetDodgeChanceFromAttribute: function(): W.MaybeSecret<number>
global GetDownloadedPercentage: function(): number
global GetDungeonDifficultyID: function(): number
global GetDungeonForRandomSlot: function(any, any): W.Data...
global GetEditBoxMetatable: function(): W.Data
global GetEquipmentNameFromSpell: function(): W.Data...
global GetErrorCallstackHeight: function(): number
global GetEventCPUUsage: function(): number, number
global GetEventTime: function(number): number, number, string, number
global GetExpansionDisplayInfo: function(number, ? number): W.ExpansionDisplayInfo
global GetExpansionForLevel: function(number): number
global GetExpansionLevel: function(): number
global GetExpansionTrialInfo: function(): boolean, number
global GetExpertise: function(): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global GetExpertisePercent: function(): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global GetFactionColor: function(string): W.colorRGBA
global GetFailedPVPTalentIDs: function(): W.Data...
global GetFailedTalentIDs: function(): W.Data...
global GetFileIDFromPath: function(string): number
global GetFileStreamingStatus: function(): number
global GetFilteredAchievementID: function(number): number
global GetFinalAtlasFromTextureKitIfExists: function(string, any): string
global GetFinalNameFromTextureKit: function(string, any): string
global GetFlexRaidDungeonInfo: function(any): W.Data...
global GetFlyoutID: function(any): W.Data...
global GetFlyoutInfo: function(number): string, string, number, boolean
global GetFlyoutSlotInfo: function(number, number): number, number, boolean, string, number
global GetFollowerTypeIDFromSpell: function(): W.Data...
global GetFontInfo: function(W.Font): W.FontScriptInfo
global GetFontStringMetatable: function(): W.Data
global GetFonts: function(): {string}
global GetFrameCPUUsage: function(W.Frame, ? boolean): number, number
global GetFrameMetatable: function(): W.Data
global GetFramerate: function(): number
global GetFramesRegisteredForEvent: function(string): W.Frame...
global GetFunctionCPUUsage: function(): number, number
global GetGMStatus: function(): W.Data...
global GetGMTicket: function(): W.Data...
global GetGameMessageInfo: function(number): string, number, number
global GetGameTime: function(): number, number
global GetGlobalEnvironment: function(): W.Data...
global GetGraphicsAPIs: function(): string
global GetGraphicsCVarValueForQualityLevel: function(): W.Data...
global GetGreetingText: function(): W.Data...
global GetGroupMemberCounts: function(): W.Data...
global GetGuildAchievementMemberInfo: function(number, number): string
global GetGuildAchievementMembers: function(number): W.Data...
global GetGuildAchievementNumMembers: function(number): number
global GetGuildBankBonusDepositMoney: function(): W.Data...
global GetGuildBankItemInfo: function(number, number): number, number, boolean, boolean, number
global GetGuildBankItemLink: function(number, number): string
global GetGuildBankMoney: function(): number
global GetGuildBankMoneyTransaction: function(number): string, string, number, number, number, number, number
global GetGuildBankTabCost: function(): W.Data...
global GetGuildBankTabInfo: function(number): string, string, boolean, boolean, number, number, boolean
global GetGuildBankTabPermissions: function(number): boolean, boolean, boolean, number
global GetGuildBankText: function(any): W.Data...
global GetGuildBankTransaction: function(number, number): string, string, string, number, number, number, number, number, number, number
global GetGuildBankWithdrawGoldLimit: function(): number
global GetGuildBankWithdrawMoney: function(): number
global GetGuildCategoryList: function(): W.Data...
global GetGuildChallengeInfo: function(any): W.Data...
global GetGuildCharterCost: function(): W.Data...
global GetGuildEventInfo: function(any): W.Data...
global GetGuildFactionGroup: function(): W.Data...
global GetGuildInfo: function(string): string, string, number, string
global GetGuildLogoInfo: function(): W.Data...
global GetGuildMemberRecipes: function(any, any): W.Data...
global GetGuildNewsFilters: function(): W.Data...
global GetGuildNewsMemberName: function(any, any): W.Data...
global GetGuildNewsSort: function(): W.Data...
global GetGuildPerkInfo: function(): W.Data...
global GetGuildRecipeInfoPostQuery: function(): number, number, number
global GetGuildRecipeMember: function(number): string, boolean
global GetGuildRenameRequired: function(): W.Data...
global GetGuildRewardInfo: function(any): W.Data...
global GetGuildRosterInfo: function(number): string, string, number, number, string, string, string, string, boolean, number, string, number, number, boolean, boolean, number, string
global GetGuildRosterLargestAchievementPoints: function(): W.Data...
global GetGuildRosterLastOnline: function(number): number, number, number, number
global GetGuildRosterSelection: function(): number
global GetGuildRosterShowOffline: function(): boolean
global GetGuildTabardFiles: function(): number, number, number, number, number, number
global GetGuildTradeSkillInfo: function(number): number, boolean, string, string, number, number, number, string, string, string, boolean, string, number, string, boolean, number
global GetHaste: function(): W.MaybeSecret<number>
global GetHighlightedNumberDifferenceString: function(string, string): string
global GetHitModifier: function(): W.MaybeSecret<number>
global GetHomePartyInfo: function(any): W.Data
global GetInboxHeaderInfo: function(number): string, string, string, string, number, number, number, number, boolean, boolean, boolean, boolean, boolean
global GetInboxInvoiceInfo: function(number): string, string, string, number, number, number, number, number, number, number, number, boolean
global GetInboxItem: function(number, number): string, number, string, number, number, boolean
global GetInboxItemLink: function(number, number): string
global GetInboxNumItems: function(): number, number
global GetInboxText: function(number): string, string, string, boolean, boolean, boolean
global GetInspectArenaData: function(number): number, number, number, number, number
global GetInspectHonorData: function(): number, number, number, number, number, number
global GetInspectSpecialization: function(string): number
global GetInspectTalent: function(): W.Data...
global GetInstanceBootTimeRemaining: function(): number
global GetInstanceInfo: function(): string, string, number, string, number, number, boolean, number, number, number, boolean
global GetInstanceLockTimeRemaining: function(): number, boolean, number, number
global GetInstanceLockTimeRemainingEncounter: function(number): string, string, boolean, boolean
global GetInventoryAlertStatus: function(string): number
global GetInventoryItemBroken: function(string, number): boolean
global GetInventoryItemCooldown: function(string, number): number, number, number
global GetInventoryItemCount: function(string, number): number
global GetInventoryItemDurability: function(number): number, number
global GetInventoryItemEquippedUnusable: function(any, any): W.Data...
global GetInventoryItemID: function(string, number): number, number
global GetInventoryItemLink: function(string, number): string
global GetInventoryItemQuality: function(string, number): number
global GetInventoryItemTexture: function(string, number): number
global GetInventoryItemsForSlot: function(number, any, ? boolean): W.Data
global GetInventorySlotInfo: function(W.Literal_7a28f0f291f7): number, string, boolean
global GetInviteConfirmationInfo: function(any): number, string, string, boolean, boolean, number, number, number
global GetItemChildInfo: function(number | string, ? number): {number}
global GetItemClassInfo: function(number): string
global GetItemCooldown: function(number | string): number, number, boolean
global GetItemCount: function(number | string, ? boolean, ? boolean, ? boolean): number
global GetItemCreationContext: function(number | string): number, string
global GetItemFamily: function(number | string): number
global GetItemGem: function(string, number): string, string
global GetItemIcon: function(number | string): number
global GetItemInfo: function(number | string): string, string, number, number, number, string, string, number, string, number, number, number, number, number, number, number, boolean
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
global GetJournalInfoForSpellConfirmation: function(any): W.Data...
global GetLFDChoiceCollapseState: function(any): W.Data...
global GetLFDChoiceEnabledState: function(any): W.Data...
global GetLFDChoiceOrder: function(any): W.Data...
global GetLFDLockInfo: function(any, any): W.Data...
global GetLFDLockPlayerCount: function(): W.Data...
global GetLFDRoleLockInfo: function(any, any): W.Data...
global GetLFDRoleRestrictions: function(any): W.Data...
global GetLFGBootProposal: function(): boolean, boolean, boolean, string, number, number, number, string
global GetLFGCategoryForID: function(any): W.Data...
global GetLFGCompletionReward: function(): W.Data...
global GetLFGCompletionRewardItem: function(any): W.Data...
global GetLFGCompletionRewardItemLink: function(any): W.Data...
global GetLFGDeserterExpiration: function(): number
global GetLFGDungeonEncounterInfo: function(number, number): string, string, boolean, boolean
global GetLFGDungeonInfo: function(number): string, number, number, number, number, number, number, number, number, number, string, number, number, string, boolean, number, number, boolean, string, number, boolean, number
global GetLFGDungeonNumEncounters: function(number): number, number
global GetLFGDungeonRewardCapBarInfo: function(number): number, number, number, number, number, number, number, number, number, number
global GetLFGDungeonRewardCapInfo: function(any): W.Data...
global GetLFGDungeonRewardInfo: function(any, any): W.Data...
global GetLFGDungeonRewardLink: function(any, any): W.Data...
global GetLFGDungeonRewards: function(any): W.Data...
global GetLFGDungeonShortageRewardInfo: function(any, any, any): W.Data...
global GetLFGDungeonShortageRewardLink: function(any, any, any): W.Data...
global GetLFGInfoServer: function(any, any): W.Data...
global GetLFGInviteRoleAvailability: function(any): W.Data...
global GetLFGInviteRoleRestrictions: function(any): W.Data...
global GetLFGProposal: function(): boolean, number, number, number, string, string, string, boolean, number, number, number, boolean, boolean, number
global GetLFGProposalEncounter: function(any): W.Data...
global GetLFGProposalMember: function(any): W.Data...
global GetLFGQueueStats: function(number, ? number): boolean, boolean, boolean, boolean, boolean, number, number, number, number, number, string, number, number, number, number, number, number, number
global GetLFGQueuedList: function(any, any): W.Data...
global GetLFGRandomCooldownExpiration: function(): number
global GetLFGRandomDungeonInfo: function(number): number, string, number, number, number, number, number, number, number, number, number, string, number, number, string, boolean, number, number, boolean, string, number, boolean
global GetLFGReadyCheckUpdate: function(): W.Data...
global GetLFGReadyCheckUpdateBattlegroundInfo: function(): W.Data...
global GetLFGRoleShortageRewards: function(number, number): boolean, boolean, boolean, boolean, number, number, number
global GetLFGRoleUpdate: function(): W.Data...
global GetLFGRoleUpdateBattlegroundInfo: function(): string
global GetLFGRoleUpdateMember: function(any): W.Data...
global GetLFGRoleUpdateSlot: function(number): number, number, number
global GetLFGRoles: function(): boolean, boolean, boolean, boolean
global GetLFGSuspendedPlayers: function(any): W.Data...
global GetLFRChoiceOrder: function(any): W.Data
global GetLanguageByIndex: function(number): string, number
global GetLatestCompletedAchievements: function(): W.Data...
global GetLatestCompletedComparisonAchievements: function(): W.Data...
global GetLatestThreeSenders: function(): string, string, string
global GetLatestUpdatedComparisonStats: function(): W.Data...
global GetLatestUpdatedStats: function(): W.Data...
global GetLegacyRaidDifficultyID: function(): number
global GetLifesteal: function(): W.MaybeSecret<number>
global GetLocalGameTime: function(): number, number
global GetLocale: function(): string
global GetLooseMacroIcons: function(): W.Data...
global GetLooseMacroItemIcons: function(): W.Data...
global GetLootInfo: function(): {W.Data}
global GetLootRollItemInfo: function(number): string, string, number, number, boolean, boolean, boolean, boolean, number, number, number, number, boolean
global GetLootRollItemLink: function(number): string
global GetLootRollTimeLeft: function(number): number
global GetLootSlotInfo: function(number): string, string, number, number, number, boolean, boolean, number, boolean, boolean
global GetLootSlotLink: function(number): string
global GetLootSlotType: function(number): number
global GetLootSourceInfo: function(number): string, number
global GetLootSpecialization: function(): number
global GetLootThreshold: function(): number
global GetMacroBody: function(number | string): string
global GetMacroIcons: function(any): W.Data...
global GetMacroIndexByName: function(string): number
global GetMacroInfo: function(number | string): string, number, string
global GetMacroItem: function(number | string): string, string
global GetMacroItemIcons: function(any): W.Data...
global GetMacroSpell: function(number | string): number
global GetManaRegen: function(): W.MaybeSecret<number>, W.MaybeSecret<number>
global GetMasterLootCandidate: function(number, number): string
global GetMastery: function(): W.MaybeSecret<number>
global GetMasteryEffect: function(): W.MaybeSecret<number>, W.MaybeSecret<number>
global GetMaxBattlefieldID: function(): number
global GetMaxLevelForExpansionLevel: function(number): number
global GetMaxLevelForLatestExpansion: function(): number
global GetMaxLevelForPlayerExpansion: function(): number
global GetMaxNumCUFProfiles: function(): W.Data...
global GetMaxPlayerLevel: function(): number
global GetMaxRenderScale: function(): W.Data...
global GetMaxRewardCurrencies: function(): W.Data...
global GetMaxSpellStartRecoveryOffset: function(): number
global GetMaxTalentTier: function(): number
global GetMaximumExpansionLevel: function(): number
global GetMeleeHaste: function(): W.MaybeSecret<number>
global GetMerchantCurrencies: function(): W.Data...
global GetMerchantFilter: function(): W.Data...
global GetMerchantItemCostInfo: function(number): number
global GetMerchantItemCostItem: function(number, number): string, number, string, string
global GetMerchantItemID: function(number): number
global GetMerchantItemInfo: function(number): string, number, number, number, number, boolean, boolean, boolean, number, number
global GetMerchantItemLink: function(number): string
global GetMerchantItemMaxStack: function(number): number
global GetMerchantNumItems: function(): number
global GetMicroIconForRole: function(string): string
global GetMinRenderScale: function(): W.Data...
global GetMinimapZoneText: function(): string
global GetMinimumExpansionLevel: function(): number
global GetMirrorTimerInfo: function(number): string, number, number, number, number, string, number
global GetMirrorTimerProgress: function(string): number
global GetModResilienceDamageReduction: function(): W.MaybeSecret<number>
global GetModifiedClick: function(string): string
global GetModifiedClickAction: function(any): W.Data...
global GetMoney: function(): number
global GetMoneyString: function(number, ? boolean): string
global GetMonitorAspectRatio: function(): W.Data...
global GetMonitorCount: function(): W.Data...
global GetMonitorName: function(): W.Data...
global GetMouseButtonClicked: function(): string
global GetMouseButtonName: function(W.Literal_b7e0de2fb994): string
global GetMouseFoci: function(): {W.ScriptRegion}
global GetMovieDownloadProgress: function(number): boolean, number, number
global GetMultiCastTotemSpells: function(number): number, number, number, number, number, number, number
global GetNativeRealmID: function(): number
global GetNegativeCorruptionEffectInfo: function(): {W.CorruptionEffectInfo}
global GetNetIpTypes: function(): number...
global GetNetStats: function(): number, number, number...
global GetNextAchievement: function(number): number, boolean
global GetNextCompleatedTutorial: function(any): W.Data...
global GetNextPendingInviteConfirmation: function(): W.Data...
global GetNormalizedRealmName: function(): string
global GetNumActiveQuests: function(): number
global GetNumArchaeologyRaces: function(): number
global GetNumArenaOpponentSpecs: function(): W.Data...
global GetNumArenaOpponents: function(): W.Data...
global GetNumArtifactsByRace: function(number): number
global GetNumAutoQuestPopUps: function(): number
global GetNumAvailableQuests: function(): number
global GetNumBattlefieldFlagPositions: function(): W.Data...
global GetNumBattlefieldScores: function(): number
global GetNumBattlefieldVehicles: function(any): W.Data...
global GetNumBattlegroundTypes: function(): number
global GetNumBindings: function(): number
global GetNumBuybackItems: function(): number
global GetNumChannelMembers: function(): W.Data...
global GetNumClasses: function(): number
global GetNumCompanions: function(string): number
global GetNumComparisonCompletedAchievements: function(number): number, number
global GetNumCompletedAchievements: function(): number, number
global GetNumDeclensionSets: function(string, ? number): number
global GetNumDisplayChannels: function(): number
global GetNumDungeonForRandomSlot: function(any): W.Data...
global GetNumExpansions: function(): number
global GetNumFilteredAchievements: function(): number
global GetNumFlexRaidDungeons: function(): number
global GetNumFlyouts: function(): W.Data...
global GetNumGroupChannels: function(): W.Data...
global GetNumGroupMembers: function(? number): number
global GetNumGuildBankMoneyTransactions: function(): W.Data...
global GetNumGuildBankTabs: function(): W.Data...
global GetNumGuildBankTransactions: function(any): W.Data...
global GetNumGuildChallenges: function(): W.Data...
global GetNumGuildEvents: function(): W.Data...
global GetNumGuildMembers: function(): number, number
global GetNumGuildNews: function(): W.Data...
global GetNumGuildPerks: function(): W.Data...
global GetNumGuildRewards: function(): W.Data...
global GetNumGuildTradeSkill: function(): W.Data...
global GetNumLanguages: function(): number
global GetNumLootItems: function(): number
global GetNumMacros: function(): number, number
global GetNumMembersInRank: function(any): W.Data...
global GetNumModifiedClickActions: function(): W.Data...
global GetNumPetitionNames: function(): number
global GetNumQuestChoices: function(): number
global GetNumQuestCurrencies: function(): W.Data...
global GetNumQuestItemDrops: function(): W.Data...
global GetNumQuestItems: function(): number
global GetNumQuestLeaderBoards: function(? number): number
global GetNumQuestLogChoices: function(number, ? boolean): number
global GetNumQuestLogRewardFactions: function(): W.Data...
global GetNumQuestLogRewards: function(? number): number
global GetNumQuestLogTasks: function(): W.Data...
global GetNumQuestRewards: function(): number
global GetNumRFDungeons: function(): number
global GetNumRaidProfiles: function(): W.Data...
global GetNumRandomDungeons: function(): number
global GetNumRandomScenarios: function(): W.Data...
global GetNumRoutes: function(any): W.Data...
global GetNumSavedInstances: function(): number
global GetNumSavedWorldBosses: function(): number
global GetNumScenarios: function(): W.Data...
global GetNumShapeshiftForms: function(): number
global GetNumSpecGroups: function(boolean): number
global GetNumSpecializations: function(? boolean, ? boolean): number
global GetNumSpecializationsForClassID: function(number): number
global GetNumSubgroupMembers: function(? number): number
global GetNumTitles: function(): number
global GetNumTotemSlots: function(): number
global GetNumTrainerServices: function(): number
global GetNumTreasurePickerItems: function(): W.Data...
global GetNumUnspentPvpTalents: function(): W.Data...
global GetNumUnspentTalents: function(): number
global GetNumWarGameTypes: function(): W.Data...
global GetOSLocale: function(): string
global GetObjectiveText: function(): W.Data...
global GetOptOutOfLoot: function(): boolean
global GetOverrideAPBySpellPower: function(): W.MaybeSecret<number>
global GetOverrideSpellPowerByAP: function(): W.MaybeSecret<number>
global GetPVPDesired: function(): boolean
global GetPVPGearStatRules: function(): boolean
global GetPVPLifetimeStats: function(): number, number
global GetPVPRoles: function(): boolean, boolean, boolean
global GetPVPSessionStats: function(): number, number
global GetPVPTimer: function(): number
global GetPVPYesterdayStats: function(): number, number
global GetParryChance: function(): W.MaybeSecret<number>
global GetParryChanceFromAttribute: function(): W.MaybeSecret<number>
global GetPartyAssignment: function(string, ? string, ? boolean): boolean
global GetPartyLFGBackfillInfo: function(): W.Data...
global GetPartyLFGID: function(): W.Data...
global GetPendingGlyphName: function(): W.Data...
global GetPendingInviteConfirmations: function(): W.Data...
global GetPersonalRatedInfo: function(number): number, number, number, number, number, number, number, number
global GetPetActionCooldown: function(number): number, number, number
global GetPetActionInfo: function(number): string, string, boolean, boolean, boolean, boolean, number, boolean, boolean
global GetPetActionSlotUsable: function(number): boolean
global GetPetActionsUsable: function(): W.Data...
global GetPetExperience: function(): number, number
global GetPetFoodTypes: function(): string...
global GetPetIcon: function(): W.Data...
global GetPetMeleeHaste: function(): W.MaybeSecret<number>
global GetPetSpellBonusDamage: function(): W.MaybeSecret<number>
global GetPetTimeRemaining: function(): W.Data...
global GetPetitionInfo: function(): string, string, string, number, string, boolean, number
global GetPetitionNameInfo: function(any): W.Data...
global GetPhysicalScreenSize: function(): number, number
global GetPlayerFacing: function(): number
global GetPlayerInfoByGUID: function(W.MaybeSecret<string>): string, string, string, string, number, string, string, number
global GetPlayerTradeCurrency: function(): W.Data...
global GetPlayerTradeMoney: function(): number
global GetPossessInfo: function(number): string, number, boolean
global GetPowerRegen: function(): W.MaybeSecret<number>, W.MaybeSecret<number>
global GetPowerRegenForPowerType: function(number): W.MaybeSecret<number>, W.MaybeSecret<number>
global GetPrevCompleatedTutorial: function(any): W.Data...
global GetPreviousAchievement: function(number): number
global GetPreviousArenaSeason: function(): W.Data...
global GetPrimarySpecialization: function(): W.Data...
global GetProfessionInfo: function(number): string, number, number, number, number, number, number, number, number, number, string
global GetProfessions: function(): number, number, number, number, number
global GetProgressText: function(): string
global GetPromotionRank: function(any): W.Data...
global GetProtocolTypes: function(): number...
global GetPvpPowerDamage: function(): W.MaybeSecret<number>
global GetPvpPowerHealing: function(): W.MaybeSecret<number>
global GetPvpTalentInfoByID: function(number, ? number, ? boolean, ? string): number, string, number, boolean, boolean, number, boolean, number, number, boolean, boolean
global GetPvpTalentInfoBySpecialization: function(): W.Data...
global GetPvpTalentLink: function(): W.Data...
global GetQuestBackgroundMaterial: function(): string
global GetQuestCurrencyID: function(any, any): W.Data...
global GetQuestExpansion: function(any): W.Data...
global GetQuestFactionGroup: function(number): number
global GetQuestID: function(): number
global GetQuestItemInfo: function(string, number): string, number, number, number, boolean, number
global GetQuestItemInfoLootType: function(any, any): W.Data...
global GetQuestItemLink: function(string, number): string
global GetQuestLink: function(number): string
global GetQuestLogChoiceInfo: function(number, ? number): string, number, number, number, boolean, number
global GetQuestLogChoiceInfoLootType: function(any): W.Data...
global GetQuestLogCompletionText: function(): W.Data...
global GetQuestLogCriteriaSpell: function(): W.Data...
global GetQuestLogItemDrop: function(any): W.Data...
global GetQuestLogItemLink: function(string, number, ? number): string
global GetQuestLogLeaderBoard: function(number, ? number): string, string, boolean
global GetQuestLogPortraitTurnIn: function(): W.Data...
global GetQuestLogQuestText: function(? number): string, string
global GetQuestLogQuestType: function(): W.Data...
global GetQuestLogRewardArtifactXP: function(): W.Data...
global GetQuestLogRewardFactionInfo: function(any): W.Data...
global GetQuestLogRewardHonor: function(any): W.Data...
global GetQuestLogRewardInfo: function(number, ? number): string, string, number, number, boolean, number, number
global GetQuestLogRewardMoney: function(? number): number
global GetQuestLogRewardSkillPoints: function(): W.Data...
global GetQuestLogRewardTitle: function(): W.Data...
global GetQuestLogRewardXP: function(any): W.Data...
global GetQuestLogSpecialItemCooldown: function(number): number, number, number
global GetQuestLogSpecialItemInfo: function(number): string, number, number, boolean
global GetQuestLogTimeLeft: function(): number
global GetQuestMoneyToGet: function(): W.Data...
global GetQuestObjectiveInfo: function(number, number, boolean): string, string, boolean, number, number
global GetQuestPOIBlobCount: function(any): W.Data...
global GetQuestPOILeaderBoard: function(any): W.Data...
global GetQuestPOIs: function(): W.Data...
global GetQuestPortraitGiver: function(): W.Data...
global GetQuestPortraitTurnIn: function(): W.Data...
global GetQuestProgressBarPercent: function(number): number
global GetQuestResetTime: function(): number
global GetQuestReward: function(number): W.Data...
global GetQuestSortIndex: function(number): number
global GetQuestText: function(): W.Data...
global GetQuestUiMapID: function(): W.Data...
global GetRFDungeonInfo: function(number): number, string, number, number, number, number, number, number, number, number, number, string, number, number, string, boolean, number, number, boolean, string, number, boolean, number
global GetRaidDifficultyID: function(): number
global GetRaidProfileFlattenedOptions: function(any): W.Data...
global GetRaidProfileName: function(any): W.Data...
global GetRaidProfileOption: function(any, any): W.Data...
global GetRaidProfileSavedPosition: function(any): W.Data...
global GetRaidRosterInfo: function(number): string, number, number, number, string, string, string, boolean, boolean, string, boolean, string
global GetRaidTargetIndex: function(string): W.MaybeSecret<number>
global GetRandomDungeonBestChoice: function(): W.Data...
global GetRandomScenarioBestChoice: function(): W.Data...
global GetRandomScenarioInfo: function(any): W.Data...
global GetRangedCritChance: function(): W.MaybeSecret<number>
global GetRangedHaste: function(): W.MaybeSecret<number>
global GetRatedBattleGroundInfo: function(): W.Data...
global GetReadyCheckStatus: function(any): W.Data...
global GetReadyCheckTimeLeft: function(): W.Data...
global GetRealZoneText: function(? number): string
global GetRealmID: function(): number
global GetRealmName: function(): string
global GetReleaseTimeRemaining: function(): number
global GetRepairAllCost: function(): number, boolean
global GetResSicknessDuration: function(): string
global GetRestState: function(): number, string, number
global GetRestrictedAccountData: function(): number, number, number
global GetRewardArtifactXP: function(): W.Data...
global GetRewardHonor: function(): W.Data...
global GetRewardMoney: function(): W.Data...
global GetRewardNumSkillUps: function(): W.Data...
global GetRewardPackArtifactPower: function(any): W.Data...
global GetRewardPackCurrencies: function(any): W.Data...
global GetRewardPackItems: function(any): W.Data...
global GetRewardPackMoney: function(any): W.Data...
global GetRewardPackTitle: function(any): W.Data...
global GetRewardPackTitleName: function(any): W.Data...
global GetRewardSkillLineID: function(): W.Data...
global GetRewardSkillPoints: function(): W.Data...
global GetRewardText: function(): string
global GetRewardTitle: function(): W.Data...
global GetRewardXP: function(): number
global GetRuneCooldown: function(number): number, number, boolean
global GetRuneCount: function(number): number
global GetRunningMacro: function(): W.Data...
global GetRunningMacroButton: function(): W.Data...
global GetSavedInstanceChatLink: function(number): string
global GetSavedInstanceEncounterInfo: function(number, number): string, number, boolean, boolean
global GetSavedInstanceInfo: function(number): string, number, number, number, boolean, boolean, number, boolean, number, string, number, number, boolean, number
global GetSavedWorldBossInfo: function(number): string, number, number
global GetScenariosChoiceOrder: function(? any): W.Data
global GetSchoolString: function(number): string
global GetScreenDPIScale: function(): number
global GetScreenHeight: function(): number
global GetScreenWidth: function(): number
global GetScriptBucketThrottleLimits: function(): W.ScriptBucketThrottleLimits
global GetScriptCPUUsage: function(): number
global GetSecondsUntilParentalControlsKick: function(): number
global GetSelectedArtifactInfo: function(): string, string, number, number, string, number, number, number
global GetSelectedDisplayChannel: function(): W.Data...
global GetSelectedWarGameType: function(): W.Data...
global GetSendMailCOD: function(): number
global GetSendMailItem: function(number): string, number, number, number, number
global GetSendMailItemLink: function(number): string
global GetSendMailMoney: function(): W.Data...
global GetSendMailPrice: function(): number
global GetServerExpansionLevel: function(): number
global GetServerTime: function(): number
global GetSessionTime: function(): number
global GetShapeshiftForm: function(? boolean): number
global GetShapeshiftFormCooldown: function(number): number, number, number
global GetShapeshiftFormID: function(): number
global GetShapeshiftFormInfo: function(number): string, boolean, boolean, number
global GetSheathState: function(): number
global GetShieldBlock: function(): W.MaybeSecret<number>
global GetSoundEntryCount: function(number): number
global GetSourceLocation: function(): string
global GetSpecialization: function(? boolean, ? boolean, ? number): number
global GetSpecializationInfo: function(number, ? boolean, ? boolean, ? string, ? number, ? number): number, string, string, number, string, number, number, string, number, boolean
global GetSpecializationInfoByID: function(number, ? number): number, string, string, number, string, string, string
global GetSpecializationInfoForClassID: function(number, number, ? number): number, string, string, number, string, boolean, boolean, number, number
global GetSpecializationInfoForSpecID: function(number, ? number): number, string, string, number, string, boolean, boolean, number, number
global GetSpecializationMasterySpells: function(number, ? boolean, ? boolean): number, number
global GetSpecializationNameForSpecID: function(number, ? number): string
global GetSpecializationRole: function(number, ? boolean, ? boolean): string
global GetSpecializationRoleByID: function(number): string
global GetSpecializationRoleEnum: function(): W.Data...
global GetSpecializationRoleEnumByID: function(): W.Data...
global GetSpecializationSpells: function(number, ? boolean, ? boolean): number, number, number...
global GetSpecializationSystem: function(): number
global GetSpecsForSpell: function(): W.Data...
global GetSpeed: function(): W.MaybeSecret<number>
global GetSpellBaseCooldown: function(number): number, number
global GetSpellBonusDamage: function(number): W.MaybeSecret<number>
global GetSpellBonusHealing: function(): W.MaybeSecret<number>
global GetSpellConfirmationPromptsInfo: function(): W.Data...
global GetSpellCooldown: W.Global_GetSpellCooldown
global GetSpellCritChance: function(): W.MaybeSecret<number>
global GetSpellHitModifier: function(): W.MaybeSecret<number>
global GetSpellPenetration: function(): W.MaybeSecret<number>
global GetSpellQueueWindow: function(): number
global GetSpellsForCharacterUpgradeTier: function(any): W.Data...
global GetStatistic: function(number, ? number): string, boolean, string
global GetStatisticsCategoryList: function(): W.Data
global GetStringFromModifiers: function(number): string
global GetSturdiness: function(): W.MaybeSecret<number>
global GetSubZoneText: function(): string
global GetSuggestedGroupSize: function(): W.Data...
global GetTabardCreationCost: function(): W.Data...
global GetTabardInfo: function(): W.Data...
global GetTalentInfo: function(number, number, number, ? boolean, ? string): number, string, number, boolean, boolean, number, boolean, number, number, boolean, boolean
global GetTalentInfoByID: function(any, any, any, any): W.Data...
global GetTalentInfoBySpecialization: function(any, any, any): W.Data...
global GetTalentLink: function(any, any, any, any, any): W.Data...
global GetTalentTierInfo: function(number, number, ? boolean, ? string): boolean, number, number
global GetTargetTradeCurrency: function(): W.Data...
global GetTargetTradeMoney: function(): number
global GetTaskInfo: function(number): boolean, boolean, number
global GetTaskPOIs: function(): W.Data...
global GetTasksTable: function(): W.Data...
global GetTaxiBenchmarkMode: function(): boolean
global GetTaxiMapID: function(): number
global GetTexCoordsByGrid: function(number, number, number, number, number, number): number, number, number, number
global GetTexCoordsForRole: function(string): number, number, number, number
global GetText: function(string, number, any): string
global GetTextureInfo: function(W.Texture): string, string, number, number, number, number, number, number, number, number
global GetThreatStatusColor: function(number): number, number, number
global GetTickTime: function(): number
global GetTime: function(): number
global GetTimePreciseSec: function(): number
global GetTitleName: function(number): string, boolean
global GetTitleText: function(): string
global GetTotalAchievementPoints: function(? boolean): number
global GetTotemCannotDismiss: function(number): boolean
global GetTotemDuration: function(number): W.DurationObject
global GetTotemInfo: function(number): W.MaybeSecret<boolean>, W.MaybeSecret<string>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global GetTotemTimeLeft: function(number): W.MaybeSecret<number>
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
global GetTrainerServiceNumAbilityReq: function(): W.Data...
global GetTrainerServiceSkillLine: function(number): string
global GetTrainerServiceSkillReq: function(number): string, number, boolean
global GetTrainerServiceStepIndex: function(): W.Data...
global GetTrainerServiceTypeFilter: function(string): boolean
global GetTrainerTradeskillRankValues: function(): W.Data...
global GetTreasurePickerItemInfo: function(any, any): W.Data...
global GetTutorialsEnabled: function(): W.Data...
global GetUICameraInfo: function(number): number, number, number, number, number, number, number, number, number, boolean
global GetUnitChargedPowerPoints: function(string): W.Secret<W.SecretTable>
global GetUnitEmpowerHoldAtMaxTime: function(string): W.MaybeSecret<number>
global GetUnitEmpowerMinHoldTime: function(string): W.MaybeSecret<number>
global GetUnitEmpowerStageDuration: function(string, number): W.MaybeSecret<number>
global GetUnitHealthModifier: function(string): number
global GetUnitMaxHealthModifier: function(string): number
global GetUnitPowerBarInfo: function(string): W.UnitPowerBarInfo
global GetUnitPowerBarInfoByID: function(number): W.UnitPowerBarInfo
global GetUnitPowerBarStrings: function(string): string, string, string
global GetUnitPowerBarStringsByID: function(number): string, string, string
global GetUnitPowerBarTextureInfo: function(string, number, ? number): number, number, number, number, number
global GetUnitPowerBarTextureInfoByID: function(number, number): number, number, number, number, number
global GetUnitPowerModifier: function(string): number
global GetUnitSpeed: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global GetUnitTotalModifiedMaxHealthPercent: function(string): number
global GetUpgradeExpansionLevel: function(): number
global GetVehicleUIIndicator: function(number): number, number
global GetVehicleUIIndicatorSeat: function(number, number): number, number, number
global GetVersatilityBonus: function(number): W.MaybeSecret<number>
global GetVideoCaps: function(): W.Data...
global GetWarGameQueueStatus: function(): W.Data...
global GetWarGameTypeInfo: function(any): W.Data...
global GetWeaponEnchantInfo: function(): boolean, number, number, number, boolean, number, number, number, boolean, number, number, number
global GetWebTicket: function(): W.Data...
global GetWorldElapsedTime: function(number): string, number, number
global GetWorldElapsedTimers: function(): W.Data...
global GetWorldMapActionButtonSpellInfo: function(): W.Data...
global GetWorldPVPQueueStatus: function(number): string, string, number, number, number, number, boolean
global GetXPExhaustion: function(): number
global GetZonePVPInfo: function(): string, boolean, string
global GetZoneText: function(): string
global GiveMasterLoot: function(number, number): W.Data...
global GroupHasOfflineMember: function(): W.Data...
global GuildControlAddRank: function(any): W.Data...
global GuildControlDelRank: function(number): W.Data...
global GuildControlGetAllowedShifts: function(any): W.Data...
global GuildControlGetNumRanks: function(): W.Data...
global GuildControlGetRank: function(): W.Data...
global GuildControlGetRankName: function(number): string
global GuildControlSaveRank: function(string): W.Data...
global GuildControlSetRank: function(number): W.Data...
global GuildControlSetRankFlag: function(number, boolean): W.Data...
global GuildControlShiftRankDown: function(any): W.Data...
global GuildControlShiftRankUp: function(any): W.Data...
global GuildDemote: function(string)
global GuildDisband: function()
global GuildInfo: function(): W.Data...
global GuildInvite: function(string)
global GuildLeave: function()
global GuildMasterAbsent: function(): W.Data...
global GuildNewsSetSticky: function(any, any): W.Data...
global GuildNewsSort: function(any): W.Data...
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
global HandleAtlasMemberCommand: function(): W.Data...
global HasAPEffectsSpellPower: function(): boolean
global HasArtifactEquipped: function(): W.Data...
global HasAttachedGlyph: function(any): W.Data...
global HasCompletedAnyAchievement: function(any): W.Data...
global HasDualWieldPenalty: function(): boolean
global HasFullControl: function(): boolean
global HasIgnoreDualWieldWeapon: function(): boolean
global HasInboxItem: function(): W.Data...
global HasKey: function(): boolean
global HasLFGRestrictions: function(): boolean
global HasLoadedCUFProfiles: function(): W.Data...
global HasLootSpecializations: function(): boolean
global HasNewMail: function(): W.Data...
global HasNoReleaseAura: function(): boolean, number, boolean
global HasPendingGlyphCast: function(): W.Data...
global HasPetUI: function(): boolean, boolean
global HasSPEffectsAttackPower: function(): boolean
global HasSendMailItem: function(any): W.Data...
global HasTimePassed: function(number, number): boolean
global HasWandEquipped: function(): boolean
global HaveQuestData: function(any): W.Data...
global HaveQuestRewardData: function(any): W.Data...
global HearthAndResurrectFromArea: function(): W.Data...
global HideRepairCursor: function(): W.Data...
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
global IsActiveQuestTrivial: function(any): W.Data...
global IsAdvancedFlyableArea: function(): boolean
global IsAllowedToUserTeleport: function(): boolean
global IsAltKeyDown: function(): boolean
global IsArenaSkirmish: function(): W.Data...
global IsArtifactCompletionHistoryAvailable: function(): boolean
global IsArtifactPowerItem: function(number | string): boolean
global IsAvailableQuestTrivial: function(any): W.Data...
global IsBNLogin: function(): W.Data...
global IsBetaBuild: function(): boolean
global IsBindingForGamePad: function(any): W.Data...
global IsBreadcrumbQuest: function(any): W.Data...
global IsCastingGlyph: function(): W.Data...
global IsCemeterySelectionAvailable: function(): boolean
global IsCharacterNewlyBoosted: function(): boolean
global IsChatAFK: function(): W.Data...
global IsChatChannelRaid: function(): W.Data...
global IsChatDND: function(): W.Data...
global IsCollisionEnabled: function(): boolean
global IsConsumableItem: function(number | string): boolean
global IsControlKeyDown: function(): boolean
global IsCorruptedItem: function(number | string): boolean
global IsCosmeticItem: function(number | string): boolean
global IsCpuBound: function(): boolean
global IsCurrentItem: function(number | string): boolean
global IsCurrentQuestFailed: function(): W.Data...
global IsDebugBuild: function(): boolean
global IsDemonHunterAvailable: function(): boolean
global IsDisplayChannelModerator: function(): W.Data...
global IsDisplayChannelOwner: function(): W.Data...
global IsDressableItem: function(number | string): boolean
global IsDrivableArea: function(): boolean
global IsDualWielding: function(): boolean
global IsEquippableItem: function(number | string): boolean
global IsEquippedItem: function(number | string): boolean
global IsEquippedItemType: function(string): boolean
global IsEuropeanNumbers: function(): boolean
global IsEveryoneAssistant: function(): W.Data...
global IsExpansionTrial: function(): boolean
global IsFalling: function(? string): boolean
global IsFishingLoot: function(): boolean
global IsFlyableArea: function(): boolean
global IsFlying: function(? string): boolean
global IsGMClient: function(): boolean
global IsGUIDInGroup: function(string, ? number): boolean
global IsGamePadCursorControlEnabled: function(): W.Data...
global IsGamePadFreelookEnabled: function(): W.Data...
global IsGraphicsCVarValueSupported: function(): W.Data...
global IsGraphicsSettingValueSupported: function(): W.Data...
global IsGuildLeader: function(): boolean
global IsGuildMember: function(string): boolean
global IsGuildRankAssignmentAllowed: function(any, any): W.Data...
global IsHarmfulItem: function(number | string): boolean
global IsHelpfulItem: function(number | string): boolean
global IsInActiveWorldPVP: function(): W.Data...
global IsInAuthenticatedRank: function(): W.Data...
global IsInCinematicScene: function(): boolean
global IsInGlobalEnvironment: function(): W.Data...
global IsInGroup: function(? number): boolean
global IsInGuild: function(): boolean
global IsInGuildGroup: function(): boolean
global IsInInstance: function(): boolean, string
global IsInJailersTower: function(): boolean
global IsInLFGDungeon: function(): boolean
global IsInRaid: function(? number): boolean
global IsInScenarioGroup: function(): W.Data...
global IsIndoors: function(): boolean
global IsInsane: function(): boolean
global IsInventoryItemLocked: function(number): boolean
global IsInventoryItemProfessionBag: function(any, any): W.Data...
global IsItemInRange: function(number | string, string): boolean
global IsItemPreferredArmorType: function(W.ItemLocation): boolean
global IsJailersTowerLayerTimeLocked: function(number): string
global IsKeyDown: function(string, ? boolean): boolean
global IsLFGComplete: function(): boolean
global IsLFGDungeonJoinable: function(number): boolean, boolean, boolean, number
global IsLeftAltKeyDown: function(): boolean
global IsLeftControlKeyDown: function(): boolean
global IsLeftMetaKeyDown: function(): boolean
global IsLeftShiftKeyDown: function(): boolean
global IsLegacyDifficulty: function(number): boolean
global IsLinuxClient: function(): boolean
global IsLoggedIn: function(): boolean
global IsMacClient: function(): boolean
global IsMasterLooter: function(): W.Data...
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
global IsOutlineModeSupported: function(): W.Data...
global IsPVPTimerRunning: function(): boolean
global IsPartyLFG: function(): W.Data...
global IsPartyWorldPVP: function(): W.Data...
global IsPendingGlyphRemoval: function(): W.Data...
global IsPetActive: function(): W.Data...
global IsPetAttackAction: function(any): W.Data...
global IsPetAttackActive: function(): boolean
global IsPlayerInGuildFromGUID: function(string): boolean
global IsPlayerInWorld: function(): boolean
global IsPlayerMoving: function(): boolean
global IsPlayerNeutral: function(): boolean
global IsPlayerSpell: function(number): boolean
global IsPublicBuild: function(): boolean
global IsPublicTestClient: function(): boolean
global IsQuestCompletable: function(): boolean
global IsQuestIDValidSpellTarget: function(any): W.Data...
global IsQuestItemHidden: function(any): W.Data...
global IsQuestLogSpecialItemInRange: function(any, any): W.Data...
global IsQuestSequenced: function(any): W.Data...
global IsRaidMarkerActive: function(number): W.MaybeSecret<boolean>
global IsRaidMarkerSystemEnabled: function(): boolean
global IsRangedWeapon: function(): boolean
global IsResting: function(): boolean
global IsRestrictedAccount: function(): boolean
global IsRightAltKeyDown: function(): boolean
global IsRightControlKeyDown: function(): boolean
global IsRightMetaKeyDown: function(): boolean
global IsRightShiftKeyDown: function(): boolean
global IsSelectedSpellBookItem: function(any): W.Data...
global IsServerControlledBackfill: function(): W.Data...
global IsShiftKeyDown: function(): boolean
global IsSpecializationActivateSpell: function(): W.Data...
global IsSpellClassOrSpec: W.Global_IsSpellClassOrSpec
global IsSpellKnown: function(number, boolean): boolean
global IsSpellKnownOrOverridesKnown: function(number, boolean): boolean
global IsSpellOverlayed: function(number): boolean
global IsSpellValidForPendingGlyph: function(any): W.Data...
global IsStealthed: function(): boolean
global IsStoryQuest: function(any): W.Data...
global IsSubZonePVPPOI: function(): boolean
global IsSubmerged: function(? string): boolean
global IsSwimming: function(? string): boolean
global IsTargetLoose: function(): boolean
global IsTestBuild: function(): boolean
global IsThreatWarningEnabled: function(): boolean
global IsTitleKnown: function(number): boolean
global IsTradeskillTrainer: function(): boolean
global IsTrialAccount: function(): boolean
global IsTutorialFlagged: function(any): W.Data...
global IsUnitModelReadyForUI: function(string): boolean
global IsUsableItem: function(number | string): boolean, boolean
global IsUsingFixedTimeStep: function(): boolean
global IsUsingGamepad: function(): boolean
global IsUsingMouse: function(): boolean
global IsUsingVehicleControls: function(): W.Data...
global IsVehicleAimAngleAdjustable: function(): boolean
global IsVehicleAimPowerAdjustable: function(): boolean
global IsVeteranTrialAccount: function(): boolean
global IsWargame: function(): boolean
global IsWindowsClient: function(): boolean
global IsXPUserDisabled: function(): boolean
global Item: W.Global_Item
global ItemAddedToArtifact: function(number): boolean
global ItemCanTargetGarrisonFollowerAbility: function(): W.Data...
global ItemHasRange: function(number | string): boolean
global ItemLocation: W.Global_ItemLocation
global ItemLocationMixin: W.Global_ItemLocationMixin
global ItemMixin: W.Global_ItemMixin
global ItemTextFontNormal: W.Global_ItemTextFontNormal
global ItemTextGetCreator: function(): string
global ItemTextGetItem: function(): string
global ItemTextGetMaterial: function(): string
global ItemTextGetPage: function(): number
global ItemTextGetText: function(): string
global ItemTextHasNextPage: function(): boolean
global ItemTextIsFullPage: function(): W.Data...
global ItemTextNextPage: function(): W.Data...
global ItemTextPrevPage: function(): W.Data...
global ItemTransmogInfoMixin: W.Global_ItemTransmogInfoMixin
global JOURNAL_LINK_COLOR: W.Global_JOURNAL_LINK_COLOR
global JoinArena: function(): W.Data...
global JoinChannelByName: function(string, ? string, ? number, boolean): number, string
global JoinLFG: function(any): W.Data...
global JoinPermanentChannel: function(string, ? string, ? number, ? boolean): number, string
global JoinRatedBattlefield: function(): W.Data...
global JoinRatedSoloShuffle: function(): W.Data...
global JoinSingleLFG: function(any, any): W.Data...
global JoinSkirmish: function(number, ? boolean): W.Data...
global JoinTemporaryChannel: function(string, ? string, ? number, ? boolean): number, string
global JumpOrAscendStart: function(): W.Data...
global KYRIAN_BLUE_COLOR: W.Global_KYRIAN_BLUE_COLOR
global LEGENDARY_ORANGE_COLOR: W.Global_LEGENDARY_ORANGE_COLOR
global LFGTeleport: function(boolean): W.Data...
global LIGHTBLUE_FONT_COLOR: W.Global_LIGHTBLUE_FONT_COLOR
global LIGHTERBLUE_FONT_COLOR: W.Global_LIGHTERBLUE_FONT_COLOR
global LIGHTERGRAY_FONT_COLOR: W.Global_LIGHTERGRAY_FONT_COLOR
global LIGHTGRAY_FONT_COLOR: W.Global_LIGHTGRAY_FONT_COLOR
global LIGHTYELLOW_FONT_COLOR: W.Global_LIGHTYELLOW_FONT_COLOR
global LINK_FONT_COLOR: W.Global_LINK_FONT_COLOR
global LOOT_LINK_COLOR: W.Global_LOOT_LINK_COLOR
global LORE_TEXT_BODY_COLOR: W.Global_LORE_TEXT_BODY_COLOR
global LaunchURL: function(string)
global LearnPvpTalent: function(number, number): W.Data...
global LearnPvpTalents: function(): W.Data...
global LearnTalent: function(number): boolean
global LearnTalents: function(): W.Data...
global LeaveBattlefield: function(): W.Data...
global LeaveChannelByLocalID: function(any): W.Data...
global LeaveChannelByName: function(string): W.Data...
global LeaveLFG: function(any): W.Data...
global LeaveSingleLFG: function(any, any): W.Data...
global ListChannelByName: function(number | string): W.Data...
global ListChannels: function(): W.Data...
global LoadBindings: function(number): W.Data...
global LoadURLIndex: function(number, ? number)
global LocalizedClassList: function(? boolean): W.Data
global LoggingChat: function(? boolean): boolean
global LoggingCombat: function(? boolean): boolean
global Logout: function()
global LootMoneyNotify: function(any, any): W.Data...
global LootSlot: function(number): W.Data...
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
global Mixin: function<T>(T, any...): T
global MouseOverrideCinematicDisable: function(? boolean)
global MouselookStart: function(): W.Data...
global MouselookStop: function(): W.Data...
global MoveAndSteerStart: function(): W.Data...
global MoveAndSteerStop: function(): W.Data...
global MoveBackwardStart: function(number): W.Data...
global MoveBackwardStop: function(): W.Data...
global MoveForwardStart: function(number): W.Data...
global MoveForwardStop: function(): W.Data...
global MoveViewDownStart: function(number): W.Data...
global MoveViewDownStop: function(): W.Data...
global MoveViewInStart: function(number): W.Data...
global MoveViewInStop: function(): W.Data...
global MoveViewLeftStart: function(number): W.Data...
global MoveViewLeftStop: function(): W.Data...
global MoveViewOutStart: function(number): W.Data...
global MoveViewOutStop: function(): W.Data...
global MoveViewRightStart: function(number): W.Data...
global MoveViewRightStop: function(): W.Data...
global MoveViewUpStart: function(number): W.Data...
global MoveViewUpStop: function(): W.Data...
global MovieSubtitleFont: W.Global_MovieSubtitleFont
global MultiSampleAntiAliasingSupported: function(): W.Data...
global MuteSoundFile: function(number | string): W.Data...
global NECROLORD_GREEN_COLOR: W.Global_NECROLORD_GREEN_COLOR
global NEUTRAL_STATUS_COLOR: W.Global_NEUTRAL_STATUS_COLOR
global NEW_FEATURE_SHADOW_COLOR: W.Global_NEW_FEATURE_SHADOW_COLOR
global NIFFEN_MAJOR_FACTION_COLOR: W.Global_NIFFEN_MAJOR_FACTION_COLOR
global NIGHT_FAE_BLUE_COLOR: W.Global_NIGHT_FAE_BLUE_COLOR
global NORMAL_FONT_COLOR: W.Global_NORMAL_FONT_COLOR
global NOT_ON_THREAT_COLOR: W.Global_NOT_ON_THREAT_COLOR
global NO_THREAT_COLOR: W.Global_NO_THREAT_COLOR
global NeutralPlayerSelectFaction: function(number): W.Data...
global NewSubSpellFont: W.Global_NewSubSpellFont
global NextView: function(): W.Data...
global NoPlayTime: function(): boolean
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
global OfferPetition: function(): W.Data...
global OpenTrainer: function(): W.Data...
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
global PartialPlayTime: function(): boolean
global PartyLFGStartBackfill: function(): W.Data...
global PetAggressiveMode: function(): W.Data...
global PetAttack: function(): W.Data...
global PetCanBeAbandoned: function(): boolean
global PetCanBeDismissed: function(): W.Data...
global PetDefensiveAssistMode: function(): W.Data...
global PetDefensiveMode: function(): W.Data...
global PetDismiss: function(): W.Data...
global PetFollow: function(): W.Data...
global PetHasActionBar: function(): boolean
global PetHasSpellbook: function(): W.Data...
global PetMoveTo: function(any): W.Data...
global PetPassiveMode: function(): W.Data...
global PetStopAttack: function(): W.Data...
global PetUsesPetFrame: function(): W.Data...
global PetWait: function(): W.Data...
global PickupAction: function(number): W.Data...
global PickupBagFromSlot: function(number): W.Data...
global PickupCompanion: function(string, number): W.Data...
global PickupGuildBankItem: function(any, any): W.Data...
global PickupGuildBankMoney: function(any): W.Data...
global PickupInventoryItem: function(number): W.Data...
global PickupItem: function(number | string)
global PickupMacro: function(number | string): W.Data...
global PickupMerchantItem: function(number): W.Data...
global PickupPetAction: function(number): W.Data...
global PickupPetSpell: function(number): W.Data...
global PickupPlayerMoney: function(number)
global PickupPvpTalent: function(): W.Data...
global PickupTalent: function(any): W.Data...
global PitchDownStart: function(): W.Data...
global PitchDownStop: function(): W.Data...
global PitchUpStart: function(): W.Data...
global PitchUpStop: function(): W.Data...
global PlaceAction: function(number): W.Data...
global PlaceRaidMarker: function(number, ? string)
global PlayAutoAcceptQuestSound: function(): W.Data...
global PlayMusic: function(number | string): boolean
global PlaySound: function(number, ? string, ? boolean, ? boolean): boolean, number
global PlaySoundFile: function(number | string, ? string): boolean, number
global PlayVocalErrorSoundID: function(number)
global PlayerCanTeleport: function(): boolean
global PlayerEffectiveAttackPower: function(): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global PlayerGetTimerunningSeasonID: function(): number
global PlayerHasToy: function(number): boolean
global PlayerIsInCombat: function(): boolean
global PlayerIsPVPInactive: function(string): boolean
global PlayerIsSpellTarget: function(string): W.MaybeSecret<boolean>
global PlayerIsTimerunning: function(): boolean
global PlayerLocation: W.Global_PlayerLocation
global PlayerLocationMixin: W.Global_PlayerLocationMixin
global PlayerVehicleHasComboPoints: function(): boolean
global PortGraveyard: function()
global PreloadMovie: function(number)
global PrevView: function(): W.Data...
global PriceFont: W.Global_PriceFont
global PriceFontGray: W.Global_PriceFontGray
global PriceFontGreen: W.Global_PriceFontGreen
global PriceFontRed: W.Global_PriceFontRed
global PriceFontWhite: W.Global_PriceFontWhite
global PriceFontYellow: W.Global_PriceFontYellow
global ProcessExceptionClient: function(string): W.Data...
global ProcessQuestLogRewardFactions: function(): W.Data...
global PromoteToAssistant: function(any): W.Data...
global PromoteToLeader: W.Global_PromoteToLeader
global PutItemInBackpack: function(): W.Data...
global PutItemInBag: function(? number): boolean
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
global QueryGuildBankLog: function(any): W.Data...
global QueryGuildBankTab: function(any): W.Data...
global QueryGuildBankText: function(): W.Data...
global QueryGuildEventLog: function(): W.Data...
global QueryGuildNews: function(): W.Data...
global QueryGuildRecipes: function(): W.Data...
global QuestChooseRewardError: function(): W.Data...
global QuestDifficulty_Difficult: W.Global_QuestDifficulty_Difficult
global QuestDifficulty_Header: W.Global_QuestDifficulty_Header
global QuestDifficulty_Impossible: W.Global_QuestDifficulty_Impossible
global QuestDifficulty_Standard: W.Global_QuestDifficulty_Standard
global QuestDifficulty_Trivial: W.Global_QuestDifficulty_Trivial
global QuestDifficulty_VeryDifficult: W.Global_QuestDifficulty_VeryDifficult
global QuestFlagsPVP: function(): W.Data...
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
global QuestGetAutoLaunched: function(): W.Data...
global QuestHasPOIInfo: function(any): W.Data...
global QuestIsDaily: function(): boolean
global QuestIsFromAdventureMap: function(): W.Data...
global QuestIsFromAreaTrigger: function(): boolean
global QuestIsWeekly: function(): boolean
global QuestLogPushQuest: function(): W.Data...
global QuestLogRewardHasTreasurePicker: function(): W.Data...
global QuestLogShouldShowPortrait: function(): W.Data...
global QuestMapRewardsFont: W.Global_QuestMapRewardsFont
global QuestMapUpdateAllQuests: function(): W.Data...
global QuestPOIUpdateIcons: function(): W.Data...
global QuestTitleFont: W.Global_QuestTitleFont
global QuestTitleFontBlackShadow: W.Global_QuestTitleFontBlackShadow
global Quit: function()
global RAF_VERSION_THREE_COLOR: W.Global_RAF_VERSION_THREE_COLOR
global RAF_VERSION_TWO_COLOR: W.Global_RAF_VERSION_TWO_COLOR
global RAID_CLASS_COLORS: W.Global_RAID_CLASS_COLORS
global RARE_BLUE_COLOR: W.Global_RARE_BLUE_COLOR
global RARE_MISSION_COLOR: W.Global_RARE_MISSION_COLOR
global RED_FONT_COLOR: W.Global_RED_FONT_COLOR
global RED_THREAT_COLOR: W.Global_RED_THREAT_COLOR
global REFUND_TOOLTIP_COLOR: W.Global_REFUND_TOOLTIP_COLOR
global RPE_FONT_COLOR: W.Global_RPE_FONT_COLOR
global RUNEFORGE_LEGEDARY_SPEC_COLOR: W.Global_RUNEFORGE_LEGEDARY_SPEC_COLOR
global RaidNotice_AddMessage: function(W.Frame, string, W.ChatTypeColor, ? number)
global RaidProfileExists: function(any): W.Data...
global RaidProfileHasUnsavedChanges: function(): W.Data...
global RaidWarningFrame: W.Global_RaidWarningFrame
global RandomRoll: function(number, number)
global RedockChatWindows: function(): W.Data...
global RefreshLFGList: function(): W.Data...
global RegisterCVar: function(string, ? string)
global RegisterEventCallback: function(string, function())
global RegisterStateDriver: function(W.Frame, string, string)
global RegisterStaticConstants: function(any): W.Data...
global RegisterUnitEventCallback: function(string, function(), string)
global RejectProposal: function(): W.Data...
global ReleaseAction: function(): W.Data...
global RemoveAutoQuestPopUp: function(): W.Data...
global RemoveChatWindowChannel: function(number, string): W.Data...
global RemoveChatWindowMessages: function(number, string): W.Data...
global RemoveItemFromArtifact: function(): boolean
global RemovePvpTalent: function(): W.Data...
global RemoveRaidTargets: function()
global RemoveTalent: function(any): W.Data...
global RenamePetition: function(string): W.Data...
global RepairAllItems: function(? boolean): W.Data...
global ReplaceEnchant: function()
global ReplaceGenderTokens: function(string, string): string
global ReplaceGuildMaster: function(): W.Data...
global ReplaceTradeEnchant: function()
global ReplaceTradeskillEnchant: function()
global RepopMe: function()
global ReportBug: function(string)
global ReportInfoMixin: W.Global_ReportInfoMixin
global ReportPlayerIsPVPAFK: function(string)
global ReportSuggestion: function(string)
global ReputationDetailFont: W.Global_ReputationDetailFont
global RequestBattlefieldScoreData: function(): W.Data...
global RequestBattlegroundInstanceInfo: function(number): W.Data...
global RequestBottomLeftActionBar: function(): W.Data...
global RequestGuildChallengeInfo: function(): W.Data...
global RequestGuildPartyState: function(): W.Data...
global RequestGuildRewards: function(): W.Data...
global RequestLFDPartyLockInfo: function(): W.Data...
global RequestLFDPlayerLockInfo: function(): W.Data...
global RequestPVPOptionsEnabled: function(): W.Data...
global RequestPVPRewards: function(): W.Data...
global RequestRaidInfo: function(): W.Data...
global RequestRandomBattlegroundInstanceInfo: function(): W.Data...
global RequestRatedInfo: function(): W.Data...
global RequestTimePlayed: function()
global RequeueSkirmish: function(): W.Data...
global ResetCPUUsage: function()
global ResetChatColors: function(): W.Data...
global ResetChatWindows: function(): W.Data...
global ResetCursor: function()
global ResetInstances: function()
global ResetSetMerchantFilter: function(): W.Data...
global ResetTestCvars: function()
global ResetTutorials: function(): W.Data...
global ResetView: function(any): W.Data...
global ResistancePercent: function(number, number): number
global RespondInstanceLock: function(boolean)
global RespondMailLockSendItem: function(any, any): W.Data...
global RespondToInviteConfirmation: function(any, any): W.Data...
global RestartGx: function()
global RestoreRaidProfileFromCopy: function(): W.Data...
global ResurrectGetOfferer: function(): string
global ResurrectHasSickness: function(): boolean
global ResurrectHasTimer: function(): boolean
global RetrieveCorpse: function()
global ReturnInboxItem: function(any): W.Data...
global RollOnLoot: function(number, ? number): W.Data...
global RunBinding: function(string, ? string): W.Data...
global RunMacro: function(number | string): W.Data...
global RunNextFrame: function(function(W.Data...))
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
global SafePack: function(any...): W.Data
global SafeUnpack: function(any): W.Data...
global SaveBindings: function(number): W.Data...
global SaveRaidProfileCopy: function(any): W.Data...
global SaveView: function(number): W.Data...
global Screenshot: function()
global SearchLFGGetEncounterResults: function(any, any): W.Data...
global SearchLFGGetJoinedID: function(): W.Data...
global SearchLFGGetNumResults: function(): number, number
global SearchLFGGetPartyResults: function(number, number): W.Data...
global SearchLFGGetResults: function(number): W.Data...
global SearchLFGJoin: function(number, number): W.Data...
global SearchLFGLeave: function(): W.Data...
global SearchLFGSort: function(any): W.Data...
global SecondsToClock: function(number, ? boolean): string
global SecondsToMinutes: function(number): number
global SecondsToTime: function(number, ? boolean, ? boolean, ? number, ? boolean): string
global SecondsToTimeAbbrev: function(number): string, number
global SecureCmdOptionParse: function(string): string, string
global SecureHandlerSetFrameRef: function(W.Frame, string, W.Frame)
global SecureHandlerWrapScript: function(W.Frame, string, W.Frame, string, ? string)
global SelectActiveQuest: function(): W.Data...
global SelectAvailableQuest: function(): W.Data...
global SelectTrainerService: function(number): W.Data...
global SelectedRealmName: function(): string
global SellCursorItem: function()
global SendChatMessage: function(string, ? W.Literal_62300f6ce66d, ? number, ? string)
global SendMail: function(string, string, ? string): W.Data...
global SendSubscriptionInterstitialResponse: function(number)
global SendSystemMessage: function(string): W.Data...
global SetAchievementComparisonUnit: function(string): boolean
global SetAchievementSearchString: function(string): boolean
global SetActionBarToggles: function(boolean, boolean, boolean, boolean, boolean, boolean, boolean, string): W.Data...
global SetAllowDangerousScripts: function(? boolean)
global SetAllowLowLevelRaid: function(? boolean)
global SetAllowRecentAlliesSeeLocation: function(boolean)
global SetAutoDeclineGuildInvites: function(? boolean)
global SetAutoDeclineNeighborhoodInvites: function(? boolean)
global SetBarSlotFromIntro: function(any): W.Data...
global SetBattlefieldScoreFaction: function(number): W.Data...
global SetBinding: function(string, ? string, ? number): boolean
global SetBindingClick: function(string, string, string): boolean
global SetBindingItem: function(string, string): boolean
global SetBindingMacro: function(string, number | string): boolean
global SetBindingSpell: function(string, string): boolean
global SetCVar: function(string, ? boolean | string | number): boolean
global SetCVarBitfield: function(string, number, boolean): boolean
global SetCemeteryPreference: function(number)
global SetChannelOwner: function(string, string): W.Data...
global SetChannelPassword: function(string, string): W.Data...
global SetChatColorNameByClass: function(any, any): W.Data...
global SetChatWindowAlpha: function(any, any): W.Data...
global SetChatWindowColor: function(any, any, any, any): W.Data...
global SetChatWindowDocked: function(any, any): W.Data...
global SetChatWindowLocked: function(any, any): W.Data...
global SetChatWindowName: function(any, any): W.Data...
global SetChatWindowSavedDimensions: function(any, any, any): W.Data...
global SetChatWindowSavedPosition: function(any, any, any, any): W.Data...
global SetChatWindowShown: function(any, any): W.Data...
global SetChatWindowSize: function(any, any): W.Data...
global SetChatWindowUninteractable: function(any, any): W.Data...
global SetClampedTextureRotation: function(W.Texture, number)
global SetConsoleKey: function(string)
global SetCurrentGraphicsSetting: function(): W.Data...
global SetCurrentGuildBankTab: function(any): W.Data...
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
global SetEveryoneIsAssistant: function(): W.Data...
global SetFocusedAchievement: function(? number): W.Data...
global SetGamePadCursorControl: function(): W.Data...
global SetGamePadFreeLook: function(): W.Data...
global SetGuildBankTabInfo: function(number, string, number): W.Data...
global SetGuildBankTabItemWithdraw: function(): W.Data...
global SetGuildBankTabPermissions: function(number, number, boolean): W.Data...
global SetGuildBankText: function(number, string): W.Data...
global SetGuildBankWithdrawGoldLimit: function(number): W.Data...
global SetGuildMemberRank: function(any, any): W.Data...
global SetGuildNewsFilter: function(any, any): W.Data...
global SetGuildRosterSelection: function(number): W.Data...
global SetGuildRosterShowOffline: function(boolean): W.Data...
global SetGuildTradeSkillCategoryFilter: function(any): W.Data...
global SetGuildTradeSkillItemNameFilter: function(any): W.Data...
global SetInWorldUIVisibility: function(boolean)
global SetLFGBootVote: function(any): W.Data...
global SetLFGComment: function(string): W.Data...
global SetLFGDungeon: function(any, any): W.Data...
global SetLFGDungeonEnabled: function(any, any): W.Data...
global SetLFGHeaderCollapsed: function(any, any): W.Data...
global SetLFGRoles: function(any, any, any, any): W.Data...
global SetLegacyRaidDifficultyID: function(number, ? boolean)
global SetLootPortrait: function(): W.Data...
global SetLootSpecialization: function(number)
global SetLootThreshold: function(number): W.Data...
global SetMacroItem: function(any, any, any): W.Data...
global SetMacroSpell: function(number | string, string, string): W.Data...
global SetMerchantFilter: function(): W.Data...
global SetModifiedClick: function(string, string): W.Data...
global SetMouselookOverrideBinding: function(any, any): W.Data...
global SetMoveEnabled: function(): W.Data...
global SetMultiCastSpell: function(number, number): W.Data...
global SetOptOutOfLoot: function(boolean): W.Data...
global SetOverrideBinding: function(W.Frame, boolean, string, string): W.Data...
global SetOverrideBindingClick: function(W.Frame, boolean, string, string, ? string): W.Data...
global SetOverrideBindingItem: function(W.Frame, boolean, string, string): W.Data...
global SetOverrideBindingMacro: function(W.Frame, boolean, string, string): W.Data...
global SetOverrideBindingSpell: function(W.Frame, boolean, string, string): W.Data...
global SetPOIIconOverlapDistance: function(any): W.Data...
global SetPOIIconOverlapPushDistance: function(any): W.Data...
global SetPVP: function(? boolean)
global SetPVPRoles: function(boolean, boolean, boolean): W.Data...
global SetPartyAssignment: function(any, any): W.Data...
global SetPortraitTexture: function(W.Texture, string, ? boolean)
global SetPortraitTextureFromCreatureDisplayID: function(W.Texture, number)
global SetPortraitToTexture: function(W.Texture, number | string)
global SetRaidDifficultyID: function(number, ? boolean)
global SetRaidProfileOption: function(any, any, any): W.Data...
global SetRaidProfileSavedPosition: function(any, any, any, any, any, any, any, any): W.Data...
global SetRaidSubgroup: function(number, number): W.Data...
global SetRaidTarget: function(string, number)
global SetSavedInstanceExtend: function(any, any): W.Data...
global SetSelectedArtifact: function(number, ? number): W.Data...
global SetSelectedDisplayChannel: function(any): W.Data...
global SetSelectedScreenResolutionIndex: function(any): W.Data...
global SetSelectedWarGameType: function(any): W.Data...
global SetSendMailCOD: function(any): W.Data...
global SetSendMailMoney: function(any): W.Data...
global SetSendMailShowing: function(boolean): W.Data...
global SetSpellbookPetAction: function(any, any): W.Data...
global SetTableSecurityOption: function(any, number)
global SetTaxiBenchmarkMode: function(? boolean)
global SetTaxiMap: function(string): W.Data...
global SetTradeCurrency: function(any, any): W.Data...
global SetTrainerServiceTypeFilter: function(string, boolean, boolean): W.Data...
global SetTurnEnabled: function(): W.Data...
global SetUIVisibility: function(boolean)
global SetUnitCursorTexture: function(W.Texture, string, ? number, ? boolean): boolean
global SetView: function(number): W.Data...
global Settings: W.Global_Settings
global SetupAtlasesOnRegions: function(W.Frame, any, ? boolean)
global SetupFullscreenScale: function(W.Frame): W.Data...
global SetupTextureKitOnFrame: W.Global_SetupTextureKitOnFrame
global SetupTextureKitOnFrames: function(any, any, boolean, ? boolean)
global SetupTextureKitOnRegions: function(any, W.Frame, any, boolean, ? boolean)
global SetupTextureKitsFromRegionInfo: function(any, W.Frame, any)
global ShouldShowExpansionUpgradeBanner: function(): boolean
global ShouldShowIslandsWeeklyPOI: function(): boolean
global ShouldShowSpecialSplashScreen: function(): boolean
global ShowBossFrameWhenUninteractable: function(? string): boolean
global ShowBuybackSellCursor: function(any): W.Data...
global ShowCloak: function(boolean)
global ShowHelm: function(boolean)
global ShowQuestComplete: function(number): W.Data...
global ShowQuestOffer: function(any): W.Data...
global ShowRepairCursor: function(): W.Data...
global ShowUIPanel: function(W.Frame, ? number)
global ShowingCloak: function(): boolean
global ShowingHelm: function(): boolean
global SignPetition: function(): W.Data...
global SimulateMouseClick: function(W.Literal_b7e0de2fb994)
global SimulateMouseDown: function(W.Literal_b7e0de2fb994)
global SimulateMouseUp: function(W.Literal_b7e0de2fb994)
global SimulateMouseWheel: function(number)
global SitStandOrDescendStart: function()
global SlashCmdList: W.Global_SlashCmdList
global SocketInventoryItem: function(any): W.Data...
global SocketItemToArtifact: function(): boolean
global SolveArtifact: function(): W.Data...
global SortBGList: function(): boolean
global SortBattlefieldScoreData: function(any): W.Data...
global SortGuildRoster: function(string): W.Data...
global SortGuildTradeSkill: function(any): W.Data...
global SortQuestSortTypes: function(): W.Data...
global SortQuests: function(): W.Data...
global Sound_ChatSystem_GetInputDriverNameByIndex: function(any): W.Data...
global Sound_ChatSystem_GetNumInputDrivers: function(): W.Data...
global Sound_ChatSystem_GetNumOutputDrivers: function(): W.Data...
global Sound_ChatSystem_GetOutputDriverNameByIndex: function(any): W.Data...
global Sound_GameSystem_GetInputDriverNameByIndex: function(any): W.Data...
global Sound_GameSystem_GetNumInputDrivers: function(): W.Data...
global Sound_GameSystem_GetNumOutputDrivers: function(): W.Data...
global Sound_GameSystem_GetOutputDriverNameByIndex: function(any): W.Data...
global Sound_GameSystem_RestartSoundSystem: function(): W.Data...
global Spell: W.Global_Spell
global SpellCanTargetGarrisonFollower: function(): W.Data...
global SpellCanTargetGarrisonFollowerAbility: function(): W.Data...
global SpellCanTargetGarrisonMission: function(): W.Data...
global SpellCanTargetItem: function(): W.Data...
global SpellCanTargetItemID: function(): W.Data...
global SpellCanTargetQuest: function(): W.Data...
global SpellCanTargetUnit: function(string): boolean
global SpellCancelQueuedSpell: function(): W.Data...
global SpellFont_Small: W.Global_SpellFont_Small
global SpellIsTargeting: function(): boolean
global SpellMixin: W.Global_SpellMixin
global SpellStopCasting: function(): boolean
global SpellStopTargeting: function(): W.Data...
global SpellTargetItem: function(any): W.Data...
global SpellTargetUnit: function(string): W.Data...
global SplashFrameCanBeShown: function(): boolean
global SplashHeaderFont: W.Global_SplashHeaderFont
global SplitGuildBankItem: function(any, any, any): W.Data...
global SplitTextIntoHeaderAndNonHeader: function(string): string, string
global SplitTextIntoLines: function(string, string): {string}
global StartAttack: function(? string, ? boolean)
global StartAutoRun: function(): W.Data...
global StartDuel: W.Global_StartDuel
global StartSoloShuffleWarGameByName: function(): W.Data...
global StartSpectatorSoloShuffleWarGame: function(): W.Data...
global StartSpectatorWarGame: function(any, any, any, any, any): W.Data...
global StartWarGame: function(any, any, any): W.Data...
global StartWarGameByName: function(any): W.Data...
global StopAttack: function()
global StopAutoRun: function(): W.Data...
global StopCinematic: function()
global StopMacro: function(): W.Data...
global StopMusic: function(): W.Data...
global StopSound: function(number, ? number): W.Data...
global StoreSecureReference: function(any, any): W.Data...
global StrafeLeftStart: function(): W.Data...
global StrafeLeftStop: function(): W.Data...
global StrafeRightStart: function(): W.Data...
global StrafeRightStop: function(): W.Data...
global Stuck: function()
global SubSpellFont: W.Global_SubSpellFont
global SubZoneTextFont: W.Global_SubZoneTextFont
global SubmitRequiredGuildRename: function(): W.Data...
global SummonRandomCritter: function(): W.Data...
global SupportsClipCursor: function(): boolean
global SurrenderArena: function(): W.Data...
global SwapRaidSubgroup: function(number, number): W.Data...
global SwapToGlobalEnvironment: function(): W.Data...
global SwitchAchievementSearchTab: function(any): W.Data...
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
global TakeInboxItem: function(number, number): W.Data...
global TakeInboxMoney: function(number): W.Data...
global TakeInboxTextItem: function(any): W.Data...
global TakeTaxiNode: function(number): W.Data...
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
global TaxiGetNodeSlot: function(): W.Data...
global TaxiGetSrcX: function(? number, ? number): number
global TaxiGetSrcY: function(? number, ? number): number
global TaxiIsDirectFlight: function(): W.Data...
global TaxiNodeCost: function(number): number
global TaxiNodeGetType: function(number): string
global TaxiNodeName: function(number): string
global TaxiNodePosition: function(number): number, number
global TaxiRequestEarlyLanding: function(): W.Data...
global TextStatusBarText: W.Global_TextStatusBarText
global TextStatusBarTextLarge: W.Global_TextStatusBarTextLarge
global TimeoutResurrect: function()
global ToggleAnimKitDisplay: function()
global ToggleAutoRun: function(): W.Data...
global ToggleCollision: function()
global ToggleCollisionDisplay: function()
global ToggleDebugAIDisplay: function()
global ToggleGravity: function()
global TogglePVP: function()
global TogglePetAutocast: function(any): W.Data...
global TogglePlayerBounds: function()
global TogglePortals: function()
global ToggleRun: function(): W.Data...
global ToggleSelfHighlight: function(): boolean
global ToggleSheath: function()
global ToggleTris: function()
global ToggleWindowed: function(): W.Data...
global Tooltip_Med: W.Global_Tooltip_Med
global Tooltip_Small: W.Global_Tooltip_Small
global TransmogLocationMixin: W.Global_TransmogLocationMixin
global TransmogPendingInfoMixin: W.Global_TransmogPendingInfoMixin
global TriggerTutorial: function(any): W.Data...
global TurnInGuildCharter: function(): W.Data...
global TurnLeftStart: function(number): W.Data...
global TurnLeftStop: function(): W.Data...
global TurnOrActionStart: function(): W.Data...
global TurnOrActionStop: function(): W.Data...
global TurnRightStart: function(number): W.Data...
global TurnRightStop: function(): W.Data...
global UIFrameFade: function(W.Frame, any)
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
global UninviteUnit: function(string, ? string): W.Data...
global UnitAffectingCombat: function(string): boolean
global UnitAlliedRaceInfo: function(string): boolean, boolean
global UnitArmor: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitAttackPower: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitAttackSpeed: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitBattlePetLevel: function(string): number
global UnitBattlePetSpeciesID: function(string): number
global UnitBattlePetType: function(string): number
global UnitCanAssist: function(string, string, ? boolean, ? boolean): boolean
global UnitCanAttack: function(string, string): boolean
global UnitCanCooperate: function(string, string): boolean
global UnitCanPetBattle: function(string, string): boolean
global UnitCastingDuration: function(string): W.Secret<W.DurationObject>
global UnitCastingInfo: function(string): W.MaybeSecret<string>, W.MaybeSecret<string>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, boolean, W.MaybeSecret<string>, W.MaybeSecret<boolean>, W.MaybeSecret<number>, number, number
global UnitChannelDuration: function(string): W.DurationObject
global UnitChannelInfo: function(string): W.MaybeSecret<string>, W.MaybeSecret<string>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, boolean, W.MaybeSecret<boolean>, W.MaybeSecret<number>, boolean, number, number
global UnitChromieTimeID: function(string): number
global UnitClass: function(string): W.MaybeSecret<string>, W.MaybeSecret<string>, W.MaybeSecret<number>
global UnitClassBase: function(string): W.MaybeSecret<string>, W.MaybeSecret<number>
global UnitClassFromGUID: function(W.MaybeSecret<string>): W.MaybeSecret<string>, string, number
global UnitClassification: function(string): string
global UnitControllingVehicle: function(string): boolean
global UnitCreatureFamily: function(string): W.MaybeSecret<string>, W.MaybeSecret<number>
global UnitCreatureID: function(string): W.MaybeSecret<number>
global UnitCreatureType: function(string): W.MaybeSecret<string>, W.MaybeSecret<number>
global UnitDamage: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitDetailedThreatSituation: function(string, string): W.MaybeSecret<boolean>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitDistanceSquared: function(string): number, boolean
global UnitEffectiveLevel: function(string): number
global UnitEmpoweredChannelDuration: function(string, ? boolean): W.DurationObject
global UnitEmpoweredStageDurations: function(string): {W.DurationObject}
global UnitEmpoweredStagePercentages: function(string, ? boolean): {number}
global UnitExists: function(? string): boolean
global UnitFactionGroup: function(string, ? boolean): string, string
global UnitFullName: function(string): W.MaybeSecret<string>, W.MaybeSecret<string>
global UnitGUID: function(string): W.MaybeSecret<string>
global UnitGetAvailableRoles: function(string): W.MaybeSecret<boolean>, W.MaybeSecret<boolean>, W.MaybeSecret<boolean>
global UnitGetDetailedHealPrediction: function(string, ? string, W.UnitHealPredictionCalculator)
global UnitGetIncomingHeals: function(string, ? string): W.MaybeSecret<number>
global UnitGetTotalAbsorbs: function(string): W.MaybeSecret<number>
global UnitGetTotalHealAbsorbs: function(string): W.MaybeSecret<number>
global UnitGroupRolesAssigned: function(? string): W.MaybeSecret<string>
global UnitGroupRolesAssignedEnum: function(? string): W.MaybeSecret<number>
global UnitHPPerStamina: function(string): number
global UnitHasIncomingResurrection: function(string): boolean
global UnitHasLFGDeserter: function(string): boolean
global UnitHasLFGRandomCooldown: function(string): boolean
global UnitHasPowerType: function(string, number): boolean
global UnitHasRelicSlot: function(string): boolean
global UnitHasVehiclePlayerFrameUI: function(? string): boolean
global UnitHasVehicleUI: function(? string): boolean
global UnitHealth: function(string, ? boolean): W.MaybeSecret<number>
global UnitHealthMax: function(string): W.MaybeSecret<number>
global UnitHealthMissing: function(string, ? boolean): W.MaybeSecret<number>
global UnitHealthPercent: W.Global_UnitHealthPercent
global UnitHonor: function(string): number
global UnitHonorLevel: function(string): W.MaybeSecret<number>
global UnitHonorMax: function(string): number
global UnitInAnyGroup: function(? string, ? number): boolean
global UnitInBattleground: function(? string, ? number): number
global UnitInOtherParty: function(string): boolean
global UnitInParty: function(? string, ? number): boolean
global UnitInPartyIsAI: function(? string): boolean
global UnitInPartyShard: function(string): boolean
global UnitInRaid: function(? string, ? number): W.MaybeSecret<number>
global UnitInRange: function(string): W.MaybeSecret<boolean>, W.MaybeSecret<boolean>
global UnitInSubgroup: function(? string, ? number): boolean
global UnitInVehicle: function(string): boolean
global UnitInVehicleControlSeat: function(? string): boolean
global UnitInVehicleHidesPetFrame: function(? string): boolean
global UnitIsAFK: function(string): W.MaybeSecret<boolean>
global UnitIsBattlePet: function(string): boolean
global UnitIsBattlePetCompanion: function(string): boolean
global UnitIsBossMob: function(string): boolean
global UnitIsCharmed: function(? string): W.MaybeSecret<boolean>
global UnitIsConnected: function(string): boolean
global UnitIsControlling: function(string): boolean
global UnitIsCorpse: function(? string): boolean
global UnitIsDND: function(string): W.MaybeSecret<boolean>
global UnitIsDead: function(string): boolean
global UnitIsDeadOrGhost: function(string): boolean
global UnitIsEnemy: function(string, string): boolean
global UnitIsFeignDeath: function(string): boolean
global UnitIsFriend: function(string, string): boolean
global UnitIsGameObject: function(? string): boolean
global UnitIsGhost: function(string): boolean
global UnitIsGroupAssistant: function(string): W.MaybeSecret<boolean>
global UnitIsGroupLeader: function(string, ? number): W.MaybeSecret<boolean>
global UnitIsHumanPlayer: function(? string, ? number): boolean
global UnitIsInMyGuild: function(string): boolean
global UnitIsInteractable: function(? string): boolean
global UnitIsLieutenant: function(string): boolean
global UnitIsMercenary: function(string): boolean
global UnitIsMinion: function(string): boolean
global UnitIsNPCAsPlayer: function(? string): boolean
global UnitIsOtherPlayersBattlePet: function(? string): boolean
global UnitIsOtherPlayersPet: function(? string): boolean
global UnitIsOwnerOrControllerOfUnit: function(string, string): W.MaybeSecret<boolean>
global UnitIsPVP: function(string): W.MaybeSecret<boolean>
global UnitIsPVPFreeForAll: function(string): boolean
global UnitIsPVPSanctuary: function(? string): boolean
global UnitIsPlayer: function(? string, ? number): boolean
global UnitIsPlayerControlledOrGroupMember: function(string): boolean
global UnitIsPossessed: function(? string): W.MaybeSecret<boolean>
global UnitIsQuestBoss: function(string): boolean
global UnitIsRaidOfficer: function(? string): W.MaybeSecret<boolean>
global UnitIsSameServer: function(string): boolean
global UnitIsTapDenied: function(string): boolean
global UnitIsTrivial: function(string): boolean
global UnitIsUnconscious: function(string): boolean
global UnitIsUnit: function(string, string): W.MaybeSecret<boolean>
global UnitIsVisible: function(? string): boolean
global UnitIsWildBattlePet: function(string): boolean
global UnitLeadsAnyGroup: function(string): W.MaybeSecret<boolean>
global UnitLevel: function(string): number
global UnitName: function(string): W.MaybeSecret<string>, W.MaybeSecret<string>
global UnitNameFromGUID: function(W.MaybeSecret<string>): W.MaybeSecret<string>, W.MaybeSecret<string>
global UnitNameUnmodified: function(string): W.MaybeSecret<string>, W.MaybeSecret<string>
global UnitNameplateShowsWidgetsOnly: function(string): boolean
global UnitNumPowerBarTimers: function(string): number
global UnitOnTaxi: function(string): boolean
global UnitOwnerGUID: function(string): W.MaybeSecret<string>
global UnitPVPName: function(string): W.MaybeSecret<string>
global UnitPartialPower: function(string, ? number, ? boolean): W.MaybeSecret<number>
global UnitPercentHealthFromGUID: function(W.MaybeSecret<string>): W.MaybeSecret<number>
global UnitPhaseReason: function(string): W.MaybeSecret<number>
global UnitPlayerControlled: function(? string): boolean
global UnitPlayerOrPetInParty: function(? string, ? number): boolean
global UnitPlayerOrPetInRaid: function(? string, ? number): boolean
global UnitPosition: function(string): number, number, number, number
global UnitPower: function(string, ? number, ? boolean): W.MaybeSecret<number>
global UnitPowerBarID: function(string): number
global UnitPowerBarTimerInfo: function(string, ? number): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitPowerDisplayMod: function(number): number
global UnitPowerMax: function(string, ? number, ? boolean): W.MaybeSecret<number>
global UnitPowerMissing: function(string, ? number, ? boolean): W.MaybeSecret<number>
global UnitPowerPercent: W.Global_UnitPowerPercent
global UnitPowerType: function(string, ? number): number, string, number, number, number
global UnitPvpClassification: function(string): number
global UnitQuestTrivialLevelRange: function(string): number
global UnitQuestTrivialLevelRangeScaling: function(string): number
global UnitRace: function(string): W.MaybeSecret<string>, W.MaybeSecret<string>, W.MaybeSecret<number>
global UnitRangedAttackPower: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitRangedDamage: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitReaction: function(string, string): number
global UnitRealmRelationship: function(string): number
global UnitSelectionColor: function(string, ? boolean): number, number, number, number
global UnitSelectionType: function(string, ? boolean): number
global UnitSetRole: function(string, ? string): boolean
global UnitSetRoleEnum: function(string, ? number): boolean
global UnitSex: function(string): W.MaybeSecret<number>
global UnitSexBase: function(string): W.MaybeSecret<number>
global UnitShouldDisplayName: function(string): boolean
global UnitShouldDisplaySpellTargetName: function(string): boolean
global UnitSpellHaste: function(string): W.MaybeSecret<number>
global UnitSpellTargetClass: function(string): W.MaybeSecret<string>
global UnitSpellTargetName: function(string): W.MaybeSecret<string>
global UnitStagger: function(string): W.MaybeSecret<number>
global UnitStat: function(string, number): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitSwitchToVehicleSeat: function(string, number)
global UnitTargetsVehicleInRaidUI: function(? string): boolean
global UnitThreatLeadSituation: function(string, string): W.MaybeSecret<number>
global UnitThreatPercentageOfLead: function(string, string): W.MaybeSecret<number>
global UnitThreatSituation: function(string, ? string): W.MaybeSecret<number>
global UnitTokenFromGUID: function(W.MaybeSecret<string>): W.MaybeSecret<string>
global UnitTreatAsPlayerForDisplay: function(string): boolean
global UnitTrialBankedLevels: function(string): number, number, number
global UnitTrialXP: function(string): number
global UnitUsingVehicle: function(string): boolean
global UnitVehicleSeatCount: function(string): number
global UnitVehicleSeatInfo: function(string, number): string, string, string, boolean, boolean
global UnitVehicleSkin: function(? string): number
global UnitWeaponAttackPower: function(string): W.MaybeSecret<number>, W.MaybeSecret<number>, W.MaybeSecret<number>
global UnitWidgetSet: function(string): number
global UnitXP: function(string): number
global UnitXPMax: function(string): number
global UnlearnSpecialization: function(any, any): W.Data...
global UnmuteSoundFile: function(number | string): W.Data...
global UnregisterEventCallback: function(string, function())
global UnregisterUnitEventCallback: function(string, function(), string)
global UpdateAddOnCPUUsage: function()
global UpdateAddOnMemoryUsage: function()
global UpdateInventoryAlertStatus: function(): W.Data...
global UpdateWarGamesList: function(): W.Data...
global UpdateWindow: function()
global UseAction: function(number, ? number, ? number): W.Data...
global UseInventoryItem: function(number): W.Data...
global UseItemByName: function(number | string, ? string)
global UseQuestLogSpecialItem: function(any): W.Data...
global UseToy: function(number): W.Data...
global UseToyByName: function(string): W.Data...
global UseWorldMapActionButtonSpellOnQuest: function(): W.Data...
global VALDRAKKEN_MAJOR_FACTION_COLOR: W.Global_VALDRAKKEN_MAJOR_FACTION_COLOR
global VENTHYR_RED_COLOR: W.Global_VENTHYR_RED_COLOR
global VERY_DARK_GRAY_COLOR: W.Global_VERY_DARK_GRAY_COLOR
global VERY_LIGHT_GRAY_COLOR: W.Global_VERY_LIGHT_GRAY_COLOR
global Vector2DMixin: W.Global_Vector2DMixin
global Vector3DMixin: W.Global_Vector3DMixin
global VehicleAimDecrement: function(): W.Data...
global VehicleAimDownStart: function(): W.Data...
global VehicleAimDownStop: function(): W.Data...
global VehicleAimGetNormPower: function(): W.Data...
global VehicleAimIncrement: function(): W.Data...
global VehicleAimRequestAngle: function(): W.Data...
global VehicleAimSetNormPower: function(): W.Data...
global VehicleAimUpStart: function(): W.Data...
global VehicleAimUpStop: function(): W.Data...
global VehicleExit: function(): W.Data...
global VehicleMenuBarStatusBarText: W.Global_VehicleMenuBarStatusBarText
global VehicleNextSeat: function(): W.Data...
global VehiclePrevSeat: function(): W.Data...
global ViewGuildRecipes: function(any): W.Data...
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
global WarGameRespond: function(any): W.Data...
global WhiteNormalNumberFont: W.Global_WhiteNormalNumberFont
global WithdrawGuildBankMoney: function(any): W.Data...
global WorldLootObjectExists: function(? string): boolean
global WorldMapFrame: W.Global_WorldMapFrame
global WorldMapTextFont: W.Global_WorldMapTextFont
global WrapTextInColorCode: function(string, string): string
global YELLOW_FONT_COLOR: W.Global_YELLOW_FONT_COLOR
global YELLOW_THREAT_COLOR: W.Global_YELLOW_THREAT_COLOR
global ZoneTextFont: W.Global_ZoneTextFont
global addframetext: function(string): W.Data...
global canaccessallvalues: function(any...): boolean
global canaccesssecrets: function(): boolean
global canaccesstable: function(any): boolean
global canaccessvalue: function(any): boolean
global debuglocals: function(number): string
global debugprofilestart: function()
global debugprofilestop: function(): number
global debugstack: W.Global_debugstack
global dropsecretaccess: function()
global dumpobject: function(? any): W.Data
global forceinsecure: function(): W.Data...
global geterrorhandler: function(): W.Data
global getglobal: function(string): W.Data
global hasanysecretvalues: function(any...): boolean
global hooksecurefunc: W.Global_hooksecurefunc
global issecrettable: function(any): boolean
global issecretvalue: function(any): boolean
global issecure: function(): boolean
global issecurevalue: function(any): boolean, string
global issecurevariable: W.Global_issecurevariable
global mapvalues: function(any, any...): W.Data...
global message: function(string, boolean)
global pcallwithenv: function(string, any): boolean
global ripairs: function(any): W.Data, W.Data, number
global scrub: function(any...): W.Data...
global scrubsecretvalues: function(any...): W.Data...
global secretunwrap: function(any...): W.Data...
global secretwrap: function(any...): W.SecretData...
global securecall: W.Global_securecall
global securecallfunction: function<A, R>(function(A...): (R...), A...): R...
global securecallmethod: function(any, string, any...): W.Data...
global securecopy: function(? any, ? W.SecureCopyOptions): W.Data
global secureexecuterange: function(any, any, any...): W.Data...
global seterrorhandler: function(any): W.Data...
global setglobal: function(string, any)
global setsecurehookforbidden: function(): W.Data...
global settablesecurity: function(any, number)
global tAppendAll: function(any, any)
global tCompare: function(any, any, ? number): boolean
global tContains: function(any, any): boolean
global tDeleteItem: function(any, any)
global tFilter: function<K, V>({K: V}, function(V): (any), ? boolean): {K: V}
global tIndexOf: function(any, any): number
global tInvert: function<K, V>({K: V}): {V: K}
global tUnorderedRemove: function(any, number)
-- Additional globals exist only in Blizzard's secure environment.
global C_CombatLogSecure: W.Global_C_CombatLogSecure
global C_PingSecure: W.Global_C_PingSecure
global GetForbiddenObjectTable: W.Global_GetForbiddenObjectTable
