package main

import (
	"errors"
	"log/slog"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/robotjoosen/go-display-driver/pkg/device"
	"github.com/robotjoosen/go-display-driver/pkg/discover"
	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/display/screen/startup"
	"github.com/robotjoosen/go-display-driver/pkg/panel"
	"github.com/robotjoosen/go-display-driver/pkg/sprite"
	"github.com/robotjoosen/go-display-driver/pkg/tca9548"
	"github.com/robotjoosen/go-rabbit"
	"github.com/wagslane/go-rabbitmq"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	maxRetries = 100
)

func main() {
	e := loadEnv()
	initLog(e.LogLevel)

	bus, tcaMux := initializeBus()
	displayList := discover.Displays(bus, tcaMux)

	if len(displayList) == 0 {
		slog.Error("no displays detected")
		os.Exit(1)
	}

	slog.Info("detected displays",
		slog.Any("displays", displayList),
	)

	p := initializePanel(bus, tcaMux, displayList)

	if err := sprite.LoadAll(e.SpritePath); err != nil {
		slog.Warn("failed to load sprites",
			slog.String("path", e.SpritePath),
			slog.String("err", err.Error()),
		)
	}
	sprite.StartFileWatcher(30*time.Second, e.SpritePath)

	sm := display.NewManager(displayList, display.NewPanelAdapter(p), e.StatePath)

	for _, d := range displayList {
		sm.SetScreen(d, display.ScreenStartup, startup.StartupData{})
		sm.Input(display.RefreshEvent{Display: d})
	}

	if err := sm.LoadState(); err != nil {
		slog.Warn("failed to load persisted state",
			slog.String("err", err.Error()),
		)
	}

	for _, d := range displayList {
		sm.Input(display.RefreshEvent{Display: d})
	}

	conn := connectMessageBus(e.MessagebusURL)

	cStatus, err := rabbit.NewConsumer(conn,
		e.MessageBusExchange,
		[]string{e.MessageBusRoutingKey},
		e.MessageBusQueueName,
	)
	if err != nil {
		panic(err)
	}

	cKeyboard, err := rabbit.NewConsumer(conn,
		e.KeyboardExchange,
		[]string{e.KeyboardRoutingKey},
		e.KeyboardQueueName,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		if err = cStatus.Run(device.HandleMessage); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err = cKeyboard.Run(display.HandleControlInstructions(sm)); err != nil {
			panic(err)
		}
	}()

	<-make(chan bool)
}

func initializeBus() (i2c.Bus, *tca9548.TCA9548) {
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

	tcaMux := tca9548.New(bus)
	if tcaMux == nil {
		slog.Error("failed to initialize multiplexer")

		os.Exit(1)
	}

	return bus, tcaMux
}

func initializePanel(bus i2c.Bus, tcaMux *tca9548.TCA9548, displayList []int) *panel.Panel {
	p, err := panel.New(
		panel.WithI2CBus(bus),
		panel.WithMultiplexer(tcaMux),
	)
	if err != nil {
		slog.Error("failed to initialize panel",
			slog.String("err", err.Error()),
		)

		os.Exit(1)
	}

	for _, d := range displayList {
		if err := p.DisplayAdd(d); err != nil {
			slog.Error("failed to configure display",
				slog.Int("display", d),
				slog.String("err", err.Error()),
			)

			os.Exit(1)
		}
	}

	return p
}

func connectMessageBus(u string) *rabbitmq.Conn {
	mbu, err := url.Parse(u)
	if err != nil {
		panic(err)
	}

	retries := 0
	for {
		if retries >= maxRetries {
			panic(errors.New("cannot connect to message bus"))
		}

		if _, err := net.DialTimeout("tcp", mbu.Host, 1*time.Second); err != nil {
			retries++

			<-time.NewTimer(2 * time.Second).C

			continue
		}

		break
	}

	conn, err := rabbit.NewConnection(u)
	if err != nil {
		panic(err)
	}

	return conn
}
