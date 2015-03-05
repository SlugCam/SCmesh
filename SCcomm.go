package main

import (
	"bufio"
	"flag"
	"io"

	"net"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus" // A replacement for the stdlib log

	"github.com/lelandmiller/SCcomm/gowifly"
	"github.com/lelandmiller/SCcomm/packet"
	"github.com/lelandmiller/SCcomm/prefilter"
)

func main() {
	log.SetLevel(log.DebugLevel)
	port := flag.Int("port", 8080, "the port on which to listen for control messages")
	// program := flag.String("program", "SCcomm", "the program to run")
	flag.Parse()
	startCommTest()
	/*
		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
	*/
}

/*
func parseHeaders(in <-chan []byte, out chan<- packet.Packet) {
	// Parse headers
	toRouter := make(chan packet.Packet, 500)
	go func() {
		for p := range in {
			out <- packet.NewPacket(p)
		}
	}()
}


func startComm() {
	log.Info("Starting SCcomm")

	// Setup serial
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 115200}
	serial, err := serial.OpenPort(c)
	if err != nil {
		log.Panic(err)
	}

	// Setup prefilter
	rawPackets := prefilter.Prefilter(serial)


	routingOutCh, localCh := routing.RoutePackets(toRouter)
	local.Process(localCh, toRouter)

	// Pack packets
	packedPackets := make(chan []byte, 500)
	go func() {
		for p := range routingOutCh {
			packedPackets <- p.ToWireFormat()
		}
	}()

	// Write output
	go func() {
		for o := range packedPackets {
			serial.Write(o)
		}
	}()
}
*/

func startCommTest() {
	log.Info("Starting SCcomm")
	// TODO is buffer size good? What happens if buffer full. Looked it up, it
	// should block

	// This is listener for control messages
	mchan := make(chan string, 500)

	// Start local processing
	go listenClients(*port, mchan)

	// Setup Wifly
	w := gowifly.NewWiFlyConnection()

	// Setup prefilter
	rawPackets, responseLines := prefilter.Prefilter(*w.Stream())

	// Print raw data
	go func() {
		for {
			select {
			case p := <-rawPackets:
				log.Debugf("Received packet: %#v", string(p))
			case r := <-responseLines:
				log.Debugf("Received response: %#v", string(r))
			}
		}
	}()

	for m := range mchan {
		args := strings.Split(m, " ")
		log.Printf("Entered command: %#v", m)
		switch args[0] {
		case "$$$":
			w.EnterCommandMode()
		case "pack":
			w.WriteRawPacket(&packet.Packet{Payload: []byte(args[1])})
		default:
			w.WriteCommand(m)
		}
	}
}

// TODO should only accept from localhost
func listenClients(port int, mchan chan<- string) {
	// TODO could change to unix socket
	// ln, err := net.Listen("tcp", "localhost:8080")
	ln, err := net.Listen("tcp", strings.Join([]string{"localhost:", strconv.Itoa(port)}, ""))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Error in TCP command connection listener")
		}
		go handleConnection(conn, mchan)
	}
}

// TODO terminate gracefully
// TODO determine if trimming behavior is correct
func handleConnection(c net.Conn, mchan chan<- string) {
	reader := bufio.NewReader(c)
	for {
		reply, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error in TCP command connection")
			}
		}
		mchan <- strings.Trim(string(reply), "\n\r ")
	}
	c.Close()
}
