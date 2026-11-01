package spoofer

type Manager interface {
	Start(exeName string, windowTitle string, steamFolder string, steamAppID string, discordAppID string) error
	Stop() error
	IsRunning() bool
	CurrentGame() string
}

func New() Manager {
	return newPlatformManager()
}
