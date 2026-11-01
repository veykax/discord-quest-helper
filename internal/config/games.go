package config

type GameItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ExeName      string `json:"exe"`
	WindowTitle  string `json:"windowTitle"`
	SteamFolder  string `json:"steamFolder,omitempty"`
	SteamAppID   string `json:"steamAppId,omitempty"`
	DiscordAppID string `json:"discordAppId,omitempty"`
}

const (
	QuestDurationSeconds = 15*60 + 15
	DefaultPort          = 39482
)

var SupportedGames = []GameItem{
	{
		ID:           "roblox",
		Title:        "Roblox",
		ExeName:      "RobloxPlayerBeta.exe",
		WindowTitle:  "Roblox",
		DiscordAppID: "363445589247131668",
	},
	{
		ID:           "cod_mw4",
		Title:        "Call of Duty: Modern Warfare 4",
		ExeName:      "cod26-cod.exe",
		WindowTitle:  "Call of Duty: Modern Warfare 4",
		DiscordAppID: "1511443377927880775",
	},
	{
		ID:           "re_sale",
		Title:        "Resident Evil Sale",
		ExeName:      "re7.exe",
		WindowTitle:  "RESIDENT EVIL 7 biohazard",
		DiscordAppID: "443160369192632330",
	},
	{
		ID:           "once_human",
		Title:        "Once Human",
		ExeName:      "once human/once_human.exe",
		WindowTitle:  "Once Human",
		SteamFolder:  "Once Human",
		SteamAppID:   "2139460",
		DiscordAppID: "1260340082150346932",
	},
	{
		ID:           "where_winds_meet",
		Title:        "Where Winds Meet",
		ExeName:      "wwm.exe",
		WindowTitle:  "Where Winds Meet",
		DiscordAppID: "1437509662303059998",
	},
	{
		ID:           "endfield",
		Title:        "Arknights: Endfield",
		ExeName:      "endfield.exe",
		WindowTitle:  "ARKNIGHTS: ENDFIELD",
		DiscordAppID: "1461154307171811401",
	},
	{
		ID:           "firefox",
		Title:        "Firefox Browser",
		ExeName:      "firefox.exe",
		WindowTitle:  "Mozilla Firefox",
		DiscordAppID: "1442729602559643668",
	},
	{
		ID:           "valorant",
		Title:        "VALORANT",
		ExeName:      "win64/VALORANT-Win64-Shipping.exe",
		WindowTitle:  "VALORANT",
		DiscordAppID: "700136079562375258",
	},
	{
		ID:           "apex",
		Title:        "Apex Legends",
		ExeName:      "r5apex.exe",
		WindowTitle:  "Apex Legends",
		DiscordAppID: "542075586886107149",
	},
	{
		ID:           "fc27",
		Title:        "EA SPORTS FC 27",
		ExeName:      "fc27.exe",
		WindowTitle:  "EA Sports FC 27",
		SteamFolder:  "EA SPORTS FC 27",
		SteamAppID:   "4080220",
		DiscordAppID: "1531874756096295054",
	},
}
