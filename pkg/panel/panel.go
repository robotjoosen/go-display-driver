package panel

import (
	"errors"
	"image"
	"sync"

	"github.com/puzpuzpuz/xsync/v4"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/devices/v3/ssd1306"
)

type Panel struct {
	mu          sync.Mutex
	bus         i2c.Bus
	multiplexer ChannelAware
	displays    *xsync.Map[int, *ssd1306.Dev]
}

func New(optionFuncs ...OptionFunc) (*Panel, error) {
	options, err := getOptions(optionFuncs...)
	if err != nil {
		return nil, err
	}

	return &Panel{
		bus:         options.bus,
		multiplexer: options.multiplexer,
		displays:    xsync.NewMap[int, *ssd1306.Dev](),
	}, nil
}

func (p *Panel) DisplayAdd(channel int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.multiplexer.SetChannel(channel)
	if err != nil {
		return err
	}

	dev, err := ssd1306.NewI2C(p.bus, &ssd1306.Opts{W: 128, H: 64, Rotated: false})
	if err != nil {
		return err
	}

	p.displays.Store(channel, dev)

	return nil
}

func (p *Panel) DisplayWrite(channel int, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.multiplexer.SetChannel(channel)
	if err != nil {
		return err
	}

	if d, ok := p.displays.Load(channel); ok {
		_, err = d.Write(data)
	}

	return err
}

func (p *Panel) DisplayDraw(channel int, img image.Image) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.multiplexer.SetChannel(channel); err != nil {
		return err
	}

	if d, ok := p.displays.Load(channel); ok {
		return d.Draw(img.Bounds(), img, image.Point{})
	}

	return errors.New("display not found")
}
