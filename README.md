# Display Driver

Multiple Display Driver using the TCA9548a I2C multiplexer and SSD1306 OLED displays.

## Architecture

The display driver uses a **Screen Manager** architecture to control multiple I2C displays:

- **Displays**: Physical OLED displays managed via TCA9548A I2C multiplexer, auto-detected at startup (`pkg/display`)
- **Devices**: SBCs, Routers, MCUs that send status data via RabbitMQ (`pkg/device`) — a fully independent package, not routed through the display Manager
- **Screens**: Visual representations shown on displays (device status, network status, menu, etc.), each pulling its own data at render time

### Display Detection

Displays are dynamically detected at startup via `pkg/discover`:

```go
displayList := discover.Displays(bus, tcaMux)
```

The detection iterates through channels 0-7 on the TCA9548 multiplexer and identifies SSD1306 OLED displays.

### Key Components

| Component | Description |
|-----------|-------------|
| `Manager` | Central coordinator for displays, screen state, and event processing (`pkg/display`) |
| `Screen Registry` | Self-registering screen implementations, one package per screen type |
| `Event System` | Buffered channel (100) for thread-safe event handling |
| Device registry | Independent registry of device data keyed by ID (`pkg/device`) |

### Data Flow

1. **Device Updates**: RabbitMQ → `device.HandleMessage` → registered in the device registry — never touches the display `Manager`
2. **Display Refresh**: `RefreshEvent` → `Screen.Render()` reads its own state/device data → draw to display
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
- `MODE` - Runtime mode (default: `DEV`)
- `LOG_LEVEL` - Logging level (default: `INFO`)
- `SPRITE_PATH` - Path to sprite assets (default: `~/.config/go-display-driver/sprites`)
- `MESSAGE_BUS_URL` - RabbitMQ connection URL
- `MESSAGE_BUS_EXCHANGE` - Exchange for device status messages
- `MESSAGE_BUS_ROUTING_KEY` - Routing key for device messages
- `MESSAGE_BUS_QUEUE_NAME` - Queue name for device messages
- `KEYBOARD_EXCHANGE` - Exchange for keyboard events
- `KEYBOARD_ROUTING_KEY` - Routing key for keyboard events
- `KEYBOARD_QUEUE_NAME` - Queue name for keyboard events