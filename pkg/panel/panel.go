package panel

import (
	"github.com/puzpuzpuz/xsync/v4"
	"github.com/robotjoosen/go-display-driver/pkg/tca9548"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/devices/v3/ssd1306"
)

type Panel struct {
	bus         i2c.Bus
	displays    *xsync.Map[byte, *ssd1306.Dev]
	multiplexer *tca9548.TCA9548
}

func New(bus i2c.Bus) *Panel {
	tca9548 := tca9548.New(bus).SetAddress(0)
	if tca9548 == nil {
		return nil
	}

	return &Panel{
		bus:         bus,
		displays:    xsync.NewMap[byte, *ssd1306.Dev](),
		multiplexer: tca9548,
	}
}

func (p *Panel) DisplayAdd(target byte) error {
	p.multiplexer.SetTarget(target)
	dev, err := ssd1306.NewI2C(p.bus, &ssd1306.Opts{W: 128, H: 64})
	if err != nil {
		return err
	}

	p.displays.Store(target, dev)

	return nil
}

func (p *Panel) DisplayWrite(target byte, data []byte) error {
	p.multiplexer.SetTarget(target)
	d, _ := p.displays.Load(target)
	_, err := d.Write(data)

	return err
}
