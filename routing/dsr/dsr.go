package dsr

import (
	"time"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

// TODO should not be a global
var Cost int

type NodeID uint32
type Route []NodeID

const RR_RESEND_TIMEOUT = 1 * time.Second

const BROADCAST_ID = 0xFFFFFFFF
const LINK_RESEND_TIMEOUT = 2 * time.Second

// MAX_SENDS=0 will not resend and PACKETS_DROPPED will become the
// acutal packets dropped before a route error is sent.
const MAX_SENDS = 0
const LINK_RESEND_JITTER = 10 * time.Millisecond

const ERROR_REPORTING_TIMEOUT = 10 * time.Second
const PACKETS_DROPPED_BEFORE_LINK_ERROR = 30

// OriginationRequest is a struct used to describe a new packet we should
// originate. They are used in RoutePackets to provide communication to this
// module from outside.
type OriginationRequest struct {
	Destination uint32
	Offset      int64
	DataHeader  header.DataHeader
	Data        []byte
}

// RoutePackets is the main pipeline function that creates a DSR router and
// manages packet origination and forwarding.
func RoutePackets(localID uint32, toForward <-chan packet.Packet, toOriginate <-chan OriginationRequest, out chan<- packet.Packet) {

	r := newRouter(localID, out)

	go func() {
		for {
			select {
			case p := <-toForward:
				r.forward(p)
			case o := <-toOriginate:
				r.originate(o)
			case ackIDToCheck := <-r.resendTimeout:
				r.resendPacket(ackIDToCheck)
			case target := <-r.routeRequestTimeout:
				r.processRouteRequestTimeout(target)
			}

		}

	}()

}
