package flooding

import (
	"crypto/rand"
	"encoding/binary"
	"log"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/golang/protobuf/proto"
)

type OriginationRequest struct {
	TTL        int
	DataHeader header.DataHeader
	Data       []byte
}

// TODO use counter
func randomUint32() uint32 {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		log.Panic("Error producing nonce", err)
	}
	return binary.LittleEndian.Uint32(b)
}

func RoutePackets(localID uint32, toForward <-chan packet.Packet, toOriginate <-chan OriginationRequest, out chan<- packet.Packet) {
	// TODO, should persist?
	//encountered := make(map[string]bool)
	go func() {
		for {
			select {
			case c := <-toForward:
				_ = c
				// Add to cache based on id and offset

			case origReq := <-toOriginate:
				// Make new packet
				p := packet.NewPacket()
				// Load data
				p.Header.DataHeader = &origReq.DataHeader
				p.Payload = origReq.Data
				// Make flooding header
				p.Header.Ttl = proto.Uint32(uint32(origReq.TTL))
				p.Header.FloodingHeader = new(header.FloodingHeader)
				p.Header.FloodingHeader.PacketId = proto.Uint32(randomUint32())
				// Assign other required fields
				p.Header.Source = proto.Uint32(localID)
				// Send to output
				out <- *p
			}

		}

	}()

}
