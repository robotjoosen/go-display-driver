package network

type NetworkInfo struct {
	Name string
	Rx   uint64
	Tx   uint64
}

type NetworkStatusData struct {
	ID         string
	Interfaces []NetworkInfo
}
