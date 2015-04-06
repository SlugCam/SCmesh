package main

import (
	"fmt"
	"net"
)

const DSR_PING = `
{
	"command": "ping-dsr",
	"options": {
		"destination": 0
	}
}
`
const FLOOD_PING = `
{
	"command": "ping-flood",
	"options": {
		"TTL": 255
	}
}
`

func ping(conn net.Conn, flood bool) {
	if flood {
		fmt.Fprintf(conn, FLOOD_PING)
	} else {
		fmt.Fprintf(conn, DSR_PING)
	}
}
