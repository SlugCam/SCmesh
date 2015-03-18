package simulation

// Node is a simulated SlugCam mesh node. This is used in integration testing
// and simulation. This struct should be created using the NewNode function.
type Node struct {
	Router          pipeline.Router
	IncomingPackets <-chan packet.Packet
	LocalPackets    <-chan packet.Packet
	mockWiFly       MockWiFly
}

// NewNode creates a new simulated node. It does this by creating a default
// pipeline configuration and then intercepting several important data points.
// The pipeline for this node will be started when calling this function.
func StartNewNode(id uint32) *Node {
	n := new(Node)

	s := StartMockWiFly()
	c := pipeline.DefaultConfig(id, s)

	n.LocalPackets = InterceptLocal(&c)
	n.IncomingPackets = InterceptIncoming(&c)
	rch := InterceptRouter(&c)

	pipeline.Start(c)

	n.Router = <-rch // r1 is now the router for node 1
	return n
}

// Link establishes a bidirectional communication line between two nodes. It
// acts as a wrapper for MockWiFly Link function.
func (n *Node) Link(n2 *Node) {
	n.mockWiFly.Link(n2.mockWiFly)
}

// InterceptRouter takes a configuration file and returns a channel that will
// eventually provide the router created when a pipeline is run with the
// provided configuration. It does this by wrapping the RoutePackets function in
// the configuration object.
func InterceptRouter(config *pipeline.Config) <-chan pipeline.Router {
	ch := make(chan pipeline.Router, 1)
	log.Info("InterceptRouter")

	prevRoutePackets := config.RoutePackets

	config.RoutePackets = func(localID uint32, toForward <-chan packet.Packet, destLocal chan<- packet.Packet, out chan<- packet.Packet) pipeline.Router {
		log.Info("NewRoutePackets")
		r := prevRoutePackets(localID, toForward, destLocal, out)
		ch <- r
		return r
	}

	return ch
}

func InterceptIncoming(config *pipeline.Config) <-chan packet.Packet {
	ch := make(chan packet.Packet, 100)

	prevParsePackets := config.ParsePackets

	config.ParsePackets = func(in <-chan packet.RawPacket, out chan<- packet.Packet) {
		splitter := make(chan packet.Packet)
		go func() {
			for c := range splitter {
				ch <- c
				out <- c
			}
		}()
		prevParsePackets(in, splitter)
	}

	return ch
}

// TODO careful about buffered channel size
func InterceptLocal(config *pipeline.Config) <-chan packet.Packet {
	ch := make(chan packet.Packet, 100)

	config.LocalProcessing = func(destLocal <-chan packet.Packet, router pipeline.Router) {
		go func() {
			for c := range destLocal {
				ch <- c
			}
		}()
	}

	return ch
}
