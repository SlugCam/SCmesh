package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

const POWER_SOCKET = "/tmp/paunix.str"
const SCMESH_CTRL = "/tmp/scmeshctrl.str"

const HELP = `help - print this message
batt - print battery life
`

type PowerReq struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type PowerResp struct {
	Type string           `json:"type"`
	Data *json.RawMessage `json:"data"`
}

func main() {
	fmt.Fprintf(os.Stderr, "scclient, enter help to see help\n")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		command := scanner.Text()
		args := strings.Fields(command)
		// If no arguments skip
		if len(args) == 0 {
			continue
		}
		// Skip comment line
		if args[0][0] == '#' {
			continue
		}
		switch args[0] {
		case "help":
			fmt.Fprintf(os.Stderr, HELP)
		case "flood-ping":
			ping(true)
		case "dsr-flood":
			ping(false)
		case "batt":
			getBattery()
		default:
			fmt.Fprintf(os.Stderr, "command \"%s\" not recognized\n", command)
		}
		fmt.Print("> ")
	}

	err := scanner.Err()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error scanning stdin:", err)
		os.Exit(1)
	}

	/*
		conn, err := net.Dial("unixpacket", "/tmp/sc")
		//check error TODO
		fmt.Println("Hello")
		fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
		status, err := bufio.NewReader(conn).ReadString('\n')
	*/
}

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
	status, err := bufio.NewReader(conn).ReadBytes('\r')

	jpow := new(map[string]interface{})
	err = json.Unmarshal(status, jpow)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling request to json: %s\n", err)
		return
	}

	fmt.Println("RETURN:", jpow)
}

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

	if flood {
		fmt.Fprintf(conn, FLOOD_PING)
	} else {
		fmt.Fprintf(conn, DSR_PING)

	}
}
