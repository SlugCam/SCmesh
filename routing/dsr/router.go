package dsr

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
)

type resendMessage struct {
	id    uint32
	count int
}

// These data structures could be optimized
type router struct {
	sentPackets         map[uint32]*packet.Packet // map from ack id to packet
	resendTimeout       chan resendMessage
	routeRequestTimeout chan uint32
	liveLinks           map[uint32]linkMaint
	localID             uint32
	routeCache          *routeCache
	sendBuffer          *sendBuffer
	requestTable        *requestTable
	out                 chan<- packet.Packet
}

func newRouter(localID uint32, out chan<- packet.Packet) *router {
	r := new(router)
	r.sentPackets = make(map[uint32]*packet.Packet)
	r.resendTimeout = make(chan resendMessage)
	r.routeRequestTimeout = make(chan uint32)
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
		//r.out <- *packet
		r.sendAlongSourceRoute(packet, 0)
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
	route := r.routeCache.getRoute(dest)
	if route == nil {
		// If no route found perform route discovery
		r.sendBuffer.addPacket(p)
		// TODO initiateRouteDiscovery

		r.requestDiscovery(dest)
	} else {
		// Otherwise add source route option to packet
		err := addSourceRoute(p, route)
		if err != nil {
			log.Error("DSR originate:", err)
		}
		// Output packet
		//r.out <- *p
		r.sendAlongSourceRoute(p, 0)
	}

}

// DISCOVERY FUNCTIONS

func (r *router) sendRouteRequest(target uint32) {
	log.WithFields(log.Fields{
		"sentTable":  fmt.Sprintf("%v", r.requestTable.sentRequests),
		"routeCache": r.routeCache.dump(),
	}).Infof("sendRR for target %d", target)

	r.requestTable.sentRequest(target)
	r.out <- *newRouteRequest(r.localID, target)
	r.out <- *newRouteRequest(r.localID, target)
	r.out <- *newRouteRequest(r.localID, target)
	// TODO set timeout
	time.AfterFunc(RR_RESEND_TIMEOUT, func() {
		r.routeRequestTimeout <- target
	})
}
func (r *router) requestDiscovery(target uint32) {
	if !r.requestTable.discoveryInProcess(target) {
		log.Info("reqdisc")
		r.sendRouteRequest(target)
	}
}
func (r *router) processRouteRequestTimeout(target uint32) {
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

	// TODO is this right place
	r.processAckRequest(&p)
	r.processAck(&p)

	cont := r.processRouteRequest(&p)
	if !cont {
		return
	}
	// TODO process route reply
	r.processRouteReply(&p)
	r.processRouteError(&p)

	// Only forward if we were the receiver in the first place, broadcast not
	// valid for source route anyway.
	// TODO this should be anywhere where we respond on header options
	if p.Preheader.Receiver == uint32(r.localID) {
		r.sendAlongSourceRoute(&p, 0)
	}
}

func (r *router) sendAlongSourceRoute(p *packet.Packet, resends int) {
	if processSourceRoute(p) {
		r.addAckRequest(p, resends)
		r.out <- *p
	}
}

func (r *router) addAckRequest(p *packet.Packet, resends int) {
	if p.Preheader.Receiver == BROADCAST_ID {
		return
	}
	id := util.RandomUint32()
	p.Header.DsrHeader.AckRequest = &header.DSRHeader_AckRequest{
		Identification: proto.Uint32(id),
		Source:         proto.Uint32(uint32(r.localID)),
	}

	// Detect broken links
	// Make entry tracking acks outstanding
	ll, ok := r.liveLinks[p.Preheader.Receiver]
	if ok {
		switch {
		case ll.sentBeforeSetTimeout < ACK_REQUEST_BEFORE_TIMEOUT:
			ll.sentBeforeSetTimeout += 1

		case ll.sentBeforeSetTimeout == ACK_REQUEST_BEFORE_TIMEOUT:
			ll.sentBeforeSetTimeout += 1
			// set timeout
			t := time.Now().Add(ERROR_REPORTING_TIMEOUT)
			ll.timeout = &t

		case ll.sentBeforeSetTimeout > ACK_REQUEST_BEFORE_TIMEOUT:
			if ll.timeout != nil {
				if time.Now().After(*ll.timeout) {
					// TODO send route error
					r.routeCache.removeNeighbor(p.Preheader.Receiver)
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

	// save packet for resending
	if resends < 4 {
		r.sentPackets[id] = p
		delay := LINK_RESEND_TIMEOUT + time.Duration(rand.NormFloat64()*float64(LINK_RESEND_JITTER))
		if int64(delay) < 0 {
			delay = time.Duration(0)
		}
		time.AfterFunc(delay, func() {
			r.resendTimeout <- resendMessage{id, resends + 1}
		})
	} else {
		log.Info("Packet hit max resends")
	}
}

func (r *router) resendPacket(m resendMessage) {
	p, ok := r.sentPackets[m.id]
	if ok {
		// packet is still cached
		delete(r.sentPackets, m.id)
		p.Header.DsrHeader.AckRequest = nil
		r.sendAlongSourceRoute(p, m.count)
	}
}

// TODO check if header exists
// pg62
// Returns true if we should forward, false if we do not forward
func processSourceRoute(p *packet.Packet) bool {
	// TODO more stuff
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
	r.requestTable.receivedReply(*p.Header.Source)
	r.routeCache.addRoute(rr.Addresses, *p.Header.Source)
	// Check send buffer
	nroute := make([]uint32, 0, len(rr.Addresses)+1)
	for _, n := range rr.Addresses {
		nroute = append(nroute, *n.Address)
	}
	nroute = append(nroute, *p.Header.Source)

	sendable := r.sendBuffer.getSendable(nroute)
	for _, op := range sendable {
		// Output packet
		r.sendAlongSourceRoute(op, 0)
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
		r.sendAlongSourceRoute(reply, 0)
		return true // Is this right?
	}

	//  In spec this is lower
	// Check for route request entry to see if we have seen this route request
	if r.requestTable.hasReceivedRequest(*p.Header.Source, *rr.Target, *rr.Id) {
		return false // route request already seen
	}

	// Check if addresses contains our ip, if so drop packet immediately
	if *p.Header.Source == uint32(r.localID) {
		return false
	}
	for _, a := range rr.Addresses {
		if *a.Address == uint32(r.localID) {
			log.Error("loop found in route request")
			return false
		}
	}

	// At this point continue processing

	// Add request to cache
	r.requestTable.receivedRequest(*p.Header.Source, *rr.Target, *rr.Id)
	// Make copy of route request option
	nr := *rr
	p.Header.DsrHeader.RouteRequest = &nr
	nr.Addresses = append(nr.Addresses, &header.DSRHeader_Node{
		Address: proto.Uint32(uint32(r.localID)),
		Cost:    proto.Uint32(localCost()),
	})

	// TODO Check if we can perform a cached route reply

	// TODO Forward RR
	// TODO BROADCAST JITTER
	r.out <- *p
	return false

}

// TODO
func localCost() uint32 {
	return uint32(0)
}

func (r *router) processAck(p *packet.Packet) {
	if p.Preheader.Receiver != r.localID {
		return
	}
	ack := p.Header.DsrHeader.Ack
	if ack == nil {
		return
	}

	// Remove packet from cache
	id := *ack.Identification
	delete(r.sentPackets, id)
	log.Infof("removed sending from cache: %d rem", len(r.sentPackets))

	// Update live link
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
	if *re.Destination == r.localID {
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

	r.out <- *newAckPacket(r.localID, *ar.Source, *ar.Identification)

}
