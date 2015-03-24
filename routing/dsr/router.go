package dsr

import (
	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/golang/protobuf/proto"
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

	cont := r.processRouteRequest(&p)
	if !cont {
		return
	}
}

func (r *router) sendAlongSourceRoute(p *packet.Packet) {
	if processSourceRoute(p) {
		log.Info("processRouteRequest: sending:", p)
		r.out <- *p
	}
}

// TODO check if header exists
// pg62
// Returns true if we should forward, false if we do not forward
func processSourceRoute(p *packet.Packet) bool {
	// TODO more stuff
	// decrement segments left
	sr := p.Header.DsrHeader.SourceRoute
	segsLeft := *sr.SegsLeft

	if segsLeft == 0 {
		if p.Header.Destination == nil {
			return false
		}
		p.Header.DsrHeader.SourceRoute = nil         // remove header
		p.Preheader.Receiver = *p.Header.Destination // last link
		return true
	}

	n := uint32(len(sr.Addresses))

	if segsLeft > n {
		// TODO send error (in spec ICMP error)
		return false
	}

	segsLeft = segsLeft - 1
	sr.SegsLeft = proto.Uint32(segsLeft)

	i := n - segsLeft - 1 // The minus one is because our array is 0 based, rfc describes 1 based

	p.Preheader.Receiver = sr.Addresses[i]
	if p.Preheader.Receiver == uint32(BROADCAST_ID) {
		// No address can be a multicast address
		return false
	}
	// TODO should this be here? this is where the spec is
	if p.Header.Destination == nil || *p.Header.Destination == uint32(BROADCAST_ID) {
		return false
	}
	// TODO process TTL
	// TODO more stuff
	// route maintainence
	return true
}

// processRouteRequest is specified by section 8.2.2.
// True means continue processing
func (r *router) processRouteRequest(p *packet.Packet) bool {
	rr := p.Header.DsrHeader.RouteRequest
	if rr == nil {
		return true // no route request on packet
	}

	// TODO First cache the route on the route request seen so far

	// Check if we are target
	if *rr.Target == uint32(r.localID) {
		reply := newRouteReply(rr.Addresses, *p.Header.Source, uint32(r.localID))
		r.sendAlongSourceRoute(reply)
		return true // Is this right?
	}

	// Check if addresses contains our ip, if so drop packet immediately
	if *p.Header.Source == uint32(r.localID) {
		return false
	}
	for _, a := range rr.Addresses {
		if a == uint32(r.localID) {
			log.Error("loop found in route request")
			return false
		}
	}

	// Check for route request entry to see if we have seen this route request
	if r.requestTable.hasReceivedRequest(NodeID(*p.Header.Source), NodeID(*rr.Target), *rr.Id) {
		return false // route request already seen
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
	r.out <- *p
	return false

}
