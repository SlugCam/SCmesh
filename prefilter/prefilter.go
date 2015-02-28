// Contains the raw stream parser for the WiFly UART connection for the SlugCam
// project. Used some ideas from
// http://blog.gopheracademy.com/advent-2014/parsers-lexers/.
package prefilter

import (
	"bufio"
	"bytes"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lelandmiller/SCcomm/packet" // A replacement for the stdlib log
)

type token int

// TODO magic number
const TIMEOUT = 50 * time.Millisecond

// Constants for the escape sequences for the parser. IMPORTANT, PACK_SEQ and
// COMM_SEQ need to be the same length and must both begin with the letter 'C'.
// If this requirement needs to change then checkEscape() will have to be
// modified.
const (
	PACK_SEQ = "CBU\r\n"
	COMM_SEQ = "CMD\r\n"
	EXIT_SEQ = "EXIT\r\n"
)

const (
	PACK_TOKEN = iota
	COMM_TOKEN
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
	b, err := s.r.Peek(n)
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

// readBytes is just a wrapper for the bufio.Reader function of the same name.
func (s *rawScanner) readBytes(delim byte) (line []byte) {
	line, err := s.r.ReadBytes(delim)
	if err != nil {
		log.Panic(err)
	}
	return
}

// checkEscape returns true and the escape sequence if it is an escape, otherwise
// returns false and token is invalid.
func (s *rawScanner) checkEscape() (found bool, val token) {
	// TODO maybe this should not be hardcoded?
	// First check if character is a C
	first := s.peek(1)
	if first[0] != 'C' {
		found = false
		return
	}

	ch := make(chan []byte)
	go func() { ch <- s.peek(len(COMM_SEQ)) }()
	select {
	case b := <-ch:
		switch {
		case bytes.Equal(b, []byte(COMM_SEQ)):
			found = true
			val = COMM_TOKEN
		case bytes.Equal(b, []byte(PACK_SEQ)):
			found = true
			val = PACK_TOKEN
		default:
			found = false
		}

	case <-time.After(TIMEOUT):
		found = false
	}
	return
}

// readCommandLines does the parsing for command mode. Since in command mode the
// WiFly output is well defined this function is simpler than scanning in data
// mode. All we have to do is get lines and look for the exit sequence. Times
// out while seeking rest of escape. If it is an escape sequence it should not
// take much time to receive the whole thing.
func (s *rawScanner) readCommandLines(responseLines chan<- []byte) {
	for {
		b := s.readBytes('\n')
		responseLines <- bytes.TrimSpace(b)
		if bytes.Equal(b, []byte(EXIT_SEQ)) {
			break
		}
	}
}

// TODO this and write tests
// readRawPacket scans a raw packet. Packet size is assumed to be the WiFly
// maximum of 1460.
func (s *rawScanner) readRawPacket(rawPackets chan<- []byte) {
	b := make([]byte, 0, packet.RAW_PACKET_SIZE)

	// First read the command sequence at the beginning of the packet
	for i := 0; i < len(PACK_SEQ); i++ {
		b = append(b, s.read())
	}

	// Now read the rest of the packet
	for len(b) < packet.RAW_PACKET_SIZE {
		found, _ := s.checkEscape()
		if found {
			log.Info("readRawPacket: packet discarded because command sequence encountered")
			return
		} else {
			b = append(b, s.read())
		}
	}
	rawPackets <- b
}

func Prefilter(in io.Reader) (rawPackets <-chan []byte, responseLines <-chan []byte) {
	// Make the channels and set the return channels, this allows us to use the
	// full value of the channels but return read only channels.
	packets := make(chan []byte, 500)
	rawPackets = packets

	responses := make(chan []byte, 500)
	responseLines = responses

	go func() {
		log.Debug("Prefilter has begun scanning input")
		s := newRawScanner(in)
		// TODO, should they be declared outside?
		for {
			// Check if it might be an escape sequence
			found, token := s.checkEscape()
			if found {
				switch token {
				case COMM_TOKEN:
					log.Printf("Prefilter command sequence detected")
					s.readCommandLines(responses)
				case PACK_TOKEN:
					log.Printf("Packet UDP command sequence detected")
					s.readRawPacket(packets)
				}
			} else {
				s.read()
			}
		}
	}()

	return
}
