package libnet

type Codec interface {
	Receive() ([]byte, error)
	Send(interface{}) error
	Close() error
}
