# Display Driver

![CI](https://github.com/robotjoosen/go-display-driver/actions/workflows/ci.yml/badge.svg)

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

## Installing

Part of the [Mini Lab](https://github.com/robotjoosen/minilab-agent) fleet — installed the same
way as `minilab-agent`. On the target device (`task` is optional; the scripts work standalone
too):

```shell
task install
```

This detects the device's architecture, pulls the matching binary from the
[latest release](https://github.com/robotjoosen/go-display-driver/releases), and walks through
setting it up as a systemd service — asking for the RabbitMQ URL, confirming before every
system-changing step (writing the unit file, enabling/starting the service).

To update later:

```shell
task update
```

`update` preserves whatever's already configured and only asks for values genuinely missing
from the existing install (e.g. a setting a newer release added).

To remove it entirely:

```shell
task uninstall
```

Each task is a thin wrapper around the matching script under `scripts/` — see
[scripts/install.sh](scripts/install.sh), [scripts/update.sh](scripts/update.sh), and
[scripts/uninstall.sh](scripts/uninstall.sh) if you want to run them directly (e.g. via
`curl | bash`) without cloning the repo or installing Taskfile.

### Building from source

Requires [Taskfile](https://taskfile.dev/docs/installation).

```shell
git clone git@github.com:robotjoosen/go-display-driver.git
task build:arm
```

Releases are built and published automatically by [CI](.github/workflows/release.yml) on every
`v*` tag — `task build`/`build:arm`/`build:darwin` are for local development builds.

## Configuration

Set via `Environment=` lines in the systemd unit that `install.sh`/`update.sh` write at
`/etc/systemd/system/display_driver.service` — not a local `.env` file. Environment variables
(see `cmd/app/setup.go`):
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