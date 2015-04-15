package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/SlugCam/SCmesh/local/escrow"
)

type Command struct {
	Command string      `json:"command"`
	Options interface{} `json:"options"`
}

type OutboundMessage struct {
	Id   int         `json:"id"`
	Cam  string      `json:"cam"`
	Time time.Time   `json:"time"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func escrowPing(conn net.Conn) {
	pingComm := OutboundMessage{
		Id:   0,
		Cam:  "ME",
		Time: time.Now(),
		Type: "ping",
		Data: interface{}("Hello!"),
	}
	var jsonPing json.RawMessage
	var err error
	jsonPing, err = json.Marshal(pingComm)
	rr := Command{
		Command: "register",
		Options: escrow.RegistrationRequest{
			DataType:    "message",
			Destination: uint32(0),
			Timestamp:   time.Now(),
			JSON:        &jsonPing,
		},
	}
	b, err := json.Marshal(&rr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error making registration request: %s\n", err)
		return
	}
	_, err = conn.Write(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing registration request: %s\n", err)
		return
	}
}

func sendVideo(conn net.Conn, path string) {
	rr := Command{
		Command: "register",
		Options: escrow.RegistrationRequest{
			DataType:    "video",
			Destination: uint32(0),
			Timestamp:   time.Now(),
			Path:        path,
		},
	}
	b, err := json.Marshal(&rr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error making registration request: %s\n", err)
		return
	}
	_, err = conn.Write(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing registration request: %s\n", err)
		return
	}

}
