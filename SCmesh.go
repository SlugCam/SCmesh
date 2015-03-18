package main

import (
	"flag"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/config"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/tarm/serial"
)

func main() {

	// Parse command flags
	_ = flag.Int("port", 8080, "the port on which to listen for control messages")
	localID := flag.Int("local-id", 0, "the id number for this node, sinks are 0")
	debug := flag.Bool("debug", false, "print debug level log messages")
	flag.Parse()

	// Modify logging level
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

	// Start pipeline
	conf := config.DefaultConfig(uint32(*localID), serial)
	pipeline.Start(conf)

	// Block forever
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
