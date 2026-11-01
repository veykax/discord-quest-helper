package config

type GameItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ExeName     string `json:"exe"`
	WindowTitle string `json:"windowTitle"`
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
		ExeName:     "RobloxPlayerBeta.exe",
		WindowTitle: "Roblox",
	},
	{
		ID:          "cod_mw4",
		Title:       "Call of Duty: Modern Warfare 4",
		ExeName:     "cod26-cod.exe",
		WindowTitle: "Call of Duty: Modern Warfare 4",
	},
	{
		ID:          "re_sale",
		Title:       "Resident Evil Sale",
		ExeName:     "re7.exe",
		WindowTitle: "RESIDENT EVIL 7 biohazard",
	},
	{
		ID:          "once_human",
		Title:       "Once Human",
		ExeName:     "once_human.exe",
		WindowTitle: "Once Human",
		SteamFolder: "Once Human",
		SteamAppID:  "2139460",
	},
	{
		ID:          "where_winds_meet",
		Title:       "Where Winds Meet",
		ExeName:     "wwm.exe",
		WindowTitle: "Where Winds Meet",
	},
	{
		ID:          "endfield",
		Title:       "Arknights: Endfield",
		ExeName:     "endfield.exe",
		WindowTitle: "ARKNIGHTS: ENDFIELD",
	},
	{
		ID:          "firefox",
		Title:       "Firefox Browser",
		ExeName:     "firefox.exe",
		WindowTitle: "Mozilla Firefox",
	},
	{
		ID:          "valorant",
		Title:       "VALORANT",
		ExeName:     "win64/VALORANT-Win64-Shipping.exe",
		WindowTitle: "VALORANT",
	},
	{
		ID:          "fc27",
		Title:       "EA SPORTS FC 27",
		ExeName:     "fc27.exe",
		WindowTitle: "EA SPORTS FC™ 27",
		SteamFolder: "EA SPORTS FC 27",
		SteamAppID:  "4080220",
	},
}
