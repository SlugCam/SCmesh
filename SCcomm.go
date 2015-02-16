package main

import (
	"bufio"
	"github.com/lelandmiller/gowifly"
	"log"
	"net"
	"strings"
)

// TODO should only accept from localhost
func listenClients(mchan chan<- string) {
	ln, err := net.Listen("tcp", ":8080")
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

func main() {
	// TODO is buffer size good? What happens if buffer full. Looked it up, it
	// should buffer
	mchan := make(chan string, 500)
	go listenClients(mchan)
	w := gowifly.NewWiFlyConnection()
	//w.serialConn.write("GOGOGO\r")

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
