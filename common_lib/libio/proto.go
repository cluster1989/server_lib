package libio

import (
	"io"

	"github.com/wqf/common_lib/libnet/def"
)

type ProtoCodec struct {
	isLittleIndian bool
	analyseHandle  PacketSpliter
}

//是否是小端解析
func New(isLittleIndian bool) *ProtoCodec {
	p := &ProtoCodec{}
	p.isLittleIndian = isLittleIndian
	if p.isLittleIndian {
		p.analyseHandle = SplitByUint16LE
	} else {
		p.analyseHandle = SplitByUint16BE
	}
	return p
}

type protoCodec struct {
	p      *ProtoCodec
	w      *Writer
	r      *Reader
	closer io.Closer
}

func (j *ProtoCodec) NewCodec(rw io.ReadWriter) def.Codec {
	codec := &protoCodec{}
	codec.p = j
	codec.w = NewWriter(rw)
	codec.r = NewReader(rw)
	codec.closer, _ = rw.(io.Closer)
	return codec
}
func (c *protoCodec) Receive() ([]byte, error) {
	data := c.r.ReadPacket(&c.p.analyseHandle)
	return data, nil
}

func (c *protoCodec) Send(msg interface{}) error {
	data := msg.([]byte)
	c.w.WritePacket(data, &c.p.analyseHandle)
	return nil
}

func (c *protoCodec) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
