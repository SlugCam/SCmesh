package dsr

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
)

// These data structures could be optimized
type router struct {
	localID      NodeID
	routeCache   *routeCache
	sendBuffer   *sendBuffer
	requestTable *requestTable
	out          chan<- packet.Packet
}

func newRouter(localID NodeID, out chan<- packet.Packet) *router {
	r := new(router)
	r.localID = localID
	r.routeCache = newRouteCache()
	r.sendBuffer = newSendBuffer()
	r.requestTable = newRequestTable()
	r.out = out
	return r
}

// The following functions are from section 8.1 (General Packet Processing) of
// RFC4728.

// originate (8.1.1) starts the process of sending a packet. The function
// processes an origination request and either outputs a packet with a source
// route, or initiates a route discovery.
func (r *router) originate(o OriginationRequest) {
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
	log.Info("sendRR")
	r.requestTable.sentRequest(target)
	r.out <- *newRouteRequest(r.localID, target)
	// TODO set timeout

}
func (r *router) requestDiscovery(target NodeID) {
	log.Info("reqdisc")
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

// TODO avoid null dereferences

// processPacket follows the procedure outlined in RFC4728 in section 8.1.4
func (r *router) forward(p packet.Packet) {
	// Make copy of header and DSR header, and make sure packet points to these,
	// we receive a new packet objects down the incoming channel, but this
	// packet is forwarded multiple places, so we should check that referenced
	// data is not modified.
	// TODO make sure this works
	nh := *p.Header
	nd := *p.Header.DsrHeader
	nh.DsrHeader = &nd
	p.Header = &nh

	err := r.processRouteRequest(&p)
	if err != nil {
		log.Error("processPacket: dropping packet:", err)
		return
	}
	r.out <- p
}

// processRouteRequest is specified by section 8.2.2. If an error is returned
// the packet should be dropped.
func (r *router) processRouteRequest(p *packet.Packet) error {
	rr := p.Header.DsrHeader.RouteRequest
	if rr == nil {
		return nil // no route request on packet
	}

	// TODO First cache the route on the route request seen so far

	// Check if we are target
	if *rr.Target == uint32(r.localID) {
		// TODO return route reply
		return nil // Is this right?
	}

	// Check if addresses contains our ip, if so drop packet immediately
	for _, a := range rr.Addresses {
		if a == uint32(r.localID) {
			return errors.New("loop found in route request")
		}
	}

	// Check for route request entry to see if we have seen this route request
	if r.requestTable.hasReceivedRequest(NodeID(*p.Header.Source), NodeID(*rr.Target), *rr.Id) {
		return errors.New("route request already seen")
	}

	// At this point continue processing

	// Add request to cache
	r.requestTable.receivedRequest(NodeID(*p.Header.Source), NodeID(*rr.Target), *rr.Id)
	// Make copy of route request option
	nr := *rr
	p.Header.DsrHeader.RouteRequest = &nr
	nr.Addresses = append(nr.Addresses, uint32(r.localID))

	// TODO Check if we can perform a cached route reply

	// TODO Forward RR
	// TODO BROADCAST JITTER
	return nil

}
