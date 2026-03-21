package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/robotjoosen/go-display-driver/pkg/panel"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
	"github.com/robotjoosen/go-display-driver/pkg/screens/device"
	"github.com/robotjoosen/go-display-driver/pkg/tca9548"
	"github.com/robotjoosen/go-rabbit"
	"github.com/wagslane/go-rabbitmq"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

type SysUsageMessage struct {
	Name string           `json:"name"`
	Mem  MemoryMessage    `json:"memory"`
	Cpu  CPUMessage       `json:"cpu"`
	Nic  []NetworkMessage `json:"network_interfaces"`
	Dsk  []DiskMessage    `json:"disks"`
}

type MemoryMessage struct {
	Free  uint64 `json:"free"`
	Used  uint64 `json:"used"`
	Total uint64 `json:"total"`
}

type CPUMessage struct {
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	User   float64 `json:"user"`
}

type NetworkMessage struct {
	Name string `json:"name"`
	Rx   uint64 `json:"rx_bytes"`
	Tx   uint64 `json:"tx_bytes"`
}

type DiskMessage struct {
	Name   string `json:"name"`
	Reads  uint64 `json:"reads"`
	Writes uint64 `json:"writes"`
}

const (
	displayCount = 3
	maxRetries   = 100
)

var (
	devices = map[string]struct {
		hostname  string
		name      string
		display   int
		online    bool
		memory    uint64
		cpu       float64
		disk      float64
		networkRx uint64
		networkTx uint64
	}{
		"rocket": {
			name:    "rocket.local",
			display: 0,
		},
		"beanie": {
			name:    "beanie.local",
			display: 1,
		},
		"orangepizerolts": {
			name:    "socks.local",
			display: 2,
		},
	}
)

func main() {
	e := loadEnv()
	initLog(e.LogLevel)

	p := initializePanel()

	screens.Register(screens.ScreenDeviceStatus, device.New(device.DeviceStatusData{}))

	conn := connectMessageBus(e.MessagebusURL)
	c, err := rabbit.NewConsumer(conn,
		e.MessageBusExchange,
		[]string{e.MessageBusRoutingKey},
		e.MessageBusQueueName,
	)
	if err != nil {
		panic(err)
	}

	for _, dev := range devices {
		screen, _ := screens.Get(screens.ScreenDeviceStatus)
		p.DisplayDraw(dev.display, screen.Render(device.DeviceStatusData{
			ID:        dev.name,
			Online:    dev.online,
			CPU:       dev.cpu,
			Memory:    dev.memory,
			NetworkRx: dev.networkRx,
			NetworkTx: dev.networkTx,
		}))
	}

	if err = c.Run(handleSysStatus(p)); err != nil {
		panic(err)
	}

	<-make(chan bool)
}

func initializePanel() *panel.Panel {
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

	for i := range displayCount {
		if err := p.DisplayAdd(i); err != nil {
			slog.Error("failed to configure displays",
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

func handleSysStatus(p *panel.Panel) func(d rabbitmq.Delivery) (action rabbitmq.Action) {
	return func(d rabbitmq.Delivery) (action rabbitmq.Action) {
		dLog := slog.With(
			slog.String("message_id", d.MessageId),
			slog.String("correlation_id", d.CorrelationId),
			slog.String("routing_key", d.RoutingKey),
		)

		var msg SysUsageMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			dLog.Error("failed to unmarshal message")

			return rabbitmq.NackDiscard
		}

		dev, ok := devices[msg.Name]
		if !ok {
			dLog.Warn("unknown device", slog.String("hostname", msg.Name))

			return rabbitmq.NackDiscard
		}

		dev.online = true
		dev.cpu = msg.Cpu.Idle
		dev.memory = msg.Mem.Free
		for _, nic := range msg.Nic {
			if nic.Name == "eth0" {
				dev.networkRx = nic.Rx
				dev.networkTx = nic.Tx
			}
		}
		devices[msg.Name] = dev

		screen, _ := screens.Get(screens.ScreenDeviceStatus)
		p.DisplayDraw(dev.display, screen.Render(device.DeviceStatusData{
			ID:        dev.name,
			Online:    dev.online,
			CPU:       dev.cpu,
			Memory:    dev.memory,
			NetworkRx: dev.networkRx,
			NetworkTx: dev.networkTx,
		}))

		return rabbitmq.Ack
	}
}
