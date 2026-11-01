package config

type GameItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"desc"`
	ExeName     string `json:"exe"`
	WindowTitle string `json:"windowTitle"`
	Badge       string `json:"badge"`
	SteamFolder string `json:"steamFolder,omitempty"`
	SteamAppID  string `json:"steamAppId,omitempty"`
}

const (
	QuestDurationSeconds = 15*60 + 15
	DefaultPort          = 39482
)

var SupportedGames = []GameItem{
	{
		ID:          "roblox",
		Title:       "Roblox",
		Description: "Scary Shawarma Kiosk, 99 Nights in the Forest",
		ExeName:     "RobloxPlayerBeta.exe",
		WindowTitle: "Roblox",
		Badge:       "700 Orbs",
	},
	{
		ID:          "cod_mw4",
		Title:       "Call of Duty: Modern Warfare 4",
		Description: "Квесты на 700 Orbs, 2XP Token, Calling Card",
		ExeName:     "cod26-cod.exe",
		WindowTitle: "Call of Duty: Modern Warfare 4",
		Badge:       "700 Orbs",
	},
	{
		ID:          "re_sale",
		Title:       "Resident Evil Sale",
		Description: "Resident Evil 7: Biohazard, Village, RE4",
		ExeName:     "re7.exe",
		WindowTitle: "Resident Evil 7",
		Badge:       "700 Orbs",
		SteamFolder: "Resident Evil 7 Biohazard",
		SteamAppID:  "418370",
	},
	{
		ID:          "once_human",
		Title:       "Once Human",
		Description: "Once Human Free Expansion Quest",
		ExeName:     "once_human.exe",
		WindowTitle: "Once Human",
		Badge:       "700 Orbs",
		SteamFolder: "Once Human",
		SteamAppID:  "2139460",
	},
	{
		ID:          "where_winds_meet",
		Title:       "Where Winds Meet",
		Description: "Version 2.3: Friends From Afar Quest",
		ExeName:     "wwm.exe",
		WindowTitle: "Where Winds Meet",
		Badge:       "700 Orbs",
	},
	{
		ID:          "endfield",
		Title:       "Arknights: Endfield",
		Description: "Endfield is now on Steam Quest",
		ExeName:     "endfield.exe",
		WindowTitle: "ARKNIGHTS: ENDFIELD",
		Badge:       "700 Orbs",
		SteamFolder: "ARKNIGHTS ENDFIELD",
		SteamAppID:  "2345020",
	},
	{
		ID:          "firefox",
		Title:       "Firefox Browser",
		Description: "Использование браузера Firefox 15 минут",
		ExeName:     "firefox.exe",
		WindowTitle: "Mozilla Firefox",
		Badge:       "700 Orbs",
	},
	{
		ID:          "valorant",
		Title:       "VALORANT",
		Description: "VALORANT Hush Quest (Декорация)",
		ExeName:     "win64/VALORANT-Win64-Shipping.exe",
		WindowTitle: "VALORANT",
		Badge:       "Декорация",
	},
	{
		ID:          "fc27",
		Title:       "EA SPORTS FC 27",
		Description: "EA SPORTS FC™ 27 (Steam SKU 4080220)",
		ExeName:     "fc27.exe",
		WindowTitle: "EA SPORTS FC™ 27",
		Badge:       "700 Orbs",
		SteamFolder: "EA SPORTS FC 27",
		SteamAppID:  "4080220",
	},
}
