package wiflyparsers

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lelandmiller/SCcomm/util"
)

// When we should timeout
const timeoutDelay = 500 * time.Millisecond

// TestPrefilterNothing tests to make sure that the prefilter will return
// nothing if it never detects an command string.
func TestPrefilterNothing(t *testing.T) {
	r := util.NewMockReader()
	pack, resp := Prefilter(r)
	go func() {
		r.Write([]byte("saldkfjl\x00\x04\x1C"))
		time.Sleep(100 * time.Millisecond)
		r.Write([]byte("asldkfj"))
	}()
	select {
	case <-resp:
		t.Errorf("Received from responses channel, should have produced nothing")
	case <-pack:
		t.Errorf("Received from packet channel, should have produced nothing")
	case <-time.After(timeoutDelay):
	}
}

// TestPrefilterCommandLine tests if we can receive one response line from
// command mode.
func TestPrefilterCommandLine(t *testing.T) {
	r := strings.NewReader("saldkfjl\x00\x04\x1CCMD\r\nThis line\r\n")
	pack, resp := Prefilter(r)
	select {
	case <-pack:
		t.Errorf("Received from packet channel, should have produced nothing")
	case r := <-resp:
		if !bytes.Equal(r, []byte("This line")) {
			t.Errorf("Received incorrect response: %#v", string(r))
		}
	case <-time.After(timeoutDelay):
		t.Errorf("Did not receive response line")
	}
}

// TODO test prefilter timeout on escape
