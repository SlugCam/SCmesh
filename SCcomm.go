package main

import (
	"bufio"
	"flag"
	"fmt"

	"net"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus" // A replacement for the stdlib log

	"github.com/lelandmiller/SCcomm/gowifly"
	"github.com/lelandmiller/SCcomm/prefilter"
)

func main() {
	log.SetLevel(log.DebugLevel)
	port := flag.Int("port", 8080, "the port on which to listen for control messages")
	flag.Parse()

	log.Info("Starting SCcomm")
	// TODO is buffer size good? What happens if buffer full. Looked it up, it
	// should block

	// This is listener for control messages
	mchan := make(chan string, 500)
	go listenClients(*port, mchan)

	// Setup Wifly
	w := gowifly.NewWiFlyConnection()

	// Setup input
	rawPackets, responseLines := prefilter.Prefilter(*w.Stream())

	// Print raw data
	go func() {
		for {
			select {
			case p := <-rawPackets:
				log.Debugf("Received packet: %#v", string(p))
			case r := <-responseLines:
				fmt.Printf("Received response: %#v", string(r))
				log.Debugf("Received response: %#v", string(r))
			}
		}
	}()

	for m := range mchan {
		switch m {
		case "$$$":
			w.EnterCommandMode()
		default:
			w.WriteCommand(m)
			log.Printf("Entered command: %#v", m)
		}
	}

	/*
		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
	*/
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
			log.Fatal(err)
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
			log.Panic(err)
		}
		mchan <- strings.Trim(string(reply), "\n\r ")
	}
	c.Close()
}
