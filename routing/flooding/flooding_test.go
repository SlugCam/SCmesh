package flooding

import (
	"reflect"
	"testing"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

// This is the node ID used to mean all nodes.
const BroadcastID = uint32(0xFFFFFFFF)

func TestPacketCreation(t *testing.T) {
	localID := uint32(23)
	offset := int64(8907)
	dh := header.DataHeader{
		Destinations: []uint32{BroadcastID},
	}
	data := []byte("TestPacketCreation Payload")
	TTL := 32

	toForward := make(chan packet.Packet)
	out := make(chan packet.Packet)
	local := make(chan packet.Packet)
	toOriginate := make(chan OriginationRequest)

	RoutePackets(localID, toForward, toOriginate, local, out)

	// Originate packet
	toOriginate <- OriginationRequest{
		TTL:           TTL,
		DataHeader:    dh,
		Data:          data,
		PayloadOffset: offset,
	}
	p1 := <-out

	// Originate a second packet
	toOriginate <- OriginationRequest{
		TTL:           TTL,
		DataHeader:    dh,
		Data:          data,
		PayloadOffset: offset,
	}
	p2 := <-out

	if p1.Header.DataHeader == nil || !reflect.DeepEqual(dh, *p1.Header.DataHeader) {
		t.Error("data header was not added to packet correctly.")
	}

	if p1.Header.FloodingHeader == nil {
		t.Error("flooding header not added to packet.")
	}

	if *p1.Header.FloodingHeader.PacketId == *p2.Header.FloodingHeader.PacketId {
		t.Error("packet IDs are not unique")
	}

	if p1.Header.Ttl == nil || *p1.Header.Ttl != uint32(TTL) {
		t.Error("TTL not added to packet")
	}

	if *p1.Header.Source != localID {
		t.Error("source field not set correctly in packet")
	}
	if p1.Preheader.Receiver != BroadcastID {
		t.Error("flooding packets should always have receiver set to broadcast.")
	}
	if p1.Preheader.PayloadOffset != offset {
		t.Error("offset not set correctly in packet.")
	}

}
