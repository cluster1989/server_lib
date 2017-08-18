package def

import "io"

type Protocol interface {
	NewCodec(rw io.ReadWriter) Codec
}

type Codec interface {
	Receive() ([]byte, error)
	Send(interface{}) error
	Close() error
}
