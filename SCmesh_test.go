package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/routing"
	"github.com/SlugCam/SCmesh/simulation"
	"github.com/SlugCam/SCmesh/util"
)

// TestFlooding is an integration test for the flooding routing type.
func TestFlooding(t *testing.T) {

	// Setup node 1
	s1 := simulation.StartMockWiFly()
	c1 := DefaultConfig(uint32(1), s1)
	r1ch := InterceptRouter(&c1)
	pipeline.Start(c1)
	r1 := <-r1ch // r1 is now the router for node 1

	// Setup node 2
	s2 := simulation.StartMockWiFly()
	c2 := DefaultConfig(uint32(2), s2)
	l2ch := InterceptLocal(&c2) // l2ch is the local packets for node 2
	pipeline.Start(c2)

	// Setup node 3
	s3 := simulation.StartMockWiFly()
	c3 := DefaultConfig(uint32(3), s3)
	l3ch := InterceptLocal(&c3) // l2ch is the local packets for node 2
	pipeline.Start(c3)

	// Link nodes
	s1.Link(s2)
	s2.Link(s3)

	// Send packet 1 from node 1
	dh := header.DataHeader{
		FileId:       proto.Uint32(1),
		Destinations: []uint32{routing.BroadcastID},
	}

	r1.OriginateFlooding(1, dh, []byte{0})

	log.Info(<-l2ch)

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
