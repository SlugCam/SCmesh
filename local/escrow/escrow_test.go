package escrow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
)

type MockRouter struct {
	id  uint32
	out chan packet.Packet
}

func NewMockRouter(id uint32) *MockRouter {
	r := &MockRouter{
		id:  id,
		out: make(chan packet.Packet, 1000),
	}
	return r
}
func (r *MockRouter) LocalID() uint32 {
	return r.id
}
func (r *MockRouter) OriginateDSR(dest uint32, offset int64, dataHeader header.DataHeader, data []byte) {
	// Make new packet
	p := packet.NewPacket()
	//p.Header.DsrHeader = new(header.DSRHeader)
	p.Header.Source = proto.Uint32(r.id)
	p.Header.DataHeader = &dataHeader
	p.Payload = data
	p.Preheader.PayloadOffset = offset
	// Send packet to out channel
	fmt.Printf("Pre:\n%v\nHeader:\n%v\nData:\n%s\n", p.Preheader, p.Header, p.Payload)
	r.out <- *p
}

func (r *MockRouter) OriginateFlooding(TTL int, dataHeader header.DataHeader, data []byte) {

}

func TestEscrow(t *testing.T) {
	dir, err := ioutil.TempDir("", "SCmesh_test")
	if err != nil {
		t.Fatal("error making tmp directory to run test in.")
	}

	r1 := NewMockRouter(1)
	r2 := NewMockRouter(2)

	out := make(chan CollectedData)

	// Distribute
	d, err := Distribute(dir, r2.out, r1)
	if err != nil {
		t.Fatal("Error in Distribute:", err)
	}

	// Collect
	_, err = Collect(dir, r1.out, out, r2)
	if err != nil {
		t.Fatal("Error in Distribute:", err)
	}

	data := util.RandomSlice(1040)
	testStruct := &struct {
		Trial string
		Data  []byte
	}{"Test", data}

	//message := json.RawMessage(`{"test":45}`)
	message, err := json.Marshal(&testStruct)
	m := json.RawMessage(message)
	_, err = d.Register(RegistrationRequest{
		DataType:    "message",
		Destination: uint32(0),
		Timestamp:   time.Now(),
		JSON:        &m,
	})

	if err != nil {
		t.Fatal("Error in Register:", err)
	}
	time.Sleep(7 * time.Second)
}
