package system

import (
	"encoding/binary"
	"unsafe"
)

var NativeEndian binary.ByteOrder

// Check the endianess of the running machine
func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		NativeEndian = binary.LittleEndian

	case [2]byte{0xAB, 0xCD}:
		NativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}
