package dsr

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
)

// These data structures could be optimized
type router struct {
	liveLinks    map[uint32]linkMaint
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
	r.liveLinks = make(map[uint32]linkMaint)
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
	packet.Header.Destination = proto.Uint32(uint32(o.Destination))
	packet.Header.Source = proto.Uint32(uint32(r.localID))
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

// TODO combine both
func (r *router) originatePacket(p *packet.Packet) {
	if p.Header == nil || p.Header.Destination == nil {
		log.Error("originateExisting: improper packet format, dropping packet")
		return
	}
	dest := *p.Header.Destination
	p.Header.Source = proto.Uint32(uint32(r.localID))
	// Check route cache for destination in packet header
	route := r.routeCache.getRoute(NodeID(dest))
	if route == nil {
		// If no route found perform route discovery
		r.sendBuffer.addPacket(p)
		// TODO initiateRouteDiscovery
		r.requestDiscovery(NodeID(dest))
	} else {
		// Otherwise add source route option to packet
		err := addSourceRoute(p, route)
		if err != nil {
			log.Error("DSR originate:", err)
		}
		// Output packet
		r.out <- *p
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
	// TODO process route reply
	r.processRouteReply(&p)
	r.processAckRequest(&p)
	r.processAck(&p)
	r.processRouteError(&p)

	// Only forward if we were the receiver in the first place, broadcast not
	// valid for source route anyway.
	// TODO this should be anywhere where we respond on header options
	if p.Preheader.Receiver == uint32(r.localID) {
		r.sendAlongSourceRoute(&p)
	}
}

func (r *router) sendAlongSourceRoute(p *packet.Packet) {
	if processSourceRoute(p) {
		r.addAckRequest(p)
		log.Info("processRouteRequest: sending:", p)
		r.out <- *p
	}
}

const ERROR_REPORTING_TIMEOUT = 2 * time.Second

func (r *router) addAckRequest(p *packet.Packet) {
	id := util.RandomUint32()
	p.Header.DsrHeader.AckRequest = &header.DSRHeader_AckRequest{
		Identification: proto.Uint32(id),
		Source:         proto.Uint32(uint32(r.localID)),
	}
	// Make entry tracking acks outstanding
	ll, ok := r.liveLinks[p.Preheader.Receiver]
	if ok {
		switch {
		case ll.sentBeforeSetTimeout < 3:
			ll.sentBeforeSetTimeout += 1

		case ll.sentBeforeSetTimeout == 3:
			ll.sentBeforeSetTimeout += 1
			// set timeout
			t := time.Now().Add(ERROR_REPORTING_TIMEOUT)
			ll.timeout = &t

		case ll.sentBeforeSetTimeout > 3:
			if ll.timeout != nil {
				if time.Now().After(*ll.timeout) {
					// TODO send route error
					r.routeCache.removeNeighbor(NodeID(p.Preheader.Receiver))
					if p.Header.Destination != nil {
						r.originatePacket(newErrorPacket(uint32(r.localID), *p.Header.Destination, &header.DSRHeader_NodeUnreachableError{
							Salvage:                proto.Uint32(uint32(0)),
							Source:                 proto.Uint32(uint32(r.localID)),
							Destination:            p.Header.Destination,
							UnreachableNodeAddress: proto.Uint32(uint32(p.Preheader.Receiver)),
						}))
					}
				}
			}
		}
	} else {
		nl := linkMaint{
			sentBeforeSetTimeout: 1,
			timeout:              nil,
		}

		r.liveLinks[p.Preheader.Receiver] = nl
	}
}

// TODO check if header exists
// pg62
// Returns true if we should forward, false if we do not forward
func processSourceRoute(p *packet.Packet) bool {
	// TODO more stuff
	// decrement segments left
	sr := p.Header.DsrHeader.SourceRoute
	if sr == nil {
		return false
		// If no source route, quietly drop
	}
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

// processRouteReply adds the route to our cache and searches for packets to
// send with new route.
func (r *router) processRouteReply(p *packet.Packet) {
	rr := p.Header.DsrHeader.RouteReply
	if rr == nil || *p.Header.Destination != uint32(r.localID) {
		return
	}

	// convert TODO remove when removing NodeID type
	nroute := make([]NodeID, 0, len(rr.Addresses)+1)
	for _, a := range rr.Addresses {
		nroute = append(nroute, NodeID(a))
	}
	nroute = append(nroute, NodeID(*p.Header.Source))
	r.routeCache.addRoute(nroute, 0)

	// Check send buffer
	sendable := r.sendBuffer.getSendable(nroute)
	for _, op := range sendable {

		// Output packet
		r.sendAlongSourceRoute(op)
	}

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

func (r *router) processAck(p *packet.Packet) {
	ack := p.Header.DsrHeader.Ack
	if ack == nil {
		return
	}

	// Drop request
	p.Header.DsrHeader.Ack = nil
	ll, ok := r.liveLinks[*ack.Source]
	if ok {
		ll.sentBeforeSetTimeout = 0
		ll.timeout = nil
	}
}
func (r *router) processRouteError(p *packet.Packet) {
	re := p.Header.DsrHeader.NodeUnreachableError
	if re == nil {
		return
	}

	// Drop request
	if *re.Destination == uint32(r.localID) {
		p.Header.DsrHeader.NodeUnreachableError = nil
	}

	r.routeCache.removeLink(*re.Source, *re.UnreachableNodeAddress)
}

func (r *router) processAckRequest(p *packet.Packet) {
	ar := p.Header.DsrHeader.AckRequest
	if ar == nil {
		return
	}

	// Drop request
	p.Header.DsrHeader.AckRequest = nil

	// Must be next hop in addresses or destination of packet
	// TODO don't check receiver, listen to RFC
	// TODO should we drop ACK request
	if p.Preheader.Receiver != uint32(r.localID) {
		return
	}
	if p.Header.DsrHeader.Ack != nil {
		return
	}

	r.out <- *newAckPacket(r.localID, NodeID(*ar.Source), *ar.Identification)

}
