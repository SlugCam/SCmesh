package dsr

import (
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/golang/protobuf/proto"
)

func newDSRPacket() *packet.Packet {
	p := packet.NewPacket()
	p.Header.DsrHeader = new(header.DSRHeader)
	return p
}

func newRouteRequest() *packet.Packet {
	p := newDSRPacket()
	p.Header.DsrHeader.RouteRequest = new(header.DSRHeader_RouteRequest)
	return p
}

func newOriginationPacket(o *OriginationRequest) *packet.Packet {
	p := newDSRPacket()
	p.Header.DataHeader = &o.DataHeader
	p.Payload = o.Data
	return p
}

// addSourceRoute adds the source route option to the given DSR packet. If the
// packet already has a source route this function will silently replace it. If
// the packet does not have a header or DSR header we add them. This process is
// outlines in RFC4728 sec. 8.1.3.
func addSourceRoute(p *packet.Packet, route []NodeID) error {
	if p.Header == nil {
		p.Header = new(header.Header)
	}
	if p.Header.DsrHeader == nil {
		p.Header.DsrHeader = new(header.DSRHeader)
	}

	// Convert the route
	addresses := make([]uint32, 0, len(route))
	for _, v := range route {
		addresses = append(addresses, uint32(v))
	}

	// Add source route option
	p.Header.DsrHeader.SourceRoute = &header.DSRHeader_SourceRoute{
		Salvage:   proto.Uint32(0),
		SegsLeft:  proto.Uint32(uint32(len(route))),
		Addresses: addresses,
	}

	return nil
}
