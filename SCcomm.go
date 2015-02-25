package main

import (
	"bufio"
	"flag"
	"github.com/lelandmiller/SCcomm/gowifly"
	//"github.com/lelandmiller/SCcomm/wiflyparsers"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	//"sync"
)

func main() {
	port := flag.Int("port", 8080, "the port on which to listen for control messages")
	flag.Parse()
	// TODO is buffer size good? What happens if buffer full. Looked it up, it
	// should buffer

	// This is listener for control messages
	mchan := make(chan string, 500)
	go listenClients(*port, mchan)

	// Setup Wifly
	w := gowifly.NewWiFlyConnection()
	//w.Write("Hello!")

	// Setup input
	//go wiflyparsers.ParseInput(*w.Stream())

	// Setup output
	// outchan := make(chan wiflyparsers.Packet, 100)
	//go wiflyparsers.WriteOutput(*w.Stream(), outchan)
	//go wiflyparsers.WriteOutput(os.Stdout, outchan)
	//outchan <- wiflyparsers.Packet{Payload: "TEST"}
	//outchan <- wiflyparsers.Packet{Payload: "TEST2"}

	for m := range mchan {
		w.Write(m)
		os.Stdout.Write([]byte(m))
		//outchan <- wiflyparsers.Packet{Payload: m}
		/*
			m := <-mchan
			switch m {
			case "comm":
				w.EnterCommandMode()
			default:
				outchan <- wiflyparsers.Packet{Payload: m}
				w.WriteCommand(m)
			}
		*/
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
	ln, err := net.Listen("tcp", strings.Join([]string{":", strconv.Itoa(port)}, ""))
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
			panic(err)
		}
		mchan <- strings.Trim(string(reply), "\n\r ")
		//fmt.Println(string(reply))
	}
	c.Close()
}
