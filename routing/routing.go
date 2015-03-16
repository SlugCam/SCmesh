package routing

import (
	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/routing/dsr"
	"github.com/SlugCam/SCmesh/routing/flooding"
)

type nodeId uint32

// This is the node ID used to mean all nodes.
const BroadcastID = uint32(0xFFFF)

// Router is the structure created when we call RoutePackets. It provides
// methods to originate packets.
type Router struct {
	forwardDSR        chan<- packet.Packet
	originateDSR      chan<- dsr.OriginationRequest
	forwardFlooding   chan<- packet.Packet
	originateFlooding chan<- flooding.OriginationRequest
}

// OriginateDSR sends data using the DSR routing scheme. Not that as with
// flooding, the destination field must be set in the data header for any node
// to process its payload.
func (r *Router) OriginateDSR(dest uint32, dataHeader header.DataHeader, data []byte) {

}

// OriginateFlood sends a flooding packet. Note that the packet will be relayed
// to all nodes, but the destination field must be set in the data header for
// any node to process its payload.
func (r *Router) OriginateFlooding(TTL int, dataHeader header.DataHeader, data []byte) {
	log.Debug("OriginateFlooding called")
	r.originateFlooding <- flooding.OriginationRequest{
		TTL:        TTL,
		DataHeader: dataHeader,
		Data:       data,
	}
}

// RoutePackets is the main pipeline function for the routing package. It is
// responsible for performing all routing functions on packets and outputting
// the resulting packets. First it will see if we are a destination in the
// data destinations list, then it forwards the packet to either the DSR or
// flooding module. It will also strip the flooding header if it sees a DSR
// header.
func RoutePackets(localID uint32, toForward <-chan packet.Packet, destLocal chan<- packet.Packet, out chan<- packet.Packet) *Router {

	r := new(Router)

	// Make subrouters

	// DSR
	forwardDSR := make(chan packet.Packet)
	originateDSR := make(chan dsr.OriginationRequest)
	r.forwardDSR = forwardDSR
	r.originateDSR = originateDSR
	dsr.RoutePackets(localID, forwardDSR, originateDSR, out)

	// Flooding
	forwardFlooding := make(chan packet.Packet)
	originateFlooding := make(chan flooding.OriginationRequest)
	r.forwardFlooding = forwardFlooding
	r.originateFlooding = originateFlooding
	flooding.RoutePackets(localID, forwardFlooding, originateFlooding, out)

	go func() {
		for c := range toForward {

			// Forward local data to destLocal
			dh := c.Header.GetDataHeader()
			if dh != nil {
				for _, d := range dh.GetDestinations() {
					if d == localID || d == BroadcastID {
						destLocal <- c
						break // otherwise could send local more than once
					}
				}
			}

			// Route packet
			if c.Header.GetDsrHeader() != nil {
				// Then this is DSR
				//c.Header.FloodingOptions = nil // Remove
			} else if c.Header.GetFloodingHeader() != nil {
				// Then this is Flooding
			}
		}
	}()

	return r
}
