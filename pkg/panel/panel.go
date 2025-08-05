package panel

import (
	"github.com/puzpuzpuz/xsync/v4"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/devices/v3/ssd1306"
)

type Panel struct {
	bus         i2c.Bus
	multiplexer ChannelAware
	displays    *xsync.Map[int, *ssd1306.Dev]
}

func New(optionFuncs ...OptionFunc) *Panel {
	options, err := getOptions(optionFuncs...)
	if err != nil {
		return nil
	}

	return &Panel{
		bus:         options.bus,
		multiplexer: options.multiplexer,
		displays:    xsync.NewMap[int, *ssd1306.Dev](),
	}
}

func (p *Panel) DisplayAdd(channel int) error {
	if err := p.multiplexer.SetChannel(channel); err != nil {
		return err
	}

	dev, err := ssd1306.NewI2C(p.bus, &ssd1306.Opts{W: 128, H: 64})
	if err != nil {
		return err
	}

	p.displays.Store(channel, dev)

	return nil
}

func (p *Panel) DisplayWrite(channel int, data []byte) error {
	if err := p.multiplexer.SetChannel(channel); err != nil {
		return err
	}

	d, _ := p.displays.Load(channel)
	_, err := d.Write(data)

	return err
}
