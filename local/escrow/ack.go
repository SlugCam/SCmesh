package escrow

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/util"
)

type ACK struct {
	FileID int64
	Offset int64
	Size   int
}

func (ack ACK) send(dest uint32, r pipeline.Router) {
	dh := header.DataHeader{
		Type:         header.DataHeader_ACK.Enum(),
		Destinations: []uint32{dest},
	}
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	enc.Encode(ack)
	r.OriginateDSR(dest, int64(0), dh, util.Encode(b.Bytes()))
}

func parseACK(p packet.Packet) (ack ACK, err error) {
	if p.Header == nil || p.Header.DataHeader == nil || *p.Header.DataHeader.Type != header.DataHeader_ACK {
		err = fmt.Errorf("parseACK: packet is not an ACK packet")
		return
	}
	var a85decoded []byte
	a85decoded, err = util.Decode(p.Payload)
	if err != nil {
		return
	}
	b := bytes.NewBuffer(a85decoded)
	dec := gob.NewDecoder(b)
	err = dec.Decode(&ack)
	return
}
