package routing

import "github.com/SlugCam/SCmesh/packet"

// RoutePackets is the main pipeline function for the routing package. It is
// responsible for performing all routing functions on packets and outputting
// the resulting packets.
func RoutePackets(in <-chan packet.Packet, destLocal chan<- packet.Packet, out chan<- packet.Packet) {
	/*
		go func() {
			for c := range in {

			}
		}()
	*/
}
