package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/crypto"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/util"
)

// TestDecodingEncoding is a functional test of the process of decoding and
// encoding packets to/from wire format.
func TestDecodingEncoding(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	// Build a packet
	p := packet.NewPacket()
	p.Preheader.Receiver = uint16(3)
	//p.Header.Type = proto.Int32(0)
	p.Payload = []byte("Hello world!")
	log.Printf("Original: %+v\n", p)

	// Get the raw packet
	ch := make(chan []byte, 100)
	enc := crypto.NewEncrypter("test")
	p.Pack(enc, ch)
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
	decrypter := crypto.NewEncrypter("/slugcam/key") // Should only be used by one goroutine at a time
	p2, err := c2.Parse(decrypter)
	log.Printf("Parsed: %+v\n", p2)
	log.Printf("Parse err: %+v\n", err)

}
