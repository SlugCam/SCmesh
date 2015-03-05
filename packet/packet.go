package packet

import (
	"errors"
	"log"

	"github.com/lelandmiller/SCcomm/constants"
)

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
