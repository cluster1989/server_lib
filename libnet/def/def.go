package def

import "io"

type Protocol interface {
	NewConn(rw io.ReadWriter) Conn
}

type Conn interface {
	Receive() ([]byte, error)
	Send(interface{}) error
	Close() error
}

type LibnetMessage struct {
	MsgID   uint16
	Content []byte
}
