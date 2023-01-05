package spdy

import (
	"encoding/binary"
	"fmt"
)

const (
	flagSYN uint8 = iota // 握手信号
	flagFIN              // 结束信号
	flagDAT              // 发送数据
)

const (
	sizeofFlag   = 1
	sizeofSid    = 4
	sizeofSize   = 2
	sizeofHeader = sizeofFlag + sizeofSid + sizeofSize
)

type frame struct {
	flag uint8
	sid  uint32
	data []byte
}

func (fm frame) pack() []byte {
	dsz := len(fm.data)
	dat := make([]byte, sizeofHeader+dsz)
	dat[0] = fm.flag
	binary.BigEndian.PutUint32(dat[sizeofFlag:], fm.sid)
	binary.BigEndian.PutUint16(dat[sizeofFlag+sizeofSid:], uint16(dsz))
	copy(dat[sizeofHeader:], fm.data)

	return dat
}

type frameHeader [sizeofHeader]byte

func (fh frameHeader) flag() uint8 {
	return fh[0]
}

func (fh frameHeader) streamID() uint32 {
	return binary.BigEndian.Uint32(fh[sizeofFlag:])
}

func (fh frameHeader) size() uint16 {
	return binary.BigEndian.Uint16(fh[sizeofFlag+sizeofSid:])
}

func (fh frameHeader) String() string {
	flag := fh.flag()
	sid := fh.streamID()
	size := fh.size()
	var str string
	switch flag {
	case flagSYN:
		str = "SYN"
	case flagFIN:
		str = "FIN"
	case flagDAT:
		str = "DAT"
	default:
		str = "ERR"
	}

	return fmt.Sprintf("Frame Flag: %s, StreamID: %d, Datasize: %d", str, sid, size)
}
