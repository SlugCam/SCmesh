package main

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/routing"
	"github.com/SlugCam/SCmesh/simulation"
	"github.com/SlugCam/SCmesh/util"
)

// readChannel simply reads a packet from the channel or returns nil if no
// packet is ready.
func readChannel(ch <-chan packet.Packet) *packet.Packet {
	select {
	case c := <-ch:
		return &c
	case <-time.After(1 * time.Second):
		return nil
	}
}

// TestFlooding is an integration test for the flooding routing type.
func TestFlooding(t *testing.T) {

	// Setup node 1
	n1 := simulation.StartNewNode(uint32(1))
	n2 := simulation.StartNewNode(uint32(2))
	n3 := simulation.StartNewNode(uint32(3))

	// Link nodes
	n1.Link(n2)
	n2.Link(n3)

	// Send packet 1 from node 1
	dh := header.DataHeader{
		FileId:       proto.Uint32(1),
		Destinations: []uint32{routing.BroadcastID},
	}

	// t1
	n1.Router.OriginateFlooding(1, dh, []byte{0})

	p := <-n2.IncomingPackets
	// Checks on packet
	if p.Header.FloodingHeader == nil {
		t.Error("packet did not have a flooding header.")
	}
	if p.Preheader.Receiver != routing.BroadcastID {
		t.Error("flooding packets should always have receiver set to broadcast.")
	}
	log.Info(p)

	// Check that no other packets were received
	select {
	case <-n1.IncomingPackets:
		t.Error("received packet on n1 during t1")
	case <-n2.IncomingPackets:
		t.Error("received multiple packets on n2 during t1.")
	case <-n3.IncomingPackets:
		t.Error("received packet on n3 during t1.")
	case <-time.After(1 * time.Second):
	}

}

// TestDecodingEncoding is a functional test of the process of decoding and
// encoding packets to/from wire format.
func TestDecodingEncoding(t *testing.T) {
	t.Skip()
	log.SetLevel(log.DebugLevel)
	// Build a packet
	p := packet.NewPacket()
	p.Preheader.Receiver = uint16(3)
	p.Header.Source = proto.Uint32(1)
	p.Payload = []byte("Hello world!")
	log.Printf("Original: %+v\n", p)

	// Get the raw packet
	ch := make(chan []byte, 100)
	p.Pack(ch)
	c := <-ch
	log.Println("Encoded Packet:", c)
	log.Println("Encoded Length:", len(c))

	// Scan the raw packet (c)
	m := util.NewMockReader()
	m.Write(c)
	rawpacks := make(chan packet.RawPacket, 5)
	prefilter.Prefilter(m, rawpacks)
	c2 := <-rawpacks
	log.Printf("Prefiltered: %+v\n", c2)

	// Parse the packet
	p2, err := c2.Parse()
	log.Printf("Parsed: %+v\n", p2)
	log.Printf("Parse err: %+v\n", err)

}
