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
