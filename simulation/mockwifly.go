package simulation

import (
	"bytes"
	"log"
	"sync"
)

// MockWiFly is a ReadWriter that can emulate our WiFly serial connections. We
// can make multiple MockWiFly objects, connect them together, create SCmesh
// pipelines with them, and simulate a network of SlugCams. This mock is not
// currently accurate to the WiFly and should not be used for testing
// interfacing with the WiFly itself. Instead this is used to test higher level
// operation of the routing algorithms.
type MockWiFly struct {
	sharedCh    chan<- []byte
	connections []chan<- []byte
	incomingBuf *bytes.Buffer
	outgoingBuf *bytes.Buffer
	readCond    *sync.Cond
}

// StartMockWiFly returns a new MockWiFly object and begins listening for
// messages from linked nodes.
func StartMockWiFly() *MockWiFly {
	sharedCh := make(chan []byte)

	m := &MockWiFly{
		sharedCh:    sharedCh,
		connections: make([]chan<- []byte, 0),
		incomingBuf: new(bytes.Buffer),
		outgoingBuf: new(bytes.Buffer),
		readCond:    sync.NewCond(new(sync.Mutex)),
	}

	// Start
	go func() {
		for c := range sharedCh {
			// Buffer input from external node
			m.readCond.L.Lock()
			m.incomingBuf.Write(c)
			m.readCond.Broadcast()
			m.readCond.L.Unlock()
		}

	}()

	return m
}

// Write is an implementation method for io.Writer. Unlike the real WiFly
// module, our write sends p over the wire immediately instead of waiting for
// the packet termination character. This will work for our simulations since
// our code always writes a whole packet in each call to write.
func (m *MockWiFly) Write(p []byte) (n int, err error) {
	// Loop through all the linked channels sending our message to each
	for _, ch := range m.connections {
		b := make([]byte, len(p))
		n := copy(b, p)
		if n != len(p) {
			log.Panic("failure in MockWiFly.Write, failed to make copy of []byte")
		}
		ch <- b
	}
	n = len(p)
	err = nil
	return
}

// Write is an implementation method for io.Reader. Since we use a blocking
// feature in our serial library, we will implement that here as well. This
// means that we will block until we can read at least 1 byte, after which we
// will return what we have.
func (m *MockWiFly) Read(p []byte) (n int, err error) {

	m.readCond.L.Lock()
	for {
		n, err = m.incomingBuf.Read(p)
		if n != 0 {
			break
		}
		m.readCond.Wait()
	}
	m.readCond.L.Unlock()

	return
}

// Link connects two MockWiFly objects together. The process of linking nodes is
// not goroutine-safe. The typical workflow would involve creating the objects
// and linking them in a single goroutine, then freely reading and writing to
// them from multiple goroutines. This method will create a bidirectional link
// as that is all we need to test our module. This could easily be changed if
// needed later.
func (n1 *MockWiFly) Link(n2 *MockWiFly) {
	n1.connections = append(n1.connections, n2.sharedCh)

}
