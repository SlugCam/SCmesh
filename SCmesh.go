package main

import (
	"flag"
	"fmt"
	"io"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/local"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/tarm/goserial" // A replacement for the stdlib log
)

func main() {
	log.SetLevel(log.DebugLevel)
	_ = flag.Int("port", 8080, "the port on which to listen for control messages")
	// program := flag.String("program", "SCcomm", "the program to run")
	flag.Parse()

	startPipeline()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func startPipeline() {
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

	local.LocalProcessing(destLocal, toRouter)

	//routing.RoutePackets(toRouter, destLocal, fromRouter)

	packet.PackPackets(fromRouter, packedPackets)

	writePackets(packedPackets, serial)

	go func() {
		// print incoming
		for c := range toRouter {
			fmt.Println(string(c.Payload))
		}
	}()

}

func writePackets(in <-chan []byte, out io.Writer) {
	// TODO flush data stream with '\x04'
	go func() {
		for c := range in {
			out.Write(c)
		}
	}()
}
