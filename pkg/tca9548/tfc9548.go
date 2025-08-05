package tca9548

import (
	"fmt"

	"periph.io/x/conn/v3/i2c"
)

type TCA9548 struct {
	address       []uint16
	targetAddress int
	bus           i2c.Bus
}

func New(bus i2c.Bus) *TCA9548 {
	return &TCA9548{
		bus:           bus,
		address:       []uint16{0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77},
		targetAddress: 0,
	}
}

func (t *TCA9548) SetAddress(i int) *TCA9548 {
	addrLen := len(t.address)
	if i >= addrLen {
		return t // TODO: decide if an error should be thrown, or if logging is enough
	}

	t.targetAddress = i

	return t
}

func (t *TCA9548) SetTarget(i byte) {
	if err := t.bus.Tx(t.getAddress(), []byte{i}, make([]byte, 0)); err != nil {
		fmt.Println(err.Error())

		return
	}

	fmt.Println("set tca9548 target:", i)
}

func (t *TCA9548) getAddress() uint16 {
	return t.address[t.targetAddress]
}
