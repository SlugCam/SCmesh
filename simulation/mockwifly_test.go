package simulation

import (
	"bytes"
	"testing"
	"time"
)

// TestMockWiFly is a very simple test. A more thorough test could be created.
func TestMockWiFly(t *testing.T) {
	data := []byte("Testing MockWiFly\x04")

	m1 := StartMockWiFly()
	m2 := StartMockWiFly()
	m3 := StartMockWiFly()

	m1.Link(m2)

	n, err := m1.Write(data)

	if err != nil || n != len(data) {
		t.Error("could not write to MockWiFly.")
	}

	m2Ch := make(chan []byte)
	m3Ch := make(chan []byte)

	go func() {
		b := make([]byte, 100)
		n, err := m2.Read(b)
		if err != nil {
			t.Error("reading from m2 returned:", err)
		}
		m2Ch <- b[:n]
	}()

	go func() {
		b := make([]byte, 100)
		n, err := m3.Read(b)
		if err != nil {
			t.Error("reading from m3 returned:", err)
		}
		m3Ch <- b[:n]
	}()

	time.Sleep(10 * time.Millisecond)

	// Should have received data on m2
	select {
	case c := <-m2Ch:
		if !bytes.Equal(c, data) {
			t.Errorf("m2 read %v, wanted %v.", c, data)
		}
	default:
		t.Error("m2 did not read data written to linked node.")
	}

	// Should not receive anything else
	select {
	case <-m2Ch:
		t.Errorf("m2 received extraneous data.")
	default:
	}

	// m3 should not receive anything
	select {
	case <-m3Ch:
		t.Errorf("m3 should not have read any data.")
	default:
	}

}
