package tcp_server

import (
	"errors"
	"io"

	"github.com/wuqifei/server_lib/libio"
)

var (
	CodecMsgSendBufferTooLong = errors.New("msg send buffer too long")
	CodecMsgRecvBufferTooLong = errors.New("msg recv buffer too long")
)

type HeadReadHandle func(r *libio.Reader) int
type HeadWriteHandle func(w *libio.Writer, l int)

type PacketSpliter struct {
	ReadHead  HeadReadHandle
	WriteHead HeadWriteHandle

	// 最大的接收字节数
	MaxRecvBufferSize int

	// 最大的发送字节数
	MaxSendBufferSize int
}

func (p *PacketSpliter) Read(r *libio.Reader) ([]byte, error) {
	n := p.ReadHead(r)
	//这里如果，字节树过长

	if n > p.MaxRecvBufferSize {
		return nil, CodecMsgRecvBufferTooLong
	}
	if r.Error() != nil {
		return nil, r.Error()
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (p *PacketSpliter) Write(w *libio.Writer, b []byte) error {
	length := len(b)
	if length > p.MaxSendBufferSize {
		return CodecMsgSendBufferTooLong
	}
	p.WriteHead(w, length)

	if w.Error() != nil {
		return w.Error()
	}
	w.Write(b)
	return nil
}

func (p *PacketSpliter) Limit(r *libio.Reader) *io.LimitedReader {
	n := p.ReadHead(r)
	return &io.LimitedReader{r, int64(n)}
}

var (
	SplitByUint16BE = PacketSpliter{
		ReadHead:  func(r *libio.Reader) int { return int(r.ReadUint16BE()) },
		WriteHead: func(w *libio.Writer, l int) { w.WriteUint16BE(uint16(l)) },
	}
	SplitByUint16LE = PacketSpliter{
		ReadHead:  func(r *libio.Reader) int { return int(r.ReadUint16LE()) },
		WriteHead: func(w *libio.Writer, l int) { w.WriteUint16LE(uint16(l)) },
	}
)
