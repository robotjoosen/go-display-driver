# Agent Documentation

## Architecture Overview

### Screen Manager System

The display driver uses a Screen Manager architecture to control multiple I2C displays via a TCA9548A I2C multiplexer.

```
┌─────────────────────────────────────────────────────────────┐
│                        main.go                               │
│                                                              │
│  ┌──────────────────┐    ┌───────────────────────────────┐  │
│  │ RabbitMQ Consumer │    │     RabbitMQ Consumer          │  │
│  │ device.HandleMessage │  │  display.HandleControlInstructions │
│  └────────┬─────────┘    └─────────────┬─────────────────┘  │
│           │                              │                    │
│           ▼                              ▼                    │
│  ┌──────────────────┐         ┌──────────────────────────┐  │
│  │  device registry  │         │   Manager.Input()         │  │
│  │  (pkg/device)      │         │                           │  │
│  │  keyed by ID       │         │   Event Queue (buffered, │  │
│  │  independent of    │         │   100)                    │  │
│  │  Manager           │         │         │                 │  │
│  └────────────────────┘         │         ▼                 │  │
│                                  │   eventLoop() goroutine   │  │
│                                  │         │                 │  │
│                                  │         ▼                 │  │
│                                  │   handle(Event)           │  │
│                                  └───────────────────────────┘  │
│                                              │                   │
└──────────────────────────────────────────────┼───────────────────┘
                                               │
                                               ▼
                                   ┌─────────────────────┐
                                   │   Screen.Render()     │
                                   │   pulls its own       │
                                   │   device/state data    │
                                   │   → panel.DisplayDraw │
                                   └─────────────────────┘
```

**Important:** device data and display/screen state are two independent subsystems. `pkg/device` has its own registry and is never routed through `Manager` or its event queue — screens read device data directly from `pkg/device` at render time.

### Display Detection

Displays are dynamically detected via the `pkg/discover` package at startup:

```go
displayList := discover.Displays(bus, tcaMux)
```

`Displays()` iterates through channels 0-7 on the TCA9548 multiplexer and attempts to open an SSD1306 OLED display on each. Only channels that respond are included in the final list.

## Key Concepts

### Displays vs Devices

- **Display**: A physical OLED display (channel 0-7 on the TCA9548A multiplexer), owned by `pkg/display`
- **Device**: A computer (SBC, Router, MCU) that reports status via RabbitMQ, owned by `pkg/device`

**Important**: Display and Device are **fully independent packages**. A display shows a screen type; if that screen needs device data, it calls into `pkg/device`'s registry itself during `Render()`. `Manager` has no concept of devices at all.

### Package Layout

| Package | Responsibility |
|---------|-----------------|
| `cmd/app` | Entrypoint: env/log setup, I2C bus + panel init, wiring RabbitMQ consumers |
| `pkg/display` | `Manager`, event loop, `Screen`/`TransitionHandler` interfaces, screen registry, per-display state store |
| `pkg/display/screen/*` | One package per screen type (`ai`, `character`, `device`, `error`, `loading`, `menu`, `network`, `startup`), each self-registering via `init()` |
| `pkg/device` | Device data model, its own registry (by ID), RabbitMQ message handling — independent of `pkg/display` |
| `pkg/discover` | Probes the TCA9548A channels for attached SSD1306 displays |
| `pkg/draw` | Pixel-level drawing primitives on `*image.Gray` (lines, circles, rectangles, text, sprites, error screen) |
| `pkg/panel` | Low-level per-channel SSD1306 device management behind the TCA9548A multiplexer |
| `pkg/sprite` | PNG sprite loading + hot-reload file watcher |
| `pkg/env` | Typed environment variable loading |
| `pkg/tca9548` | TCA9548A I2C multiplexer driver |

### Event System

Events flow through a buffered channel (100) and are processed by the `eventLoop()` goroutine in `pkg/display/manager.go`:

| Event | Description | Handler Action |
|-------|-------------|----------------|
| `RefreshEvent` | Request display refresh | `queueRefresh(display)` |
| `ControlEvent` | Keyboard input | `handleControl()` - navigation, selection |
| `SelectEvent` | User selected current item | Call `TransitionHandler.HandleSelect()` for the current screen |
| `NextDisplayEvent` | Navigate to next display | `selectNext()` + refresh |
| `PrevDisplayEvent` | Navigate to prev display | `selectPrev()` + refresh |
| `ListUpEvent` | Navigate up in list | `listUp()` + refresh |
| `ListDownEvent` | Navigate down in list | `listDown()` + refresh |

There is no device-update event — device data changes never touch the `Manager` or its event queue (see [Device Data](#device-data) below).

### Screen Types

Defined in `pkg/display/constants.go`:

| Screen | Purpose |
|--------|---------|
| `ScreenStartup` | Boot screen, draws the logo sprite |
| `ScreenLoading` | Loading state with a progress bar and animated spinner |
| `ScreenDeviceStatus` | Shows SBC device data, paginated |
| `ScreenNetworkStatus` | Shows network interface data |
| `ScreenAIInteraction` | AI-related display |
| `ScreenUICharacter` | Draws a character sprite |
| `ScreenError` | Error state display |
| `ScreenMenu` | Menu navigation, implements `TransitionHandler` |

`ScreenTypeCycleOrder` (cycled via the `KeyCycleScreen` key) is: `DeviceStatus → NetworkStatus → AIInteraction → Menu → Loading → Error → Startup`.

### Device Data

Located in `pkg/device/types.go` — entirely separate from `pkg/display`:

```go
type DeviceData interface {
    ID() string
    Type() DeviceType  // sbc, router, mcu
    CPU() float64
    Memory() uint64
    NetworkRx() uint64
    NetworkTx() uint64
    IsOnline() bool  // offline after 20s without update
}
```

**Concrete Types** (all embed `baseDeviceData`):
- `sbcData` - Single Board Computer (Raspberry Pi, Orange Pi, etc.)
- `routerData` - Router with additional `clients` field
- `mcuData` - Microcontroller

`pkg/device/registry.go` keeps its own `xsync.Map[string, DeviceData]` keyed by device ID, with `Register()`, `Get()`, `GetByType()`, and `All()`. `HandleMessage()` is wired directly as the RabbitMQ consumer callback in `main.go` — it unmarshals a `DeviceMessage`, calls `CreateDevice()`, and registers the result. Screens (e.g. the `device` screen) call `device.GetByType(...)` themselves during `Render()`.

### Screen Registry

Screens self-register via `init()` functions:

```go
// In pkg/display/screen/device/device.go
func init() {
    display.Register(display.ScreenDeviceStatus, New())
}
```

`pkg/display/screen/all.go` blank-imports every screen package so their `init()` functions run:

```go
import (
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/ai"
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/character"
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/device"
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/error"
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/loading"
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/menu"
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/network"
    _ "github.com/robotjoosen/go-display-driver/pkg/display/screen/startup"
)
```

### Screen Interface

```go
type Screen interface {
    Render(display int, m *Manager) image.Image
}
```

Screens are stateless singletons — they pull their own data from `m.GetState(display).Data` (type-asserted to their own `*Data` struct) rather than receiving it as a parameter. Per-display state lives in `Manager`'s stores, not in the screen struct.

### TransitionHandler

Screens that handle user selection implement `TransitionHandler`:

```go
type TransitionHandler interface {
    HandleSelect(display int, m *Manager)
}
```

Currently only the `menu` screen implements this.

## Data Flow

### Device Update Flow
```
RabbitMQ → device.HandleMessage() → CreateDevice() → device registry (by ID)
```
This never touches `Manager` — no event is queued, no refresh is triggered. Screens read the registry directly on their next `Render()`.

### Display Refresh Flow
```
RefreshEvent → eventLoop()
            → queueRefresh(display)   [debounced by RefreshDebounceMs]
            → render(display)
            → Get(screenType) → Screen.Render(display, m)
            → panel.DisplayDraw(display, image)
```

### User Navigation Flow
```
RabbitMQ → HandleControlInstructions() → ControlEvent → eventLoop()
                                                        → handleControl(keyID)
                                                        → selectNext/Prev, cycleScreenType, listUp/listDown, or HandleSelect
                                                        → queueRefresh(selectedDisplay)
```

## Manager API

```go
// Creation (starts the eventLoop goroutine)
NewManager(displays []int, p Panel) *Manager

// Event input (non-blocking; drops the event if the queue is full)
Input(e Event)

// Display management
SetScreen(display int, screenType ScreenType, data any)
GetState(display int) (DisplayState, bool)
DisplayList() []int

// List navigation (used by paginated screens like device status)
SetListLength(display int, length int)
SetListIndex(display int, index int)

// Lifecycle
Close()  // stops eventLoop goroutine
```

Device management is **not** part of the Manager API — use `pkg/device` directly (`device.Get`, `device.GetByType`, `device.All`, `device.Register`).

## Selection Highlight

When a display is the currently-selected one and was selected within the last second, `render()` draws a border around the frame before handing it to the panel:

```go
showBorder := display == m.selectedDisplay() &&
    time.Since(m.lastSelectionChange) < 1*time.Second

if showBorder {
    draw.Rectangle(img.(*image.Gray), 0, 0, 128, 64)
}
```

## Offline Detection

Device offline status is checked via `IsOnline()`:

```go
func (b *baseDeviceData) IsOnline() bool {
    return time.Since(b.lastSeen) < 20*time.Second
}
```

`lastSeen` is set when the device is constructed in `CreateDevice()` — receiving a new message effectively refreshes it.

## Key Constants

Defined in `pkg/display/constants.go`:

```go
const (
    KeyCycleDisplay = 7  // Cycle through displays
    KeyCycleScreen  = 6  // Cycle screen types on current display
    KeyPrev         = 4  // List down (screen-dependent)
    KeyNext         = 5  // List up (screen-dependent)
    KeySelect       = 3  // Select/confirm/enter

    ActionRelease = "RELEASE"

    RefreshDebounceMs = 100  // Min ms between renders
)

var ScreenTypeCycleOrder = []ScreenType{
    ScreenDeviceStatus,
    ScreenNetworkStatus,
    ScreenAIInteraction,
    ScreenMenu,
    ScreenLoading,
    ScreenError,
    ScreenStartup,
}
```

## Adding New Screen Types

1. Create a package under `pkg/display/screen/<name>/`
2. Add a `types.go` with a `<Name>Data` struct for the screen's data
3. Implement `Screen` with `Render(display int, m *Manager) image.Image`, reading data via `m.GetState(display).Data`
4. Optionally implement `TransitionHandler` for selection handling
5. Add `init()` to register: `display.Register(display.ScreenXXX, New())`
6. Add the `ScreenType` constant (and `ScreenTypeCycleOrder` entry if it should cycle) in `pkg/display/constants.go`
7. Blank-import the new package in `pkg/display/screen/all.go`

## Adding New Device Types

All in `pkg/device/types.go` — this package has nothing to do with screens:

1. Add a `DeviceType` constant
2. Create a struct embedding `baseDeviceData`
3. Implement the `DeviceData` interface (`ID()`, `Type()`)
4. Add a factory function (`NewXxxData(...)`)
5. Wire it into the `CreateDevice()` switch statement

## Important Notes

- **Device data never triggers a refresh or touches `Manager`** — it's a fully separate registry that screens poll on render
- **`Manager` tracks only per-display UI state and animation state**, never device data
- **Screens are stateless singletons** registered once via `init()`; all mutable state lives in `Manager`'s `Store[T]` instances
- **`eventLoop` runs in its own goroutine** — always call `Close()` before shutdown

## Deployment

Distributed the same way as [`minilab-agent`](https://github.com/robotjoosen/minilab-agent):
[`.github/workflows/release.yml`](.github/workflows/release.yml) builds `linux/arm64` and
`linux/arm` binaries and publishes them as GitHub Release assets (`display-driver-linux-arm64`,
`display-driver-linux-arm`) on every `v*` tag. `scripts/install.sh`, `scripts/update.sh`, and
`scripts/uninstall.sh` download the matching asset and manage the `display_driver` systemd unit
interactively; `task install`/`update`/`uninstall` are thin wrappers around them. There is no
`.env` file at runtime — configuration lives in `Environment=` lines the scripts write into the
systemd unit.
