// Contains the raw stream parser for the WiFly UART connection for the SlugCam
// project. Used some ideas from
// http://blog.gopheracademy.com/advent-2014/parsers-lexers/.
package wiflyparsers

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"time"
)

type token int

// TODO magic number
const TIMEOUT = 50 * time.Millisecond

const (
	PACKET_ESCAPE_SEQUENCE  = "CBU\r\n"
	COMMAND_ESCAPE_SEQUENCE = "CMD\r\n"
	ESCAPE_SEQUENCE_LENGTH  = 5
)

const (
	ENTER_PACKET_TOKEN = iota
	ENTER_COMMAND_TOKEN
)

// rawScanner represents a lexical scanner.
type rawScanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func newRawScanner(r io.Reader) *rawScanner {
	return &rawScanner{r: bufio.NewReader(r)}
}

// peek peeks n bytes
func (s *rawScanner) peek(n int) []byte {
	b, err := s.r.Peek(ESCAPE_SEQUENCE_LENGTH)
	if err != nil {
		log.Panic(err)
	}
	return b
}

// read reads the next byte from the bufferred reader.
func (s *rawScanner) read() byte {
	c, err := s.r.ReadByte()
	if err != nil {
		log.Panic(err)
	}
	return c
}

// unread places the previously read byte back on the reader.
func (s *rawScanner) unread() {
	err := s.r.UnreadByte()
	if err != nil {
		log.Panic(err)
	}
}

// readEscape returns true and the escape sequence if it is an escape, otherwise
// returns false and token is invalid. If it returns true it will consume the
// input for the escape sequence otherwise it will not.
func (s *rawScanner) readEscape() (found bool, val token) {
	ch := make(chan []byte)
	go func() { ch <- s.peek(ESCAPE_SEQUENCE_LENGTH) }()
	select {
	case b := <-ch:
		switch {
		case bytes.Equal(b, []byte(COMMAND_ESCAPE_SEQUENCE)):
			found = true
			val = ENTER_COMMAND_TOKEN
		case bytes.Equal(b, []byte(PACKET_ESCAPE_SEQUENCE)):
			found = true
			val = ENTER_PACKET_TOKEN
		default:
			found = false
		}

	case <-time.After(TIMEOUT):
		found = false
	}
	return
}

func Prefilter(in io.Reader) (rawPackets <-chan []byte, responseLines <-chan []byte) {
	// Make the channels and set the return channels, this allows us to use the
	// full value of the channels but return read only channels.
	packets := make(chan []byte, 500)
	rawPackets = packets

	responses := make(chan []byte, 500)
	responseLines = responses

	go func() {
		s := newRawScanner(in)
		for {
			// TODO, should they be declared outside?
			c := s.read()
			// Check if it might be an escape sequence
			if c == 'C' {
				s.unread()
				found, token := s.readEscape()
				if found {
					switch token {
					case ENTER_COMMAND_TOKEN:
					case ENTER_PACKET_TOKEN:
					}
				} else {
					s.read()
				}
			}
		}
	}()

	return
}

/* scanRawPacket
func scanRawPacket() (p []byte, err error) {
	// Read length

	// Read until escape or length
	return nil, nil

}
*/
