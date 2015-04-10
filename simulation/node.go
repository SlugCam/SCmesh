// TODO figure out channels so buffering is not needed.
// BUG(lelandmiller@gmail.com): Due to issues with logger and channel buffers,
// strange behavior occurs.
package simulation

import (
	"fmt"
	"html/template"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/config"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
)

const BUFFER_SIZE = 1000

const TEMPLATE = `
<html>
<header><title>SCmesh Log</title></header>
<body>
<h1>SCmesh Log</h1>
{{ range $key, $value := . }}
<h2>{{ $key }}</h2>
<ol>
{{ range $p := $value }}
<li>{{ $p }}</li>
{{end}}
</ol>
{{ end }}
</body>
</html>
`

type logEntry struct {
	id     uint32
	packet packet.Packet
}

type Logger struct {
	in   chan<- logEntry
	data map[uint32][]packet.Packet
}

func StartNewLogger() *Logger {
	l := new(Logger)
	l.data = make(map[uint32][]packet.Packet)
	ch := make(chan logEntry)
	go func() {
		for e := range ch {
			l.data[e.id] = append(l.data[e.id], e.packet)
		}
	}()
	l.in = ch
	return l
}

func (l *Logger) WriteToHTML(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	tmpl, err := template.New("log").Parse(TEMPLATE)
	if err != nil {
		log.Error(err)
		return
	}
	err = tmpl.Execute(f, l.data)
	// TODO
	f.Sync()
}

// Node is a simulated SlugCam mesh node. This is used in integration testing
// and simulation. This struct should be created using the NewNode function.
type Node struct {
	Router          pipeline.Router
	IncomingPackets chan packet.Packet
	LocalPackets    <-chan packet.Packet
	mockWiFly       *MockWiFly
}

func StartNewNode(id uint32) *Node {
	return StartNewNodeLogged(id, nil)
}

// NewNode creates a new simulated node. It does this by creating a default
// pipeline configuration and then intercepting several important data points.
// The pipeline for this node will be started when calling this function.
func StartNewNodeLogged(id uint32, log *Logger) *Node {
	n := new(Node)
	n.mockWiFly = StartMockWiFly()
	c := config.DefaultConfig(id, n.mockWiFly)

	n.LocalPackets = InterceptLocal(&c)
	n.IncomingPackets = InterceptIncoming(&c)
	rch := InterceptRouter(&c)

	pipeline.Start(c)

	n.Router = <-rch // r1 is now the router for node 1

	if log != nil {
		// Intercept incoming for logging
		oldIncoming := n.IncomingPackets
		n.IncomingPackets = make(chan packet.Packet, BUFFER_SIZE)
		go func() {
			for p := range oldIncoming {
				fmt.Println(id, p)
				log.in <- logEntry{id, p}
				n.IncomingPackets <- p
			}
		}()
	}

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

	prevRoutePackets := config.RoutePackets

	config.RoutePackets = func(localID uint32, toForward <-chan packet.Packet, destLocal chan<- packet.Packet, out chan<- packet.Packet) pipeline.Router {
		r := prevRoutePackets(localID, toForward, destLocal, out)
		ch <- r
		return r
	}

	return ch
}

func InterceptIncoming(config *pipeline.Config) chan packet.Packet {
	ch := make(chan packet.Packet, BUFFER_SIZE)

	prevParsePackets := config.ParsePackets

	config.ParsePackets = func(in <-chan packet.RawPacket, out chan<- packet.Packet) {
		splitter := make(chan packet.Packet)
		go func() {
			for c := range splitter {
				select {
				case ch <- c:
				default:
				}
				out <- c
			}
		}()
		prevParsePackets(in, splitter)
	}

	return ch
}

// TODO careful about buffered channel size
func InterceptLocal(config *pipeline.Config) <-chan packet.Packet {
	ch := make(chan packet.Packet, BUFFER_SIZE)

	config.LocalProcessing = func(destLocal <-chan packet.Packet, router pipeline.Router) {
		go func() {
			for c := range destLocal {
				ch <- c
			}
		}()
	}

	return ch
}
