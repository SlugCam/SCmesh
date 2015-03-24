package dsr

// TODO set required fields for packet including RECEIVER

import (
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
)

// newDSRPacket creates a new packet.Packet with a DSRHeader attached.
func newDSRPacket() *packet.Packet {
	p := packet.NewPacket()
	p.Header.DsrHeader = new(header.DSRHeader)
	return p
}

func newRouteRequest(source NodeID, dest NodeID) *packet.Packet {
	p := newDSRPacket()
	p.Header.Source = proto.Uint32(uint32(source))
	p.Preheader.Receiver = uint32(BROADCAST_ID)

	rr := new(header.DSRHeader_RouteRequest)
	p.Header.DsrHeader.RouteRequest = rr
	// TODO Set up route request
	rr.Id = proto.Uint32(util.RandomUint32()) // TODO not random
	rr.Target = proto.Uint32(uint32(dest))
	//Addresses
	return p
}

func newOriginationPacket(o OriginationRequest) *packet.Packet {
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
	// Ensure that the proper headers exist
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

// TODO rr id included? not shown in packet specs
func newRouteReply(addresses []uint32, orig uint32, target uint32) *packet.Packet {
	p := newDSRPacket()

	reply := new(header.DSRHeader_RouteReply)
	p.Header.DsrHeader.RouteReply = reply

	reply.Addresses = make([]uint32, len(addresses))
	copy(reply.Addresses, addresses)

	// find return route
	returnRoute := make([]NodeID, len(addresses))
	for i, a := range addresses {
		returnRoute[len(addresses)-1-i] = NodeID(a)
	}

	addSourceRoute(p, returnRoute)

	p.Header.Destination = proto.Uint32(orig)
	p.Header.Source = proto.Uint32(target)

	return p
}
