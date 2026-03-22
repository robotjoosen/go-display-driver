package display

import (
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/panel"
)

var _ Panel = (*panelAdapter)(nil)

type panelAdapter struct {
	p *panel.Panel
}

func (a *panelAdapter) DisplayDraw(channel int, img image.Image) error {
	return a.p.DisplayDraw(channel, img)
}

func NewPanelAdapter(p *panel.Panel) Panel {
	return &panelAdapter{p: p}
}
