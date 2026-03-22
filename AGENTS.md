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
│  │ handleDeviceUpdate│    │  HandleControlInstructions     │  │
│  └────────┬─────────┘    └─────────────┬─────────────────┘  │
│           │                              │                    │
│           ▼                              ▼                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              screens.Manager.Input()                  │   │
│  │                                                       │   │
│  │   Event Queue (buffered channel, 100)                │   │
│  │         │                                              │   │
│  │         ▼                                              │   │
│  │   eventLoop() goroutine                               │   │
│  │         │                                              │   │
│  │         ▼                                              │   │
│  │   handle(Event)                                        │   │
│  └──────────────────────────────────────────────────────┘   │
│                              │                               │
└──────────────────────────────┼───────────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
          ▼                    ▼                    ▼
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │   devices   │    │  displays   │    │   render    │
    │  map[id]    │    │  map[id]    │    │  pipeline   │
    │ DeviceData  │    │ScreenState  │    │             │
    └─────────────┘    └─────────────┘    └─────────────┘
```

### Display Detection

Displays are dynamically detected via the `pkg/discover` package at startup:

```go
displayList := discover.Displays(bus, tcaMux)
```

The `Displays()` function iterates through channels 0-7 on the TCA9548 multiplexer and attempts to detect SSD1306 OLED displays. Only channels with detected displays are included in the final list.

## Key Concepts

### Displays vs Devices

- **Display**: Physical OLED display (0, 1, 2) managed by the TCA9548A multiplexer
- **Device**: A computing device (SBC, Router, MCU) that sends status data via RabbitMQ

**Important**: Display and Device are **independent**. A display shows a screen type, and the screen fetches device data from the Manager's device registry.

### Event System

Events flow through a buffered channel (100) and are processed by the `eventLoop()` goroutine:

| Event | Description | Handler Action |
|-------|-------------|----------------|
| `RefreshEvent` | Request display refresh | `queueRefresh(display)` |
| `ControlEvent` | Keyboard input | `handleControl()` - navigation, selection |
| `UpdateDeviceEvent` | Device data received | `updateDevice()` - store only, no refresh |
| `SelectEvent` | User selected current item | Call `TransitionHandler.HandleSelect()` |
| `NextDisplayEvent` | Navigate to next display | `selectNext()` + refresh |
| `PrevDisplayEvent` | Navigate to prev display | `selectPrev()` + refresh |
| `ListUpEvent` | Navigate up in list | `listUp()` + refresh |
| `ListDownEvent` | Navigate down in list | `listDown()` + refresh |

### Screen Types

| Screen | Purpose |
|--------|---------|
| `ScreenStartup` | Boot/initialization screen |
| `ScreenLoading` | Loading state |
| `ScreenDeviceStatus` | Shows SBC/MCU device data |
| `ScreenNetworkStatus` | Shows router/network data |
| `ScreenAIInteraction` | AI-related display (future) |
| `ScreenUICharacter` | Character display (future) |
| `ScreenError` | Error state display |
| `ScreenMenu` | Menu navigation |

### Device Data Interface

Located in `pkg/screens/device/types.go`:

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

### Screen Registry

Screens self-register via `init()` functions:

```go
// In pkg/screens/device/device.go
func init() {
    screens.Register(screens.ScreenDeviceStatus, New())
}
```

### TransitionHandler

Screens that handle user selection implement `TransitionHandler`:

```go
type TransitionHandler interface {
    HandleSelect(display int, m *Manager)
}
```

## Data Flow

### Device Update Flow
```
RabbitMQ → handleDeviceUpdate() → UpdateDeviceEvent → eventLoop()
                                                    → updateDevice()
                                                    → stored in Manager.devices[]
```

### Display Refresh Flow
```
RefreshEvent → eventLoop()
            → queueRefresh(display)
            → render(display)
            → Get(screenType) → Screen.Render(state.Data)
            → panel.DisplayDraw(display, image)
```

### User Navigation Flow
```
RabbitMQ → HandleControlInstructions() → ControlEvent → eventLoop()
                                                        → handleControl(keyID)
                                                        → selectNext/Prev/listUp/listDown
                                                        → queueRefresh(selectedDisplay)
```

## Manager API

```go
// Creation
NewManager(displays []int, p *panel.Panel) *Manager

// Event input (thread-safe)
Input(e Event)

// Display management
SetScreen(display int, screenType ScreenType, data any)
GetState(display int) (DisplayState, bool)

// Device management
UpdateDevice(data any)  // stores by ID
GetDevice(id string) any

// List navigation
SetListLength(display int, length int)
SetListIndex(display int, index int)

// State persistence (reboot recovery)
LoadState() error  // loads persisted state from ~/.config/go-display-driver/state.json
SaveState() error  // saves current state

// Lifecycle
Close()  // stops eventLoop goroutine
```

## Offline Detection

Device offline status is checked via `IsOnline()`:

```go
func (b *baseDeviceData) IsOnline() bool {
    return time.Since(b.lastSeen) < 20*time.Second
}
```

`lastSeen` is updated in `CreateDevice()` when a device message is received.

## Key Constants

```go
const (
    KeyDown   = 7  // Cycle through displays (forward)
    KeyUp     = 6  // Cycle screen types on current display
    KeyPrev   = 4  // List down (screen-dependent)
    KeyNext   = 5  // List up (screen-dependent)
    KeySelect = 3  // Select/confirm/enter

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
}
```

## Adding New Screen Types

1. Create package under `pkg/screens/<name>/`
2. Implement `Screen` interface with `Render(data any) image.Image`
3. Optionally implement `TransitionHandler` for selection handling
4. Add `init()` to register: `screens.Register(screens.ScreenXXX, New())`
5. Add constant in `pkg/screens/constants.go`

## Adding New Device Types

1. Add `DeviceType` constant in `pkg/screens/device/types.go`
2. Create struct embedding `baseDeviceData`
3. Implement `DeviceData` interface
4. Add factory function
5. Update `CreateDevice()` switch statement

## Important Notes

- **UpdateDeviceEvent does NOT trigger refresh** - data is stored only
- **Displays and devices are independent** - screen fetches device data on render
- **Screens filter devices** - each screen determines which devices it can display
- **eventLoop runs in separate goroutine** - always call `Close()` before shutdown

## State Persistence

On reboot, the following state is restored from `~/.config/go-display-driver/state.json`:

| State | Description |
|-------|-------------|
| `selectedIndex` | Which physical display (0, 1, 2) is selected |
| `displays[d].ScreenType` | What screen type each display is showing |
| `displays[d].ListIndex` | Navigation position per display |

**NOT persisted** (ephemeral):
- `ListLength` - reconstructed when screen loads
- `Data` - ephemeral device/network data

**Auto-save behavior**:
- State is marked dirty on any navigation/display change
- Saves are debounced: 5s after last change, or every 30s if dirty
- Final save occurs on `Close()`