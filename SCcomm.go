package main

import (
	"bufio"
	"flag"
	"github.com/lelandmiller/SCcomm/gowifly"
	"log"
	"net"
	"strconv"
	"strings"
)

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

func MakePacket(payload []byte)

func main() {
	port := flag.Int("port", 8080, "the port on which to listen for control messages")
	flag.Parse()
	// TODO is buffer size good? What happens if buffer full. Looked it up, it
	// should buffer
	mchan := make(chan string, 500)
	go listenClients(*port, mchan)
	w := gowifly.NewWiFlyConnection()
	//w.serialConn.write("GOGOGO\r")
	w.Write("Hello!")

	//var m string
	for {
		m := <-mchan
		switch m {
		case "comm":
			w.EnterCommandMode()
		default:
			w.WriteCommand(m)
		}
	}

	/*
		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
	*/
}
