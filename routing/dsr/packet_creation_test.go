package dsr

import "testing"

// These requirements are outlined in section 8.2.1 of the RFC.
func TestNewRouteRequest(t *testing.T) {
	sources := []NodeID{NodeID(3), NodeID(43)}
	destinations := []NodeID{NodeID(31), NodeID(56), NodeID(1)}

	ids := make([]uint32, 0, len(destinations))

	for _, src := range sources {
		for _, dest := range destinations {
			p := newRouteRequest(src, dest)
			if p.Header == nil {
				t.Fatal("newRouteRequest: new packet lacks header.")
			}
			if p.Header.DsrHeader == nil {
				t.Fatal("newRouteRequest: new packet lacks DSR header.")
			}
			if p.Header.DsrHeader.RouteRequest == nil {
				t.Fatal("newRouteRequest: new packet lacks DSR route request option.")
			}
			if *p.Header.DsrHeader.RouteRequest.Target != uint32(dest) {
				t.Fatal("newRouteRequest: new packet has incorrect target")

			}
			if *p.Header.Source != uint32(src) {
				t.Fatal("newRouteRequest: new packet has incorrect source")
			}

			if p.Preheader.Receiver != uint32(BROADCAST_ID) {
				t.Fatal("newRouteRequest: route request receiver must be set to the broadcast id.")
			}

			// Check ids for uniqueness
			id := *p.Header.DsrHeader.RouteRequest.Id
			for _, v := range ids {
				if id == v {
					t.Fatal("newRouteRequest: route request id must be unique for each route request produced.")
				}
			}
			ids = append(ids, id)
		}
	}
}
