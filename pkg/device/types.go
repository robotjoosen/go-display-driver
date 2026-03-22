package device

import "time"

type DeviceType string

const (
	DeviceTypeSBC    DeviceType = "sbc"
	DeviceTypeRouter DeviceType = "router"
	DeviceTypeMCU    DeviceType = "mcu"
)

type DeviceData interface {
	ID() string
	Type() DeviceType
	CPU() float64
	Memory() uint64
	NetworkRx() uint64
	NetworkTx() uint64
	IsOnline() bool
}

type baseDeviceData struct {
	lastSeen  time.Time
	cpu       float64
	memory    uint64
	networkRx uint64
	networkTx uint64
}

func (b *baseDeviceData) CPU() float64      { return b.cpu }
func (b *baseDeviceData) Memory() uint64    { return b.memory }
func (b *baseDeviceData) NetworkRx() uint64 { return b.networkRx }
func (b *baseDeviceData) NetworkTx() uint64 { return b.networkTx }
func (b *baseDeviceData) IsOnline() bool    { return time.Since(b.lastSeen) < 20*time.Second }

type sbcData struct {
	baseDeviceData
	id string
}

func (s *sbcData) ID() string       { return s.id }
func (s *sbcData) Type() DeviceType { return DeviceTypeSBC }

func NewSBCData(id string, cpu float64, memory, networkRx, networkTx uint64) *sbcData {
	return &sbcData{
		baseDeviceData: baseDeviceData{
			lastSeen:  time.Now(),
			cpu:       cpu,
			memory:    memory,
			networkRx: networkRx,
			networkTx: networkTx,
		},
		id: id,
	}
}

type routerData struct {
	baseDeviceData
	id      string
	clients int
}

func (r *routerData) ID() string       { return r.id }
func (r *routerData) Type() DeviceType { return DeviceTypeRouter }

func NewRouterData(id string, cpu float64, memory, networkRx, networkTx uint64, clients int) *routerData {
	return &routerData{
		baseDeviceData: baseDeviceData{
			lastSeen:  time.Now(),
			cpu:       cpu,
			memory:    memory,
			networkRx: networkRx,
			networkTx: networkTx,
		},
		id:      id,
		clients: clients,
	}
}

type mcuData struct {
	baseDeviceData
	id string
}

func (m *mcuData) ID() string       { return m.id }
func (m *mcuData) Type() DeviceType { return DeviceTypeMCU }

func NewMCUData(id string, cpu float64, memory, networkRx, networkTx uint64) *mcuData {
	return &mcuData{
		baseDeviceData: baseDeviceData{
			lastSeen:  time.Now(),
			cpu:       cpu,
			memory:    memory,
			networkRx: networkRx,
			networkTx: networkTx,
		},
		id: id,
	}
}

type DeviceStatusData = sbcData

func CreateDevice(msg DeviceMessage) DeviceData {
	switch msg.Type {
	case "sbc":
		return NewSBCData(msg.Name, msg.Cpu.System+msg.Cpu.User, msg.Mem.Used, getEth0Rx(msg.Nic), getEth0Tx(msg.Nic))
	case "router":
		return NewRouterData(msg.Name, msg.Cpu.System+msg.Cpu.User, msg.Mem.Used, getEth0Rx(msg.Nic), getEth0Tx(msg.Nic), 0)
	case "mcu":
		return NewMCUData(msg.Name, msg.Cpu.System+msg.Cpu.User, msg.Mem.Used, getEth0Rx(msg.Nic), getEth0Tx(msg.Nic))
	default:
		return NewSBCData(msg.Name, msg.Cpu.System+msg.Cpu.User, msg.Mem.Used, getEth0Rx(msg.Nic), getEth0Tx(msg.Nic))
	}
}

func getEth0Rx(nics []NicMessage) uint64 {
	for _, nic := range nics {
		if nic.Name == "eth0" {
			return nic.Rx
		}
	}
	return 0
}

func getEth0Tx(nics []NicMessage) uint64 {
	for _, nic := range nics {
		if nic.Name == "eth0" {
			return nic.Tx
		}
	}
	return 0
}

type DeviceMessage struct {
	Name string       `json:"name"`
	Type string       `json:"type"`
	Mem  MemMessage   `json:"memory"`
	Cpu  CpuMessage   `json:"cpu"`
	Nic  []NicMessage `json:"network_interfaces"`
}

type MemMessage struct {
	Free uint64 `json:"free"`
	Used uint64 `json:"used"`
}

type CpuMessage struct {
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	User   float64 `json:"user"`
}

type NicMessage struct {
	Name string `json:"name"`
	Rx   uint64 `json:"rx_bytes"`
	Tx   uint64 `json:"tx_bytes"`
}
