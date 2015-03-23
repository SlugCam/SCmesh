package dsr

import (
	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
)

// These data structures could be optimized
type router struct {
	localID      NodeID
	routeCache   *routeCache
	sendBuffer   *sendBuffer
	requestTable requestTable
	out          chan<- packet.Packet
}

func newRouter(localID NodeID) *router {
	r := new(router)
	r.localID = localID
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
		r.requestDiscovery(o.Destination)
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

// DISCOVERY FUNCTIONS

func (r *router) sendRouteRequest(target NodeID) {
	r.requestTable.sentRequest(target)
	r.out <- *newRouteRequest(r.localID, target)
	// TODO set timeout

}
func (r *router) requestDiscovery(target NodeID) {
	if !r.requestTable.discoveryInProcess(target) {
		r.sendRouteRequest(target)
	}
}
func (r *router) processRouteRequestTimeout(target NodeID) {
	if r.requestTable.discoveryInProcess(target) {
		r.sendRouteRequest(target)
	}
}

// PROCESSING FUNCTIONS

// processPacket follows the procedure outlined in RFC4728 in section 8.1.4
func (r *router) processPacket(p *packet.Packet) {
	r.processRouteRequest(p)
}

// processRouteRequest is specified by section 8.2.2
func (r *router) processRouteRequest(p *packet.Packet) {

	// First cache the route on the route request seen so far

}
