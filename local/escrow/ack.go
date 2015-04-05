package escrow

import (
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
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
	// TODO make ACK
	r.OriginateDSR(dest, uint32(0), dh, []byte{})
}
