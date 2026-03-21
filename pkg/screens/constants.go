package screens

type ScreenType string

const (
	ScreenNetworkStatus ScreenType = "network_status"
	ScreenDeviceStatus  ScreenType = "device_status"
	ScreenAIInteraction ScreenType = "ai_interaction"
	ScreenUICharacter   ScreenType = "ui_character"
	ScreenLoading       ScreenType = "loading"
	ScreenError         ScreenType = "error"
	ScreenMenu          ScreenType = "menu"
)
