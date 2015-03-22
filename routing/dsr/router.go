package dsr

import (
	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
)

// These data structures could be optimized
type router struct {
	routeCache   *routeCache
	sendBuffer   *sendBuffer
	routeRequest map[uint32]requestTableEntry // node id -> table entry
}

func newRouter() *router {
	r := new(router)
	r.routeCache = newRouteCache()
	return r
}

// The following functions are from section 8.1 (General Packet Processing) of
// RFC4728.

// Originate (8.1.1) starts the process of sending a packet. The function
// processes an origination request and outputs a packet that should be sent.
// This packet will either be the packet with the source route option included,
// or a new route request.
func (r *router) originate(o *OriginationRequest) *packet.Packet {
	// Make packet
	packet := newOriginationPacket(o)

	// Check route cache for destination in packet header
	route := r.routeCache.getRoute(o.Destination)
	if route == nil {
		// If no route found perform route discovery
		//r.initiateDiscovery(o.Destination)
		// Add packet to send buffer
		r.sendBuffer.addPacket(packet)
		// Output discovery TODO
	} else {
		// Add source route option to packet
		err := addSourceRoute(packet, route)
		if err != nil {
			log.Error("DSR originate:", err)
		}
		// Output this packet TODO
	}
	return packet
}

func (r *router) newDiscoveryPacket(destination NodeID) {

}

// processPacket follows the procedure outlined in RFC4728 in section 8.1.4
func (r *router) processPacket(p *packet.Packet) {
	processRouteRequest(p)
}

// processRouteRequest
func processRouteRequest(p *packet.Packet) {
	// First cache the route on the route request seen so far

}
