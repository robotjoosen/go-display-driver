package device

type DeviceStatusData struct {
	ID        string
	Online    bool
	CPU       float64
	Memory    uint64
	NetworkRx uint64
	NetworkTx uint64
}
