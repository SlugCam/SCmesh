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
	out          chan<- packet.Packet
}

func newRouter() *router {
	r := new(router)
	r.routeCache = newRouteCache()
	r.sendBuffer = newSendBuffer()
	return r
}

// The following functions are from section 8.1 (General Packet Processing) of
// RFC4728.

// originate (8.1.1) starts the process of sending a packet. The function
// processes an origination request and either outputs a packet with a source
// route, or initiates a route discovery.
func (r *router) originate(o *OriginationRequest) {
	// Make packet
	packet := newOriginationPacket(o)
	// Check route cache for destination in packet header
	route := r.routeCache.getRoute(o.Destination)
	if route == nil {
		// If no route found perform route discovery
		r.sendBuffer.addPacket(packet)
		// TODO initiateRouteDiscovery
		r.initiateRouteDiscovery(o.Destination)
	} else {
		// Otherwise add source route option to packet
		err := addSourceRoute(packet, route)
		if err != nil {
			log.Error("DSR originate:", err)
		}
		// Output packet
		r.out <- *packet
	}
}

func (r *router) initiateRouteDiscovery(dest NodeID) {

}

// processPacket follows the procedure outlined in RFC4728 in section 8.1.4
func (r *router) processPacket(p *packet.Packet) {
	processRouteRequest(p)
}

// processRouteRequest
func processRouteRequest(p *packet.Packet) {
	// First cache the route on the route request seen so far

}
