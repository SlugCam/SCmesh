// TODO could isolate logic from RoutePackets into separate functions for better
// unit testing.

// The flooding package provides functions for facilitating flood style routing
// in the SCmesh software for the SlugCam project.
package flooding

import (
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
)

const BROADCAST_ID = uint32(0xFFFF)

type OriginationRequest struct {
	TTL           int
	PayloadOffset uint32
	DataHeader    header.DataHeader
	Data          []byte
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

type cacheEntry struct {
	node   uint32 // originating node
	id     uint32 // flooding id
	offset uint32 // payload offset
}

// RoutePackets starts a goroutine that performs flooding routing functions.
// None of the functions in the module check for the existence of a packet
// header or a flooding header, this needs to be done by the routing module that
// uses this function.
// TODO check for used headers?
func RoutePackets(localID uint32, toForward <-chan packet.Packet, toOriginate <-chan OriginationRequest, localOut chan<- packet.Packet, out chan<- packet.Packet) {
	// TODO, should persist?
	encountered := make(map[cacheEntry]bool)
	go func() {
		for {
			select {
			case p := <-toForward:

				test := cacheEntry{
					*p.Header.Source,
					*p.Header.FloodingHeader.PacketId,
					p.Preheader.PayloadOffset,
				}

				if !encountered[test] && *p.Header.Source != localID {
					// Send to local
					localOut <- p
					// Add to cache based on id and offset
					encountered[test] = true
					// Adjust TTL
					forward := processTTL(&p)
					// Forward
					if forward {
						out <- p
					}
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

				// TODO use counter
				p.Header.FloodingHeader.PacketId = proto.Uint32(util.RandomUint32())

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
