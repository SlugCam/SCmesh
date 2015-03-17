package main

import (
	"flag"
	"io"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/local"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/routing"
	"github.com/tarm/serial"
)

func main() {
	_ = flag.Int("port", 8080, "the port on which to listen for control messages")
	localID := flag.Int("local-id", 0, "the id number for this node, sinks are 0")
	debug := flag.Bool("debug", false, "print debug level log messages")
	// program := flag.String("program", "SCcomm", "the program to run")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Setup serial
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 115200}
	serial, err := serial.OpenPort(c)
	if err != nil {
		log.Panic(err)
	}

	conf := pipeline.Configuration{
		LocalID:         uint32(*localID),
		Serial:          serial,
		Prefilter:       prefilter.Prefilter,
		ParsePackets:    packet.ParsePackets,
		RoutePackets:    routing.RoutePackets,
		LocalProcessing: local.LocalProcessing,
		PackPackets:     packet.PackPackets,
		WritePackets:    writePackets,
	}

	pipeline.Start(conf)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func writePackets(in <-chan []byte, out io.Writer) {
	out.Write([]byte{'\x04'}) // Send any extraneous data
	go func() {
		for c := range in {
			out.Write(c)
		}
	}()
}
