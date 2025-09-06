package panel

import (
	"errors"

	"periph.io/x/conn/v3/i2c"
)

type OptionFunc func(*Options)

type Options struct {
	bus         i2c.Bus
	multiplexer ChannelAware
}

func getOptions(of ...OptionFunc) (Options, error) {
	o := Options{}
	for _, optionFunc := range of {
		optionFunc(&o)
	}

	if o.bus == nil {
		return Options{}, errors.New("required bus is missing")
	}

	if o.multiplexer == nil {
		return Options{}, errors.New("required multiplexer is missing")
	}

	return o, nil
}

func WithI2CBus(bus i2c.Bus) OptionFunc {
	return func(o *Options) {
		o.bus = bus
	}
}

func WithMultiplexer(mutliplexer ChannelAware) OptionFunc {
	return func(o *Options) {
		o.multiplexer = mutliplexer
	}
}
