package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/simulation"
	"github.com/SlugCam/SCmesh/util"
)

// InterceptRouter takes a configuration file and returns a channel that will
// eventually provide the router created when a pipeline is run with the
// provided configuration. It does this by wrapping the RoutePackets function in
// the configuration object.
func InterceptRouter(config *pipeline.Config) <-chan pipeline.Router {
	ch := make(chan pipeline.Router, 1)
	log.Info("InterceptRouter")

	prevRoutePackets := config.RoutePackets

	config.RoutePackets = func(localID uint32, toForward <-chan packet.Packet, destLocal chan<- packet.Packet, out chan<- packet.Packet) pipeline.Router {
		log.Info("NewRoutePackets")
		r := prevRoutePackets(localID, toForward, destLocal, out)
		ch <- r
		return r
	}

	return ch
}

// TestFlooding is an integration test for the flooding routing type.
func TestFlooding(t *testing.T) {

	// Setup node 1
	s1 := simulation.StartMockWiFly()
	c1 := DefaultConfig(uint32(1), s1)
	r1ch := InterceptRouter(&c1)

	pipeline.Start(c1)
	log.Info("Started pipeline")
	r1 := <-r1ch

	r1.OriginateFlooding(1, header.DataHeader{FileId: proto.Uint32(0)}, []byte{0})
}

// TestDecodingEncoding is a functional test of the process of decoding and
// encoding packets to/from wire format.
func TestDecodingEncoding(t *testing.T) {
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
