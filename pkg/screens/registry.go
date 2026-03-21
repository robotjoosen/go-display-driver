package screens

import "github.com/puzpuzpuz/xsync/v4"

var registry = xsync.NewMap[ScreenType, Screen]()

func Register(screenType ScreenType, screen Screen) {
	registry.Store(screenType, screen)
}

func Get(screenType ScreenType) (Screen, bool) {
	return registry.Load(screenType)
}
