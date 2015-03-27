package escrow

import (
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
)

type ACK struct {
	FileNumber uint32
	Offset     uint32
	Size       uint32
	Type       int
}

func sendAck(ack ACK, dest uint32, r pipeline.Router) {
	dh := header.DataHeader{
		Type:         header.DataHeader_ACK.Enum(),
		Destinations: []uint32{dest},
	}
	r.OriginateDSR(dest, dh, []byte{})
}
