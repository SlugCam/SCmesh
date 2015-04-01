package local

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/SlugCam/SCmesh/local/distribute"
	"github.com/SlugCam/SCmesh/packet/header"
)

type MockRouter struct {
	id uint32
}

func NewMockRouter(id uint32) *MockRouter {
	r := &MockRouter{
		id: id,
	}
	return r
}

func (r *MockRouter) LocalID() uint32 {
	return r.id
}
func (r *MockRouter) OriginateDSR(dest, offset uint32, dataHeader header.DataHeader, data []byte) {

}

func (r *MockRouter) OriginateFlooding(TTL int, dataHeader header.DataHeader, data []byte) {

}

func TestEscrow(t *testing.T) {
	dir, err := ioutil.TempDir("", "SCmesh_test")
	if err != nil {
		t.Fatal("error making tmp directory to run test in.")
	}

	r := NewMockRouter(0)

	acks := make(chan distribute.ACK)

	d, err := distribute.Distribute(dir, r, acks)
	if err != nil {
		t.Fatal("Error in Distribute:", err)
	}

	message := json.RawMessage(`{"test":45}`)
	_, err = d.Register(distribute.RegistrationRequest{
		DataType:    "message",
		Destination: uint32(0),
		Timestamp:   time.Now(),
		JSON:        &message,
	})
	if err != nil {
		t.Fatal("Error in Register:", err)
	}
	time.Sleep(2 * time.Second)
}
