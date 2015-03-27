package ipc

import (
	"encoding/json"
	"io"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/pipeline"
)

const SCMESH_CTRL = "/tmp/scmeshctrl.str"

type Command struct {
	Command  string
	DataType string
	Options  json.RawMessage
	Data     json.RawMessage
}

func ListenIPC(r pipeline.Router) {

	mchan := make(chan Command)
	listenClients(SCMESH_CTRL, mchan)

	go func() {
		log.Debug("scanning mchan")
		for m := range mchan {
			log.Debug("Received message in LocalProcessing")
			switch m.Command {
			case "ping-flood":
				ping(r, m.Options, PING_FLOOD)
			case "ping-dsr":
				ping(r, m.Options, PING_DSR)
			}
		}
		log.Debug("no longer scanning mchan")
	}()
}

// TODO should only accept from localhost
func listenClients(port string, mchan chan<- Command) {
	go func() {
		os.Remove(port)
		// TODO could change to unix socket
		ln, err := net.Listen("unix", port)
		if err != nil {
			log.Fatal(err)
		}
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {

				log.WithFields(log.Fields{
					"error": err,
				}).Fatal("Error in TCP command connection listener")

			}
			go handleConnection(conn, mchan)
		}
	}()
}

// TODO terminate gracefully
// TODO determine if trimming behavior is correct
func handleConnection(c net.Conn, mchan chan<- Command) {
	dec := json.NewDecoder(c)
	//enc := json.NewEncoder(c)
	for {
		log.Debug("handleConnection running")
		comm := new(Command)
		err := dec.Decode(comm)
		if err != nil {
			if err != io.EOF {
				log.Error("control connection error: ", err)
			}
			break
		}
		mchan <- *comm
	}

	c.Close()
}
