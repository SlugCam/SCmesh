// The packet package provides data structures and routines for the handling of
// packet objects. Along with the data structure of packets it also defines two
// pipeline functions for handling them: ParsePackets and PackPackets.
package packet

import (
	"encoding/binary"
	"errors"

	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/golang/protobuf/proto"
)

const (
	MAX_PACKET_LEN            = 1460
	SERIALIZED_PREHEADER_SIZE = 8
)

type RawPacket struct {
	Preheader []byte // Ascii85 encoded plaintext
	Header    []byte // Ascii85 encoded serialized header object
	Payload   []byte // Encoded encrypted payload
}

type Preheader struct {
	Receiver      uint32
	PayloadOffset uint32
}

type Packet struct {
	Preheader Preheader
	Header    *header.Header
	Payload   []byte // Encoded encrypted payload
}

func NewPacket() *Packet {
	p := new(Packet)
	p.Header = new(header.Header)
	return p
}

// serializePreheader provides serialization of the packet preheader.
func (p *Preheader) Serialize() []byte {
	out := make([]byte, 0, 8)
	receiver := make([]byte, 4)
	offset := make([]byte, 4)
	binary.LittleEndian.PutUint32(receiver, p.Receiver)
	binary.LittleEndian.PutUint32(offset, p.PayloadOffset)
	out = append(out, receiver...)
	out = append(out, offset...)
	return out
}

// Pack takes a single packet and encodes it to the wire format. It will
// fragment the data if necessary.
func (p *Packet) Pack(out chan<- []byte) {

	originalOffset := int(p.Preheader.PayloadOffset)
	payloadLen := len(p.Payload)
	relativeOffset := 0
	var nextOffset int

	serializedHeader, err := proto.Marshal(p.Header) // Will remain the same for all packets
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	encodedHeader := encode(serializedHeader)

	headerSize := len(encodedHeader)
	maxPreheaderSize := maxEncodedLen(SERIALIZED_PREHEADER_SIZE)
	delimiterSize := 4
	maxPayloadSize := MAX_PACKET_LEN - delimiterSize - headerSize - maxPreheaderSize

	for {
		b := make([]byte, 0, MAX_PACKET_LEN)

		// Preheader
		newPreheader := p.Preheader // Will make a copy of the preheader
		newPreheader.PayloadOffset = uint32(originalOffset + relativeOffset)
		serializedPreheader := newPreheader.Serialize()
		encodedPreheader := encode(serializedPreheader)

		// Fit Payload
		remainingPayloadLen := payloadLen - relativeOffset
		if maxPayloadSize < remainingPayloadLen {
			// We have less room in the packet than we need
			nextOffset = relativeOffset + maxPayloadSize
		} else {
			// We can fit the rest of the packet
			nextOffset = payloadLen
		}
		payloadSlice := p.Payload[relativeOffset:nextOffset]
		relativeOffset = nextOffset

		// Build packet
		b = append(b, '\x01') // Packet delimiter
		b = append(b, encodedPreheader...)
		b = append(b, '\x00') // Section delimiter
		b = append(b, encodedHeader...)
		b = append(b, '\x00') // Section delimiter
		b = append(b, payloadSlice...)
		b = append(b, '\x04') // Section delimiter

		out <- b
		log.Debug("Sending this packet: %+v", struct {
			Preheader Preheader
			Header    *header.Header
			Data      []byte
		}{newPreheader, p.Header, payloadSlice})

		if nextOffset == payloadLen {
			break
		}
	}
}

func (raw *RawPacket) Parse() (pack Packet, err error) {
	// Copy reference to payload
	pack.Payload = raw.Payload
	log.Debug("Parsing this raw packet:", raw)

	// Decode preheader
	decodedPreheader, err := decode(raw.Preheader)
	if err != nil {
		return
	}
	// Parse preheader
	log.Debug("Decoded preheader is:", decodedPreheader)
	if len(decodedPreheader) != SERIALIZED_PREHEADER_SIZE {
		err = errors.New("incorrect preheader length")
		return
	}
	pack.Preheader.Receiver = binary.LittleEndian.Uint32(decodedPreheader[0:4])
	pack.Preheader.PayloadOffset = binary.LittleEndian.Uint32(decodedPreheader[4:8])

	// TODO If receiver is incorrect we can drop (or continue if peeking is desired)

	serializedHeader, err := decode(raw.Header)
	if err != nil {
		return
	}

	// Parse header with protobuffer
	pack.Header = &header.Header{}
	err = proto.Unmarshal(serializedHeader, pack.Header)

	return
}

// ParsePackets is intended to be used in the main SCmesh pipeline to parse raw
// packets provided from the in channel and push them to the out channel.
func ParsePackets(in <-chan RawPacket, out chan<- Packet) {
	go func() {
		for c := range in {
			p, err := c.Parse()
			if err != nil {
				log.Error("Packet dropped during parsing.", err)
			} else {
				out <- p
				log.Debug("Parsed:", p)
			}
		}
	}()
}

// PackPackets is intended to be used in the main SCmesh pipeline to pack
// packets provided from the in channel and push them to the out channel.
func PackPackets(in <-chan Packet, out chan<- []byte) {
	go func() {
		for p := range in {
			p.Pack(out)
		}
	}()
}
