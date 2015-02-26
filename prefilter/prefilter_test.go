package prefilter

import (
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
	r := util.NewMockReader()
	pack, resp := Prefilter(r)
	cmdValidator := util.MakeSequenceValidator([]string{"CMD", "This line"})
	go func() {
		r.Write([]byte("saldkfjl\x00\x04\x1C"))
		time.Sleep(50 * time.Millisecond)
		r.Write([]byte("asldkfjCMD\r\nThis line\r\n"))
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-pack:
			t.Errorf("Received from packet channel, should have produced nothing")
		case r := <-resp:
			valid, mess := cmdValidator(string(r))
			if !valid {
				t.Errorf(mess)
			}
		case <-time.After(timeoutDelay):
			t.Errorf("Did not receive needed response line")
		}
	}
}

// TestPrefilterCommandLine tests if we can receive one response line from
// command mode.
func TestPrefilterExitCommandLine(t *testing.T) {
	r := util.NewMockReader()
	_, resp := Prefilter(r)
	cmdValidator := util.MakeSequenceValidator([]string{"CMD", "TEST", "EXIT", "CMD", "This line"})
	go func() {
		r.Write([]byte("CMD\r\nTEST\r\nEXIT\r\nasldfjlkasdfj\r\nasdkfj"))
		time.Sleep(50 * time.Millisecond)
		r.Write([]byte("asldkfjCMD\r\nThis line\r\n"))
	}()
	for i := 0; i < 3; i++ {
		select {
		case r := <-resp:
			valid, mess := cmdValidator(string(r))
			if !valid {
				t.Errorf(mess)
			}
		case <-time.After(timeoutDelay):
			t.Errorf("Did not receive needed response line")
		}
	}
}

// TODO test prefilter timeout on escape
