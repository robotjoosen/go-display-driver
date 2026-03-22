package display

import (
	"encoding/json"
	"log/slog"

	"github.com/wagslane/go-rabbitmq"
)

func HandleControlInstructions(sm *Manager) func(d rabbitmq.Delivery) (action rabbitmq.Action) {
	return func(d rabbitmq.Delivery) (action rabbitmq.Action) {
		slog.Debug("keyboard delivery received",
			"body", string(d.Body),
			"routingKey", d.RoutingKey,
		)

		var msg ControlMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			slog.Error("failed to unmarshal control message",
				"err", err,
				"body", string(d.Body),
			)
			return rabbitmq.NackDiscard
		}

		slog.Debug("control message unmarshaled",
			"keyID", msg.KeyID,
			"action", msg.Action,
			"timestamp", msg.Timestamp,
		)

		sm.Input(ControlEvent{
			KeyID:     msg.KeyID,
			Action:    msg.Action,
			Timestamp: msg.Timestamp,
		})

		return rabbitmq.Ack
	}
}

type ControlMessage struct {
	KeyID     int    `json:"key_id"`
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
}
