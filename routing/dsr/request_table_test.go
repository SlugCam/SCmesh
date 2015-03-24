package dsr

import "testing"

func TestIncomingRequestCache(t *testing.T) {
	rt := newRequestTable()

	toAdd := []struct {
		initiator NodeID
		target    NodeID
		id        uint32
	}{
		{NodeID(3), NodeID(8), uint32(1)},
		{NodeID(3), NodeID(8), uint32(4)},
		{NodeID(5), NodeID(0), uint32(4)},
		{NodeID(5), NodeID(1), uint32(4)},
		{NodeID(7), NodeID(3), uint32(1)},
		{NodeID(8), NodeID(3), uint32(1)},
	}
	for _, c := range toAdd {
		if rt.hasReceivedRequest(c.initiator, c.target, c.id) {
			t.Fatal("received route request returned that it was seen in an initialized cache.")
		}
		rt.receivedRequest(c.initiator, c.target, c.id)
		if !rt.hasReceivedRequest(c.initiator, c.target, c.id) {
			t.Fatal("received route request not marked as seen after adding it to cache.")
		}

	}
}
