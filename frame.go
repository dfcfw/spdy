package spdy

import (
	"encoding/binary"
	"fmt"
)

const (
	flagNOP uint8 = 0b00000000 // 没有操作
	flagSYN       = 0b00000001 // 握手信号
	flagFIN       = 0b00000010 // 结束信号
	flagPSH       = 0b00000100 // 发送数据

	sizeofFlag   = 1
	sizeofSid    = 4
	sizeofSize   = 2
	sizeofHeader = sizeofFlag + sizeofSid + sizeofSize
)

type frame struct {
	flag uint8
	sid  uint32
	size uint16
	data []byte
}

func (fm frame) pack() []byte {
	return nil
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
	if flag&flagSYN == flagSYN {
		str += "SYN|"
	}
	if flag&flagFIN == flagFIN {
		str += "FIN|"
	}
	if flag&flagPSH == flagPSH {
		str += "PSH|"
	}
	if sz := len(str); sz > 0 {
		str = str[:sz-1]
	} else {
		if flag == flagNOP {
			str = "NOP"
		} else {
			str = "UNKNOWN"
		}
	}

	return fmt.Sprintf("Frame Flag: %s, StreamID: %d, Datasize: %d", str, sid, size)
}
