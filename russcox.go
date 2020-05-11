package main

// Thank you, Russ Cox
// https://groups.google.com/forum/#!topic/golang-nuts/3GEzwKfRRQw

import (
	"encoding/binary"
	"unsafe"
)

var NativeEndian binary.ByteOrder

func init() {
	i := uint32(1)
	b := (*[4]byte)(unsafe.Pointer(&i))
	if b[0] == 1 {
		NativeEndian = binary.LittleEndian
	} else {
		NativeEndian = binary.BigEndian
	}
}
