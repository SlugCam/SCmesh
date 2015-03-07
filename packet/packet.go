// The packet package provides data structures and routines for the handling of
// packet objects. Along with the data structure of packets it also defines two
// pipeline functions for handling them: ParsePackets and PackPackets.
package packet

import (
	"errors"
	"log"

	"github.com/SlugCam/SCmesh/constants"
)

const RAW_PACKET_SIZE = 1460
const DELIMITER = "CBU\r\n"

type Packet struct {
	Payload []byte // Should contain the encrypted data for the packet
}

func (p *Packet) WireFormat() []byte {
	b := make([]byte, 0, 1460)
	b = append(b, constants.PACK_SEQ...)
	b = append(b, p.Payload...)
	remainingLen := constants.RAW_PACKET_SIZE - len(b)
	if remainingLen >= 0 {
		b = append(b, make([]byte, remainingLen)...)
	} else {
		log.Panic(errors.New("Payload too large"))
	}
	return b
}

// ParsePackets is intended to be used in the main SCmesh pipeline to parse raw
// packets provided from the in channel and push them to the out channel.
func ParsePackets(in <-chan []byte, out chan<- packet.Packet) {
	go func() {
		for c := range in {
			out <- packet.ParsePacket(c)
		}
	}()
}

// PackPackets is intended to be used in the main SCmesh pipeline to pack
// packets provided from the in channel and push them to the out channel.
func PackPackets(in <-chan Packet, out chan<- []byte) {
	go func() {
		for p := range routingOutCh {
			packedPackets <- p.ToWireFormat()
		}
	}()
}
