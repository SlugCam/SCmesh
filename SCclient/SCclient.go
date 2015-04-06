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
	fmt.Fprintf(os.Stderr, "SCclient, enter help to see help\n")

	conn, err := net.Dial("unix", SCMESH_CTRL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening SCMESH_CTRL: %s\n", err)
		return
	}
	defer conn.Close()

	// Read incoming messages from connection
	go func() {
		b := make([]byte, 4096)
		for {
			n, err := conn.Read(b)
			fmt.Println(b[:n])
			if err != nil {
				break
			}
		}
	}()

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
		case "ping-flood":
			ping(conn, true)
		case "ping-dsr":
			ping(conn, false)
		case "send-video":
			if len(args) >= 2 {
				sendVideo(conn, args[1])
			} else {
				fmt.Fprintf(os.Stderr, "command \"%s\" missing arguments\n", command)
			}
		case "batt":
			getBattery()
		default:
			fmt.Fprintf(os.Stderr, "command \"%s\" not recognized\n", command)
		}
		fmt.Print("> ")
	}

	err = scanner.Err()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error scanning stdin:", err)
		os.Exit(1)
	}

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
	defer conn.Close()

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
