package wiflyparsers

import (
	"encoding/gob"
	"log"
	//"encoding/json"
	"fmt"
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

func ParseInput(in io.Reader) {
	dec := gob.NewDecoder(in)
	// Decode (receive) and print the values.
	var q Packet
	for {
		err := dec.Decode(&q)
		if err != nil {
			log.Print("decode error 1:", err)
		}
		fmt.Printf("Packet: %q\n", q.Payload)
	}
}

// WriteOutput handles the writing of packets to the WiFly module.
func WriteOutput(out io.Writer, in <-chan Packet) {
	enc := gob.NewEncoder(out)

	for c := range in {
		err := enc.Encode(c)
		if err != nil {
			log.Fatal("encode error:", err)
		}
	}
	/*
		for c := range in {
			// TODO check error
			b, _ := json.Marshal(c)
			out.Write(b)
		}
	*/

}
