package main

import (
	"fmt"
	"testing"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/crypto"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
)

func TestDecodingEncoding(t *testing.T) {
	p := new(packet.Packet)
	p.Header = new(header.Header)
	p.Header.Type = proto.Int32(0)
	fmt.Printf("Original: %+v\n", p)

	ch := make(chan []byte, 100)
	enc := crypto.NewEncrypter("test")
	p.Pack(enc, ch)
	c := <-ch
	fmt.Println("Encoded Packet:", c)

	// Scan the raw packet (c)
	m := util.NewMockReader()
	m.Write(c)
	rawpacks := make(chan packet.RawPacket, 5)
	prefilter.Prefilter(m, rawpacks)
	c2 := <-rawpacks
	fmt.Printf("Prefiltered: %+v\n", c2)

}
