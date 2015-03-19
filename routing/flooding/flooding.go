// TODO could isolate logic from RoutePackets into separate functions for better
// unit testing.

// The flooding package provides functions for facilitating flood style routing
// in the SCmesh software for the SlugCam project.
package flooding

import (
	"crypto/rand"
	"encoding/binary"
	"log"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/golang/protobuf/proto"
)

const BROADCAST_ID = uint32(0xFFFF)

type OriginationRequest struct {
	TTL           int
	PayloadOffset uint32
	DataHeader    header.DataHeader
	Data          []byte
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

//TODO needs testing, better way of copying struct?

// processTTL takes the packet and decrements the TTL. It returns a boolean
// indicating whether the packet should be forwarded or not (true means
// forward). This function will ensure to only change values on new structs
// instead of modifying existing structs to avoid modifying values that may be
// in use concurrently in other parts of the system. However, p should not be in
// use anywhere in the system (but the header, etc. may be).  TODO should TTL be
// optional?
func processTTL(p *packet.Packet) bool {
	if p.Header.Ttl == nil {
		return true
	}

	newHeader := *p.Header
	newTTL := *newHeader.Ttl
	newTTL = newTTL - 1
	newHeader.Ttl = &newTTL

	if newTTL > uint32(0) {
		return true
	} else {
		return false
	}
}

// RoutePackets starts a goroutine that performs flooding routing functions.
// None of the functions in the module check for the existence of a packet
// header or a flooding header, this needs to be done by the routing module that
// uses this function.
// TODO check for used headers?
func RoutePackets(localID uint32, toForward <-chan packet.Packet, toOriginate <-chan OriginationRequest, out chan<- packet.Packet) {
	// TODO, should persist?
	//encountered := make(map[string]bool)
	go func() {
		for {
			select {
			case p := <-toForward:
				// Add to cache based on id and offset

				// Adjust TTL
				forward := processTTL(&p)

				// Forward
				if forward {
					out <- p
				}

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
				p.Preheader.PayloadOffset = origReq.PayloadOffset
				p.Preheader.Receiver = BROADCAST_ID
				// Send to output
				out <- *p
			}

		}

	}()

}
