# Display Driver

Multiple Display Driver using the TCA9548a I2C multiplexer and SSD1306 OLED displays.

## Architecture

The display driver uses a **Screen Manager** architecture to control multiple I2C displays:

- **Displays**: Physical OLED displays managed via TCA9548A I2C multiplexer, auto-detected at startup
- **Devices**: SBCs, Routers, MCUs that send status data via RabbitMQ
- **Screens**: Visual representations shown on displays (device status, network status, menu, etc.)

### Display Detection

Displays are dynamically detected at startup via `pkg/discover`:

```go
displayList := discover.Displays(bus, tcaMux)
```

The detection iterates through channels 0-7 on the TCA9548 multiplexer and identifies SSD1306 OLED displays.

### Key Components

| Component | Description |
|-----------|-------------|
| `Manager` | Central coordinator managing displays, devices, and event processing |
| `Screen Registry` | Self-registering screen implementations |
| `Event System` | Buffered channel (100) for thread-safe event handling |

### Data Flow

1. **Device Updates**: RabbitMQ → `handleDeviceUpdate` → `UpdateDeviceEvent` → stored in Manager
2. **Display Refresh**: `RefreshEvent` → render screen → draw to display
3. **User Input**: RabbitMQ → `HandleControlInstructions` → `ControlEvent` → navigation/selection

## How to install

### Requirements

- [Taskfile](https://taskfile.dev/docs/installation)

### TLDR;
```shell
git clone git@github.com:robotjoosen/go-display-driver.git
task build
task install
```

## Configuration

Environment variables (see `cmd/app/setup.go`):
- `MESSAGEBUS_URL` - RabbitMQ connection URL
- `MESSAGEBUS_EXCHANGE` - Exchange for device status messages
- `MESSAGEBUS_ROUTING_KEY` - Routing key for device messages
- `MESSAGEBUS_QUEUE_NAME` - Queue name for device messages
- `KEYBOARD_EXCHANGE` - Exchange for keyboard events
- `KEYBOARD_QUEUE_NAME` - Queue name for keyboard events
- `LOG_LEVEL` - Logging level (default: info)
- `SPRITE_PATH` - Path to sprite assets