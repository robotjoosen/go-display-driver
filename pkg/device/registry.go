package device

import (
	"encoding/json"
	"log/slog"

	"github.com/puzpuzpuz/xsync/v4"
	"github.com/wagslane/go-rabbitmq"
)

var registry = xsync.NewMap[string, DeviceData]()

func Register(id string, data DeviceData) {
	registry.Store(id, data)
}

func Get(id string) (DeviceData, bool) {
	return registry.Load(id)
}

func GetByType(t DeviceType) []DeviceData {
	var result []DeviceData
	registry.Range(func(id string, data DeviceData) bool {
		if data.Type() == t {
			result = append(result, data)
		}
		return true
	})
	return result
}

func All() []DeviceData {
	var result []DeviceData
	registry.Range(func(id string, data DeviceData) bool {
		result = append(result, data)
		return true
	})
	return result
}

func HandleMessage(d rabbitmq.Delivery) rabbitmq.Action {
	var msg DeviceMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		slog.Error("failed to unmarshal device message",
			slog.String("message_id", d.MessageId),
			slog.String("correlation_id", d.CorrelationId),
			slog.String("routing_key", d.RoutingKey),
		)
		return rabbitmq.NackDiscard
	}

	devData := CreateDevice(msg)
	Register(devData.ID(), devData)

	return rabbitmq.Ack
}
