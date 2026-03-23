package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/robotjoosen/go-display-driver/pkg/env"
)

const (
	modeDevelopment = "DEV"

	defaultMode               = modeDevelopment
	defaultLogLevel           = "INFO"
	defaultMessageBusURL      = "amqp://guest:guest@localhost:5672"
	defaultRoutingKey         = "health.ping"
	defaultExchange           = "health"
	defaultQueueName          = "display-driver"
	defaultKeyboardExchange   = "keyboard"
	defaultKeyboardRoutingKey = "keyboard.event"
	defaultKeyboardQueueName  = "display-keyboard"
	defaultSpritePath         = "~/.config/go-display-driver/sprites"
	defaultStatePath          = "~/.config/go-display-driver/state.json"
)

type Environment struct {
	Mode                 string     `mapstructure:"MODE"`
	LogLevel             slog.Level `mapstructure:"LOG_LEVEL"`
	MessagebusURL        string     `mapstructure:"MESSAGE_BUS_URL"`
	MessageBusExchange   string     `mapstructure:"MESSAGE_BUS_EXCHANGE"`
	MessageBusRoutingKey string     `mapstructure:"MESSAGE_BUS_ROUTING_KEY"`
	MessageBusQueueName  string     `mapstructure:"MESSAGE_BUS_QUEUE_NAME"`
	KeyboardExchange     string     `mapstructure:"KEYBOARD_EXCHANGE"`
	KeyboardRoutingKey   string     `mapstructure:"KEYBOARD_ROUTING_KEY"`
	KeyboardQueueName    string     `mapstructure:"KEYBOARD_QUEUE_NAME"`
	StatePath            string     `mapstructure:"STATE_PATH"`
	SpritePath           string     `mapstructure:"SPRITE_PATH"`
}

func initLog(level slog.Level) {
	hostname, err := os.Hostname()
	if err != nil {
		os.Exit(1)
	}

	slog.SetDefault(slog.
		New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).
		With(
			slog.String("hostname", hostname),
		),
	)
}

func expandPath(path string) string {
	if path == "" {
		return ""
	}

	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[1:])
		}
	}

	return path
}

func loadEnv() Environment {
	e, err := env.Load[Environment](map[string]any{
		"MODE":                    defaultMode,
		"LOG_LEVEL":               defaultLogLevel,
		"SPRITE_PATH":             defaultSpritePath,
		"MESSAGE_BUS_URL":         defaultMessageBusURL,
		"MESSAGE_BUS_EXCHANGE":    defaultExchange,
		"MESSAGE_BUS_ROUTING_KEY": defaultRoutingKey,
		"MESSAGE_BUS_QUEUE_NAME":  defaultQueueName,
		"KEYBOARD_EXCHANGE":       defaultKeyboardExchange,
		"KEYBOARD_ROUTING_KEY":    defaultKeyboardRoutingKey,
		"KEYBOARD_QUEUE_NAME":     defaultKeyboardQueueName,
		"STATE_PATH":              defaultStatePath,
	})
	if err != nil {
		slog.Error("failed to load environment", "err", err.Error())

		os.Exit(1)
	}

	e.StatePath = expandPath(e.StatePath)
	e.SpritePath = expandPath(e.SpritePath)

	return e
}
