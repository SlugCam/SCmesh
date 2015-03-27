package main

import (
	"fmt"
	"net"
	"os"
)

const DSR_PING = `
{
	"command": "dsr-ping",
	"options": {
		"destination": 0
	}
}
`
const FLOOD_PING = `
{
	"command": "flood-ping",
	"options": {
		"TTL": 255
	}
}
`

func ping(flood bool) {
	conn, err := net.Dial("unix", SCMESH_CTRL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening power socket: %s\n", err)
		return
	}
	defer conn.Close()

	if flood {
		fmt.Fprintf(conn, FLOOD_PING)
	} else {
		fmt.Fprintf(conn, DSR_PING)

	}
}
