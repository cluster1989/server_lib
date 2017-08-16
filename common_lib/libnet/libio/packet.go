package libio

import (
	"io"
)

type HeadReadHandle func(r *Reader) int
type HeadWriteHandle func(w *Writer, l int)

type PacketSpliter struct {
	ReadHead  HeadReadHandle
	WriteHead HeadWriteHandle
}

func (p *PacketSpliter) Read(r *Reader) []byte {
	n := p.ReadHead(r)
	if r.Error() != nil {
		return nil
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil
	}
	return b
}

func (p *PacketSpliter) Write(w *Writer, b []byte) {
	p.WriteHead(w, len(b))
	if w.Error() != nil {
		return
	}
	w.Write(b)
}

func (p *PacketSpliter) Limit(r *Reader) *io.LimitedReader {
	n := p.ReadHead(r)
	return &io.LimitedReader{r, int64(n)}
}

var (
	SplitByUint16BE = PacketSpliter{
		ReadHead:  func(r *Reader) int { return int(r.ReadUint16BE()) },
		WriteHead: func(w *Writer, l int) { w.WriteUint16BE(uint16(l)) },
	}
	SplitByUint16LE = PacketSpliter{
		ReadHead:  func(r *Reader) int { return int(r.ReadUint16LE()) },
		WriteHead: func(w *Writer, l int) { w.WriteUint16LE(uint16(l)) },
	}
)
