package main

import (
	"bufio"
	"fmt"
	"github.com/lelandmiller/gowifly"
	"log"
	"net"
	"sync"
)

func listenClients() {
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
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	reader := bufio.NewReader(c)
	for {
		reply, err := reader.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		fmt.Println(string(reply))
	}
	c.Close()
}

func main() {
	go listenClients()
	w := gowifly.Setup()
	w.serialConn.write("GOGOGO\r")

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
