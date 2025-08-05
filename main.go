package main

import (
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/robotjoosen/go-display-driver/pkg/image"
	"github.com/robotjoosen/go-display-driver/pkg/panel"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	p := panel.New(bus)

	for i := range 4 {
		if err := p.DisplayAdd(1 << i); err != nil {
			log.Fatal(err.Error())

			os.Exit(1)
		}
	}

	for {
		for i := range 4 {
			if err := p.DisplayWrite(1<<i, image.Images[rand.IntN(4)]); err != nil {
				log.Println(err.Error())
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}
