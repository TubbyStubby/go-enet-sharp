package enet

// #include "enet.h"
import "C"
import (
	"errors"
)

// Host for communicating with peers
type Host interface {
	Destroy()
	Service(timeout uint32) Event
	ServiceV2(event *enetEvent, timeout uint32) int

	Connect(addr Address, channelCount int, data uint32) (Peer, error)

	BroadcastBytes(data []byte, channel uint8, flags PacketFlags) error
	BroadcastPacket(packet Packet, channel uint8) error
	BroadcastString(str string, channel uint8, flags PacketFlags) error

	GetBytesSent() uint32
	GetBytesReceived() uint32
	GetPacketsSent() uint32
	GetPacketsReceived() uint32

	ResetBytesSent()
	ResetBytesReceived()
	ResetPacketsSent()
	ResetPacketsReceived()
}

type enetHost struct {
	cHost *C.ENetHost
}

func (host *enetHost) Destroy() {
	C.enet_host_destroy(host.cHost)
}

func (host *enetHost) Service(timeout uint32) Event {
	ret := &enetEvent{}
	C.enet_host_service(
		host.cHost,
		&ret.cEvent,
		(C.uint32_t)(timeout),
	)
	return ret
}

func (host *enetHost) ServiceV2(event *enetEvent, timeout uint32) int {
	ret := C.enet_host_service(
		host.cHost,
		&event.cEvent,
		(C.uint32_t)(timeout),
	)
	return int(ret)
}

func (host *enetHost) Connect(addr Address, channelCount int, data uint32) (Peer, error) {
	peer := C.enet_host_connect(
		host.cHost,
		&(addr.(*enetAddress)).cAddr,
		(C.size_t)(channelCount),
		(C.uint32_t)(data),
	)

	if peer == nil {
		return nil, errors.New("couldn't connect to foreign peer")
	}

	return enetPeer{
		cPeer: peer,
	}, nil
}

// NewHost creats a host for communicating to peers
func NewHost(addr Address, peerCount, channelLimit uint64, incomingBandwidth, outgoingBandwidth uint32, bufferLimit int) (Host, error) {
	var cAddr *C.ENetAddress
	if addr != nil {
		cAddr = &(addr.(*enetAddress)).cAddr
	}

	host := C.enet_host_create(
		cAddr,
		(C.size_t)(peerCount),
		(C.size_t)(channelLimit),
		(C.uint32_t)(incomingBandwidth),
		(C.uint32_t)(outgoingBandwidth),
		(C.int)(bufferLimit),
	)

	if host == nil {
		return nil, errors.New("unable to create host")
	}

	return &enetHost{
		cHost: host,
	}, nil
}

func (host *enetHost) BroadcastBytes(data []byte, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket(data, flags)
	if err != nil {
		return err
	}
	return host.BroadcastPacket(packet, channel)
}

func (host *enetHost) BroadcastPacket(packet Packet, channel uint8) error {
	C.enet_host_broadcast(
		host.cHost,
		(C.uint8_t)(channel),
		packet.(enetPacket).cPacket,
	)
	return nil
}

func (host *enetHost) BroadcastString(str string, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket([]byte(str), flags)
	if err != nil {
		return err
	}
	return host.BroadcastPacket(packet, channel)
}

func (host *enetHost) GetBytesSent() uint32 {
	return uint32(C.enet_host_get_bytes_sent(host.cHost))
}

func (host *enetHost) GetPacketsSent() uint32 {
	return uint32(C.enet_host_get_packets_sent(host.cHost))
}

func (host *enetHost) GetBytesReceived() uint32 {
	return uint32(C.enet_host_get_bytes_received(host.cHost))
}

func (host *enetHost) GetPacketsReceived() uint32 {
	return uint32(C.enet_host_get_packets_received(host.cHost))
}

func (host *enetHost) ResetBytesSent() {
	host.cHost.totalSentData = 0
}

func (host *enetHost) ResetBytesReceived() {
	host.cHost.totalReceivedData = 0
}

func (host *enetHost) ResetPacketsSent() {
	host.cHost.totalSentPackets = 0
}

func (host *enetHost) ResetPacketsReceived() {
	host.cHost.totalReceivedPackets = 0
}
