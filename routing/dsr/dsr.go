package dsr

import (
	"time"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

type NodeID uint32
type Route []NodeID

const RR_RESEND_TIMEOUT = 1 * time.Second

const BROADCAST_ID = 0xFFFFFFFF
const LINK_RESEND_TIMEOUT = 1 * time.Second
const LINK_RESEND_JITTER = 1 * time.Second

const ERROR_REPORTING_TIMEOUT = 10 * time.Second
const ACK_REQUEST_BEFORE_TIMEOUT = 25

type linkMaint struct {
	sentBeforeSetTimeout int
	timeout              *time.Time
}

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
