package codec

import (
	"io"
	"time"

	"github.com/wqf/common_lib/libnet"
)

type HeadReadHandle func(r *Reader) int
type HeadWriteHandle func(w *Writer, l int)

type ProtobufProtoCol struct {
	ReadHead  HeadReadHandle
	WriteHead HeadWriteHandle
}

func Protobuf(readHead HeadReadHandle, writeHead HeadWriteHandle) *ProtobufProtoCol {
	p := &ProtobufProtoCol{}
	p.ReadHead = readHead
	p.WriteHead = writeHead
	return p
}

type protobufCodec struct {
	p            *ProtobufProtoCol
	w            *Writer
	r            *Reader
	h            *HeadSpliter
	readTimeOut  time.Duration
	writeTimeOut time.Duration
	closer       io.Closer
}

func (j *ProtobufProtoCol) NewCodec(rw io.ReadWriter) libnet.Codec {
	codec := &protobufCodec{}
	codec.p = j
	codec.w = NewWriter(rw)
	codec.r = NewReader(rw)
	codec.h = &HeadSpliter{
		ReadHead:  j.ReadHead,
		WriteHead: j.WriteHead,
	}
	codec.closer, _ = rw.(io.Closer)
	return codec
}
func (c *protobufCodec) Receive() ([]byte, error) {
	data := c.r.ReadPacket(c.h)
	return data, nil
}

func (c *protobufCodec) Send(msg interface{}) error {
	data := msg.([]byte)
	c.w.WritePacket(data, c.h)
	return nil
}
func (c *protobufCodec) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
