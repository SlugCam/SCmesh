package escrow

import (
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/routing"
	"github.com/golang/protobuf/proto"
)

type ACK struct {
	FileNumber uint32
	Offset     uint32
	Size       uint32
}

func sendAck(ack ACK, dest uint32, r pipeline.Router) {
	dh := header.DataHeader{
		Type:         header.DataHeader_ACK.Enum(),
		Destinations: []uint32{dest},
	}
	r.OriginateDSR(dest, dh, []byte{})
}
