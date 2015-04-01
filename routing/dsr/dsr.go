package dsr

import (
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

type NodeID uint32
type Route []NodeID

const BROADCAST_ID = NodeID(0xFFFF)

// OriginationRequest is a struct used to describe a new packet we should
// originate. They are used in RoutePackets to provide communication to this
// module from outside.
type OriginationRequest struct {
	Destination NodeID
	Offset      uint32
	DataHeader  header.DataHeader
	Data        []byte
}

// RoutePackets is the main pipeline function that creates a DSR router and
// manages packet origination and forwarding.
func RoutePackets(localID uint32, toForward <-chan packet.Packet, toOriginate <-chan OriginationRequest, out chan<- packet.Packet) {

	r := newRouter(NodeID(localID), out)

	go func() {
		for {
			select {
			case p := <-toForward:
				r.forward(p)
			case o := <-toOriginate:
				r.originate(o)
			}

		}

	}()

}
