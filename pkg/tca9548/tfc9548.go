package tca9548

import (
	"errors"
	"fmt"
	"log/slog"

	"periph.io/x/conn/v3/i2c"
)

type TCA9548 struct {
	addresses []uint16
	address   uint16
	channels  []byte
	bus       i2c.Bus
}

func New(bus i2c.Bus) *TCA9548 {
	return &TCA9548{
		bus:       bus,
		channels:  []byte{0b00000001, 0b00000010, 0b00000100, 0b00001000, 0b00010000, 0b00100000, 0b01000000, 0b10000000},
		addresses: []uint16{0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77},
		address:   0x70,
	}
}

func (t *TCA9548) SetAddress(i int) error {
	addrLen := len(t.addresses)
	if i >= addrLen {
		return errors.New("invalid address")
	}

	t.address = t.addresses[i]

	return nil
}

func (t *TCA9548) SetChannel(i int) error {
	if len(t.channels) < i {
		return errors.New("invalid channel")
	}

	fmt.Printf("setting channel: %d : %d\n", i, t.channels[i])

	if err := t.bus.Tx(t.address, []byte{t.channels[i]}, make([]byte, 0)); err != nil {
		slog.Error("failed setting bus",
			slog.Int("index", i),
			slog.Any("target", t.channels[i]),
			slog.String("err", err.Error()),
		)

		return err
	}

	return nil
}
