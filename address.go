package enet

import (
	"unsafe"
)

// #include "enet/enet.h"
import "C"

// Address specifies a portable internet address structure.
type Address interface {
	SetHostAny()
	SetHost(ip string)
	SetPort(port uint16)

	String() string
	GetPort() uint16
}

type enetAddress struct {
	cAddr C.ENetAddress
}

func (addr *enetAddress) SetHostAny() {
	//TODO: fix ipv6 ENET_HOST_ANY assignment
	addr.SetHost("::")
}

func (addr *enetAddress) SetHost(hostname string) {
	cHostname := C.CString(hostname)
	C.enet_address_set_hostname(
		&addr.cAddr,
		cHostname,
	)
	C.free(unsafe.Pointer(cHostname))
}

func (addr *enetAddress) SetPort(port uint16) {
	addr.cAddr.port = (C.uint16_t)(port)
}

func (addr *enetAddress) String() string {
	buffer := C.malloc(1025)
	C.enet_address_get_ip(
		&addr.cAddr,
		(*C.char)(buffer),
		1025,
	)
	ret := C.GoString((*C.char)(buffer))
	C.free(buffer)
	return ret
}

func (addr *enetAddress) GetPort() uint16 {
	return uint16(addr.cAddr.port)
}

// NewAddress creates a new address
func NewAddress(ip string, port uint16) Address {
	ret := enetAddress{}
	ret.SetHost(ip)
	ret.SetPort(port)
	return &ret
}

// NewListenAddress makes a new address ready for listening on ENET_HOST_ANY
func NewListenAddress(port uint16) Address {
	ret := enetAddress{}
	ret.SetHostAny()
	ret.SetPort(port)
	return &ret
}
