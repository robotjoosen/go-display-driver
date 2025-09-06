package main

import (
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	"github.com/robotjoosen/go-display-driver/pkg/image"
	"github.com/robotjoosen/go-display-driver/pkg/panel"
	"github.com/robotjoosen/go-display-driver/pkg/tca9548"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const displayCount = 3

func main() {
	if _, err := host.Init(); err != nil {
		slog.Error("failed to initialize host",
			slog.String("err", err.Error()),
		)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		slog.Error("failed to register i2c bus",
			slog.String("err", err.Error()),
		)

		os.Exit(1)
	}

	tca9548 := tca9548.New(bus)
	if tca9548 == nil {
		slog.Error("failed to initialize multiplexer")

		os.Exit(1)
	}

	p, err := panel.New(
		panel.WithI2CBus(bus),
		panel.WithMultiplexer(tca9548),
	)
	if err != nil {
		slog.Error("failed to initialize panel",
			slog.String("err", err.Error()),
		)

		os.Exit(1)
	}

	slog.Debug("module initialized")

	for i := range displayCount {
		if err := p.DisplayAdd(i); err != nil {
			slog.Error("failed to configure displays",
				slog.String("err", err.Error()),
			)

			os.Exit(1)
		}
	}

	slog.Debug("displays configured")

	randomImageDisplayer(p)
}

func randomImageDisplayer(p *panel.Panel) {
	slog.Debug("running random image displayer")

	for {
		for i := range displayCount {
			if err := p.DisplayWrite(i, image.Images[rand.IntN(4)]); err != nil {
				slog.Error("failed to write to display",
					slog.Int("display", i),
					slog.String("err", err.Error()),
				)
			}

			time.Sleep(time.Millisecond * 100)
		}
	}
}
