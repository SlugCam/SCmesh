package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

const POWER_SOCKET = "/tmp/paunix.str"

type PowerReq struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func main() {

	getBattery()
	/*
		conn, err := net.Dial("unixpacket", "/tmp/sc")
		//check error TODO
		fmt.Println("Hello")
		fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
		status, err := bufio.NewReader(conn).ReadString('\n')
	*/
}

/*
func serve() {

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		//go handleConnection(conn)
	}
}
*/

func getBattery() {
	r := PowerReq{
		Type: "status-request",
		Data: "battery",
	}
	b, err := json.Marshal(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling request to json: %s\n", err)
		return
	}

	conn, err := net.Dial("unix", POWER_SOCKET)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening power socket: %s\n", err)
		return
	}

	fmt.Fprintf(conn, "%s\r", b) // NOTE could change to \n?
	status, err := bufio.NewReader(conn).ReadString('\r')

	fmt.Println("RETURN:", status)
}
