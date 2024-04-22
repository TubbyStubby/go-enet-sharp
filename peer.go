package enet

// #include "enet.h"
import "C"
import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

// Peer is a peer which data packets may be sent or received from
type Peer interface {
	GetAddress() Address

	Disconnect(data uint32)
	DisconnectNow(data uint32)
	DisconnectLater(data uint32)

	// Sets a timeout parameters for a peer. The timeout parameters control how and
	// when a peer will timeout from a failure to acknowledge reliable traffic.
	// Timeout values used in the semi-linear mechanism, where if a reliable packet
	// is not acknowledged within an average round-trip time plus a variance tolerance
	// until timeout reaches a set limit. If the timeout is thus at this limit and reliable
	// packets have been sent but not acknowledged within a certain minimum time period, the peer
	// will be disconnected. Alternatively, if reliable packets have been sent but not
	// acknowledged for a certain maximum time period, the peer will be disconnected regardless
	// of the current timeout limit value.
	SetTimeout(limit uint32, min uint32, max uint32)

	SendBytes(data []byte, channel uint8, flags PacketFlags) error
	SendString(str string, channel uint8, flags PacketFlags) error
	SendPacket(packet Packet, channel uint8) error

	// SetData sets an arbitrary value against a peer. This is useful to attach some
	// application-specific data for future use, such as an identifier.
	//
	// http://enet.bespin.org/structENetPeer.html#a1873959810db7ac7a02da90469ee384e
	//
	// Note that due to the way the enet library works, if using this you are
	// responsible for clearing this data when the peer is finished with.
	// SetData(nil) will free underlying memory and avoid any leaks.
	//
	// See http://enet.bespin.org/Tutorial.html#ManageHost for an example of this
	// in the underlying library.
	SetData(data []byte)

	// GetData returns an application-specific value that's been set
	// against this peer. This returns nil if no data has been set.
	//
	// http://enet.bespin.org/structENetPeer.html#a1873959810db7ac7a02da90469ee384e
	GetData() []byte

	PingInterval(intterval uint32)

	GetBytesSent() uint64
	GetBytesReceived() uint64
	GetPacketsSent() uint64
	GetPacketsLost() uint64
}

type enetPeer struct {
	cPeer *C.ENetPeer
}

func (peer enetPeer) GetAddress() Address {
	return &enetAddress{
		cAddr: peer.cPeer.address,
	}
}

func (peer enetPeer) Disconnect(data uint32) {
	C.enet_peer_disconnect(
		peer.cPeer,
		(C.uint32_t)(data),
	)
}

func (peer enetPeer) DisconnectNow(data uint32) {
	C.enet_peer_disconnect_now(
		peer.cPeer,
		(C.uint32_t)(data),
	)
}

func (peer enetPeer) DisconnectLater(data uint32) {
	C.enet_peer_disconnect_later(
		peer.cPeer,
		(C.uint32_t)(data),
	)
}

func (peer enetPeer) SetTimeout(limit uint32, min uint32, max uint32) {
	C.enet_peer_timeout(
		peer.cPeer,
		(C.uint32_t)(limit),
		(C.uint32_t)(min),
		(C.uint32_t)(max),
	)
}

func (peer enetPeer) SendBytes(data []byte, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket(data, flags)
	if err != nil {
		return err
	}
	return peer.SendPacket(packet, channel)
}

func (peer enetPeer) SendString(str string, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket([]byte(str), flags)
	if err != nil {
		return err
	}
	return peer.SendPacket(packet, channel)
}

func (peer enetPeer) SendPacket(packet Packet, channel uint8) error {
	C.enet_peer_send(
		peer.cPeer,
		(C.uint8_t)(channel),
		packet.(enetPacket).cPacket,
	)
	return nil
}

func (peer enetPeer) SetData(data []byte) {
	if len(data) > math.MaxUint32 {
		panic(fmt.Sprintf("maximum peer data length is uint32 (%d)", math.MaxUint32))
	}

	// Free any data that was previously stored against this peer.
	existing := unsafe.Pointer(peer.cPeer.data)
	if existing != nil {
		C.free(existing)
	}

	// If nil, set this explicitly.
	if data == nil {
		peer.cPeer.data = nil
		return
	}

	// First 4 bytes stores how many bytes we have. This is so we can C.GoBytes when
	// retrieving which requires a byte length to read.
	b := make([]byte, len(data)+4)
	binary.LittleEndian.PutUint32(b, uint32(len(data)))
	// Join this header + data in to a contiguous slice
	copy(b[4:], data)
	// And write it out to C memory, storing our pointer.
	peer.cPeer.data = unsafe.Pointer(C.CBytes(b))
}

func (peer enetPeer) GetData() []byte {
	ptr := unsafe.Pointer(peer.cPeer.data)

	if ptr == nil {
		return nil
	}

	// First 4 bytes are the bytes length.
	header := []byte{
		*(*byte)(unsafe.Add(ptr, 0)),
		*(*byte)(unsafe.Add(ptr, 1)),
		*(*byte)(unsafe.Add(ptr, 2)),
		*(*byte)(unsafe.Add(ptr, 3)),
	}

	return []byte(C.GoBytes(
		// Take from the start of the data.
		unsafe.Add(ptr, 4),
		// As many bytes as were indicated in the header.
		C.int(binary.LittleEndian.Uint32(header)),
	))
}

func (peer enetPeer) PingInterval(interval uint32) {
	C.enet_peer_ping_interval(
		peer.cPeer,
		(C.uint32_t)(interval),
	)
}

func (peer enetPeer) GetBytesSent() uint64 {
	return uint64(C.enet_peer_get_bytes_sent(peer.cPeer))
}

func (peer enetPeer) GetPacketsSent() uint64 {
	return uint64(C.enet_peer_get_packets_sent(peer.cPeer))
}

func (peer enetPeer) GetBytesReceived() uint64 {
	return uint64(C.enet_peer_get_bytes_received(peer.cPeer))
}

func (peer enetPeer) GetPacketsLost() uint64 {
	return uint64(C.enet_peer_get_packets_lost(peer.cPeer))
}
