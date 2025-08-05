package panel

type ChannelAware interface {
	SetChannel(int) error
}
