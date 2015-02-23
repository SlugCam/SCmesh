package wiflyparsers

import (
	"encoding/json"
	"io"
)

const (
	initial = iota
	inPacket
)

type Packet struct {
	Payload string
}

// TODO byte at a time or not?

/*
// ParseWiflyStream takes a reader and parses packets and commands from the stream
func ParseWiflyStream(in *io.Reader, delim string) (packet chan string, responses chan string) {
	state := initial

	nextChar := make([]byte, 1)
	bufferChar := make([]byte, 0, max(len(delim)+1, len("CMD\r")))
	for {
		// TODO check err and n
		(*in).Read(nextChar)

		if state == initial {
			append(bufferChar, nextChar...)
		}
	}
}
*/

// WriteOutput handles the writing of packets to the WiFly module.
func WriteOutput(out io.Writer, in <-chan Packet) {
	for c := range in {
		// TODO check error
		b, _ := json.Marshal(c)
		out.Write(b)
	}

}
