package main

import (
	"flag"
	"io"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/local"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/routing"
	"github.com/tarm/goserial" // A replacement for the stdlib log
)

func main() {
	_ = flag.Int("port", 8080, "the port on which to listen for control messages")
	localId := flag.Int("local-id", 0, "the id number for this node, sinks are 0")
	debug := flag.Bool("debug", false, "print debug level log messages")
	// program := flag.String("program", "SCcomm", "the program to run")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	startPipeline(uint32(*localId))

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func startPipeline(localId uint32) {
	log.Info("Starting SCmesh")

	// Setup serial
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 115200}
	serial, err := serial.OpenPort(c)
	if err != nil {
		log.Panic(err)
	}

	// Make channels
	rawPackets := make(chan packet.RawPacket)
	toRouter := make(chan packet.Packet)
	destLocal := make(chan packet.Packet)
	fromRouter := make(chan packet.Packet)
	packedPackets := make(chan []byte)

	// Setup pipeline
	prefilter.Prefilter(serial, rawPackets)

	packet.ParsePackets(rawPackets, toRouter)

	r := routing.RoutePackets(localId, toRouter, destLocal, fromRouter)
	local.LocalProcessing(destLocal, r)

	packet.PackPackets(fromRouter, packedPackets)

	writePackets(packedPackets, serial)

}

func writePackets(in <-chan []byte, out io.Writer) {
	// TODO flush data stream with '\x04'
	go func() {
		for c := range in {
			out.Write(c)
		}
	}()
}
