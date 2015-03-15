package flooding

import (
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

type OriginationRequest struct {
	TTL        int
	DataHeader header.DataHeader
	Data       []byte
}

func RoutePackets(toForward <-chan packet.Packet, toOriginate <-chan OriginationRequest, out chan<- packet.Packet) {
	go func() {
		for {
			select {
			case c := <-toForward:

			case origReq := <-toOriginate:
				// Make new packet
				p := packet.NewPacket()
				// Load data
				p.Header.DataHeader = origReq.DataHeader
				p.Payload = origReq.Data
				// Make flooding header
				p.Header.Ttl = uint32(origReq.TTL)
				p.Header.FloodingHeader = new(header.FloodingHeader)
				// Assign random
				// p.Header.FloodingHeader.PacketId = proto.Uint32()
				// Send to output
			}

		}

	}()

}
