package libnet

import "io"

type Protocol interface {
	NewCodec(rw io.ReadWriter) Codec
}
