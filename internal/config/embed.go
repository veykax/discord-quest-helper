package config

import (
	_ "embed"
)

//go:embed default_games.json
var defaultGamesJSON []byte

