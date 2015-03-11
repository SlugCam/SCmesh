// The packet package provides data structures and routines for the handling of
// packet objects. Along with the data structure of packets it also defines two
// pipeline functions for handling them: ParsePackets and PackPackets.
package packet

import (
	"encoding/binary"
	"errors"

	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/packet/crypto"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/golang/protobuf/proto"
)

const (
	MAX_PACKET_LEN            = 1460
	SERIALIZED_PREHEADER_SIZE = 6
	NONCE_LENGTH              = 12
)

type RawPacket struct {
	Preheader []byte // Ascii85 Encoded Plaintext
	Header    []byte // Ascii85 Encoded Encrypted Serialized Header Object
	Payload   []byte // Encoded encrypted payload
}

type Preheader struct {
	Receiver      uint16
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
}

// serializePreheader provides serialization of the packet preheader.
func (p *Preheader) Serialize() []byte {
	out := make([]byte, 0, 6)
	receiver := make([]byte, 2)
	offset := make([]byte, 4)
	binary.LittleEndian.PutUint16(receiver, p.Receiver)
	binary.LittleEndian.PutUint32(offset, p.PayloadOffset)
	out = append(out, receiver...)
	out = append(out, offset...)
	return out
}

// Pack takes a single packet and encodes it to the wire format. It will
// fragment the data if necessary.
func (p *Packet) Pack(encrypter *crypto.Encrypter, out chan<- []byte) {

	originalOffset := int(p.Preheader.PayloadOffset)
	payloadLen := len(p.Payload)
	relativeOffset := 0
	var nextOffset int

	serializedHeader, err := proto.Marshal(p.Header) // Will remain the same for all packets
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	maxHeaderSize := encrypter.MaxEncryptedLen(len(serializedHeader))
	maxPreheaderSize := encrypter.MaxEncodedLen(encrypter.NonceSize() + SERIALIZED_PREHEADER_SIZE)
	delimiterSize := 4
	maxPayloadSize := MAX_PACKET_LEN - delimiterSize - maxHeaderSize - maxPreheaderSize

	for {
		b := make([]byte, 0, MAX_PACKET_LEN)

		// Preheader
		newPreheader := p.Preheader // Will make a copy of the preheader
		newPreheader.PayloadOffset = uint32(originalOffset + relativeOffset)
		serializedPreheader := newPreheader.Serialize()

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

		// Encrypt and encode headers
		encryptedHeader, nonce := encrypter.HeaderToWireFormat(serializedPreheader, serializedHeader, payloadSlice)
		noncePreheader := append(nonce, serializedPreheader...)
		encodedPreheader := crypto.Encode(noncePreheader)

		// Build packet
		b = append(b, '\x01') // Packet delimiter
		b = append(b, encodedPreheader...)
		b = append(b, '\x00') // Section delimiter
		b = append(b, encryptedHeader...)
		b = append(b, '\x00') // Section delimiter
		b = append(b, payloadSlice...)
		b = append(b, '\x04') // Section delimiter

		out <- b

		if nextOffset == payloadLen {
			break
		}
	}
}

func (raw *RawPacket) Parse(crypter *crypto.Encrypter) (pack Packet, err error) {
	// Copy reference to payload
	pack.Payload = raw.Payload

	// Decode preheader
	decodedPreheader, err := crypto.Decode(raw.Preheader)
	if err != nil {
		return
	}
	// Parse preheader
	if len(decodedPreheader) != SERIALIZED_PREHEADER_SIZE+NONCE_LENGTH {
		err = errors.New("Incorrect preheader length")
		return
	}
	nonce := decodedPreheader[0:12]
	pack.Preheader.Receiver = binary.LittleEndian.Uint16(decodedPreheader[12:14])
	pack.Preheader.PayloadOffset = binary.LittleEndian.Uint32(decodedPreheader[14:18])

	// TODO If receiver is incorrect we can drop (or continue if peeking is desired)

	// Unseal header with preheader 0x00 payload as authenticated data
	serializedHeader, err := crypter.HeaderFromWireFormat(nonce, decodedPreheader[12:18], raw.Header, raw.Payload)
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
	encrypter := crypto.NewEncrypter("/slugcam/key") // Should only be used by one goroutine at a time
	go func() {
		for c := range in {
			p, err := c.Parse(encrypter)
			if err != nil {
				log.Error("Packet dropped during parsing.", err)
			} else {
				out <- p
			}
		}
	}()
}

// PackPackets is intended to be used in the main SCmesh pipeline to pack
// packets provided from the in channel and push them to the out channel.
func PackPackets(in <-chan Packet, out chan<- []byte) {
	encrypter := crypto.NewEncrypter("/slugcam/key") // Should only be used by one goroutine at a time
	go func() {
		for p := range in {
			p.Pack(encrypter, out)
		}
	}()
}
