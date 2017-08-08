package zone_server

import (
	"github.com/wqf/common_lib/codec"
)

//组包
func Packet(w *codec.Writer, l int) {
	w.WriteUint16BE(uint16(l))
}

//解包
func Unpack(r *codec.Reader) int {
	length := r.ReadInt16BE() //读取总长度
	return int(length)
}
