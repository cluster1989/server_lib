package libio

import (
	"errors"
	"io"

	"github.com/wuqifei/server_lib/libnet/def"
)

var (
	CodecSendMsgNilError = errors.New("send msg is nil")
)

type ProtoCodec struct {
	isLittleIndian bool
	analyseHandle  PacketSpliter

	// 最大的接收字节数
	MaxRecvBufferSize int

	// 最大的发送字节数
	MaxSendBufferSize int
}

//是否是小端解析
func New(isLittleIndian bool, maxRecvBufferSize, maxSendBufferSize int) *ProtoCodec {
	p := &ProtoCodec{}
	p.isLittleIndian = isLittleIndian
	if p.isLittleIndian {
		p.analyseHandle = SplitByUint16LE
	} else {
		p.analyseHandle = SplitByUint16BE
	}
	p.analyseHandle.MaxRecvBufferSize = maxRecvBufferSize
	p.analyseHandle.MaxSendBufferSize = maxSendBufferSize
	return p
}

type protoCodec struct {
	p      *ProtoCodec
	w      *Writer
	r      *Reader
	closer io.Closer
}

func (j *ProtoCodec) NewConn(rw io.ReadWriter) def.Conn {
	codec := &protoCodec{}
	codec.p = j
	codec.w = NewWriter(rw)
	codec.r = NewReader(rw)
	codec.closer, _ = rw.(io.Closer)
	return codec
}
func (c *protoCodec) Receive() ([]byte, error) {
	data, err := c.r.ReadPacket(&c.p.analyseHandle)
	return data, err
}

func (c *protoCodec) Send(msg interface{}) error {
	if msg == nil {
		return CodecSendMsgNilError
	}
	data := msg.([]byte)
	err := c.w.WritePacket(data, &c.p.analyseHandle)
	return err
}

func (c *protoCodec) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
