package escrow

import (
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
)

type ACK struct {
	FileID uint64
	Offset uint64
	Size   uint64
}

func (ack ACK) send(dest uint32, r pipeline.Router) {
	dh := header.DataHeader{
		Type:         header.DataHeader_ACK.Enum(),
		Destinations: []uint32{dest},
	}
	// TODO make ACK
	r.OriginateDSR(dest, uint32(0), dh, []byte{})
}
