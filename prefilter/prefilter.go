// Contains the raw stream parser for the WiFly UART connection for the SlugCam
// project. Used some ideas from
// http://blog.gopheracademy.com/advent-2014/parsers-lexers/.
package prefilter

import (
	"bufio"
	"errors"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
)

// scanner represents a lexical scanner.
type scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

// read reads the next byte from the bufferred reader.
func (s *scanner) read() byte {
	c, err := s.r.ReadByte()
	if err != nil {
		log.Panic(err)
	}
	return c
}

// unread places the previously read byte back on the reader.
func (s *scanner) unread() {
	err := s.r.UnreadByte()
	if err != nil {
		log.Panic(err)
	}
}

// readUntil reads until either the packet delimiter or the argument c is
// detected. If c is detected err is nil, data contains the data read, and c is
// consumed. If the packet delimiter is detected err is not nil and data should
// be ignored.
func (s *scanner) readUntil(delim byte) (data []byte, err error) {
	for {
		c := s.read()
		if c == '\x01' {
			s.unread()
			err = errors.New("readUntil: packet delimiter encountered")
			return
		} else if c == delim {
			break
		} else {
			data = append(data, c)
		}
	}
	return
}

// TODO discard packet if MAX_PACKET_LEN reached!!!
func (s *scanner) readRawPacket() (p packet.RawPacket, err error) {
	start := time.Now()
	preheader, err := s.readUntil('\x00')
	if err != nil {
		return
	}

	header, err := s.readUntil('\x00')
	if err != nil {
		return
	}

	payload, err := s.readUntil('\x04')
	if err != nil {
		return
	}

	p.Preheader = preheader
	p.Header = header
	p.Payload = payload

	end := time.Now()
	log.Info("Scanned packet in: ", end.Sub(start))
	return
}

func Prefilter(in io.Reader, out chan<- packet.RawPacket) {
	go func() {
		log.Debug("Prefilter has begun scanning input")
		s := newScanner(in)
		for {
			c := s.read()
			if c == '\x01' {
				p, err := s.readRawPacket()
				if err != nil {
					log.Info("Packet discarded due to error:", err)
				} else {
					out <- p
				}
			}
		}
	}()
}
