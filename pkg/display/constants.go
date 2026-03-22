package display

type ScreenType string

const (
	ScreenStartup       ScreenType = "startup"
	ScreenNetworkStatus ScreenType = "network_status"
	ScreenDeviceStatus  ScreenType = "device_status"
	ScreenAIInteraction ScreenType = "ai_interaction"
	ScreenUICharacter   ScreenType = "ui_character"
	ScreenLoading       ScreenType = "loading"
	ScreenError         ScreenType = "error"
	ScreenMenu          ScreenType = "menu"
)

const (
	KeyCycleDisplay = 7
	KeyCycleScreen  = 6
	KeyPrev         = 4
	KeyNext         = 5
	KeySelect       = 3

	ActionRelease = "RELEASE"

	RefreshDebounceMs = 100
)

var ScreenTypeCycleOrder = []ScreenType{
	ScreenDeviceStatus,
	ScreenNetworkStatus,
	ScreenAIInteraction,
	ScreenMenu,
	ScreenLoading,
	ScreenError,
	ScreenStartup,
}
