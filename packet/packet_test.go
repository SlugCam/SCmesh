package packet

import (
	"fmt"
	"testing"

	"github.com/SlugCam/SCmesh/packet/crypto"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/golang/protobuf/proto"
)

func TestPackingPacket(t *testing.T) {
	p := new(Packet)
	p.Header = new(header.Header)
	ch := make(chan []byte, 100)
	enc := crypto.NewEncrypter("test")
	p.Header.Type = proto.Int32(0)

	p.Pack(enc, ch)
	c := <-ch
	fmt.Println(c)

	fmt.Printf("%+v\n", p)

}
