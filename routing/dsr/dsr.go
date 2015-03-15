package dsr

import (
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

type id uint32

type OriginationRequest struct {
	TTL        int
	DataHeader header.DataHeader
	Data       []byte
}

type cacheEntry struct {
	route []id
	cost  int
}

// These data structures could be optimized
type DsrRouter struct {
	routeCache []cacheEntry
	sendBuffer []packet.Packet
}

func (r *DsrRouter) updateRouteCache() {
	// Update cache
	// Check Send buffer if any packets destination field is now reachable
	// If so send packet and remove from buffer

}

// getCachedRoute looks into the route cache and returns a route
func (r *DsrRouter) getCachedRoute() {

}

// Originate starts the process of sending a packet. It should not be called
func (r *DsrRouter) originate(p *packet.Packet) {

}

func sendRouteRequest() {
	p := packet.NewPacket()
	p.Header.DsrHeader = new(header.DSRHeader)
	p.Header.DsrHeader.RouteRequest = new(header.DSRHeader_RouteRequest)
}

func RoutePackets(localID uint32, toForward <-chan packet.Packet, toOriginate <-chan OriginationRequest, out chan<- packet.Packet) {
	//r := new(DsrRouter)
	go func() {
		for {
			select {
			case c := <-toForward:
				_ = c
			case c := <-toOriginate:
				_ = c
			}

		}

	}()

}
