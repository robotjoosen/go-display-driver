package display

import "image"

type Screen interface {
	Render(display int, m *Manager) image.Image
}

type TransitionHandler interface {
	HandleSelect(display int, m *Manager)
}
