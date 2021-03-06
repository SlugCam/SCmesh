package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"sync"

	_ "net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/config"
	"github.com/SlugCam/SCmesh/local/gateway"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/routing/dsr"
	"github.com/SlugCam/SCmesh/simulation"
	"github.com/tarm/serial"
)

func main() {
	go func() {
		log.Info(http.ListenAndServe(":6060", nil))
	}()

	// Parse command flags
	localID := flag.Int("id", 0, "the id number for this node, sinks are 0")
	baudRate := flag.Int("baud", 115200, "the baud rate for the serial connection")
	debug := flag.Bool("debug", false, "print debug level log messages")
	gwFlag := flag.Bool("gw", false, "run this server as a gateway")
	messageServer := flag.String("ms", "localhost:7892", "address for the message server")
	videoServer := flag.String("vs", "localhost:7893", "address for the video server")
	serialDev := flag.String("serial", "/dev/ttyAMA0", "path of the serial device to use")
	// TODO this is tacky
	packetLog := flag.String("plog", "none", "path to log all packets to, or none for no logging")
	cost := flag.Int("cost", 0, "the cost of the node")
	flag.Parse()

	packet.LocalID = uint32(*localID)
	dsr.Cost = *cost

	// Modify logging level
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Setup serial
	c := &serial.Config{Name: *serialDev, Baud: *baudRate}
	serial, err := serial.OpenPort(c)
	if err != nil {
		log.Panic(err)
	}

	conf := config.DefaultConfig(uint32(*localID), serial)

	// Setup packet logging if desired
	if *packetLog == "none" {
		log.Info("not logging incoming packets")
	} else {
		incomingPath := fmt.Sprintf("%s.in", *packetLog)
		outgoingPath := fmt.Sprintf("%s.out", *packetLog)

		// Outgoing packets
		log.Infof("logging outgoing packets to: %s", outgoingPath)
		fo, err := os.OpenFile(outgoingPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			log.Error("error opening packet log file: ", err)
		}

		// TODO is this correct?
		defer fo.Close()
		enco := json.NewEncoder(fo)

		recordPacket := func(p packet.Packet) {
			err := enco.Encode(p.Abbreviate())
			if err != nil {
				log.Error("error logging packet: ", err)
			}
			err = fo.Sync()
			if err != nil {
				log.Error("error logging packet: ", err)
			}
		}

		packet.SendingLogCallback = &recordPacket

		// Incoming packets
		log.Infof("logging incoming packets to: %s", incomingPath)
		incoming := simulation.InterceptIncoming(&conf)

		f, err := os.OpenFile(incomingPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			log.Error("error opening packet log file: ", err)
		}

		// TODO is this correct?
		defer f.Close()
		enc := json.NewEncoder(f)

		go func() {
			for p := range incoming {
				// log packet
				err := enc.Encode(p.Abbreviate())
				if err != nil {
					log.Error("error logging packet: ", err)
				}
				err = f.Sync()
				if err != nil {
					log.Error("error logging packet: ", err)
				}
			}
		}()
	}

	// Check if gateway
	if *gwFlag {
		gw := &gateway.Gateway{
			MessageAddress: *messageServer,
			VideoAddress:   *videoServer,
		}
		conf.LocalProcessing = gw.LocalProcessing
	}

	pipeline.Start(conf)

	// Block forever
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func appendToFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
