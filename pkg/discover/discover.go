package discover

import (
	"github.com/robotjoosen/go-display-driver/pkg/tca9548"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/devices/v3/ssd1306"
)

const (
	MaxChannels = 8
)

func Displays(bus i2c.Bus, mux *tca9548.TCA9548) []int {
	var displays []int

	for channel := 0; channel < MaxChannels; channel++ {
		if err := mux.SetChannel(channel); err != nil {
			continue
		}

		if isDisplayPresent(bus, channel) {
			displays = append(displays, channel)
		}
	}

	return displays
}

func isDisplayPresent(bus i2c.Bus, channel int) bool {
	_, err := ssd1306.NewI2C(bus, &ssd1306.Opts{W: 128, H: 64, Rotated: false})
	return err == nil
}
