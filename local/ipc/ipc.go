package ipc

import (
	"encoding/json"
	"io"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/local/escrow"
	"github.com/SlugCam/SCmesh/pipeline"
)

const SCMESH_CTRL = "/tmp/scmeshctrl.str"

type Command struct {
	Command string
	Options json.RawMessage
}

func ListenIPC(r pipeline.Router, writeOut <-chan escrow.CollectedData, d *escrow.Distributor) {

	mchan := make(chan Command)
	listenClients(SCMESH_CTRL, mchan, writeOut)

	go func() {
		log.Debug("scanning mchan")
		for m := range mchan {
			log.Debug("Received message in LocalProcessing")
			switch m.Command {
			case "register":
				var rr escrow.RegistrationRequest
				err := json.Unmarshal(m.Options, &rr)
				if err != nil {
					log.Error("ListenIPC: Error processing registration request")
					continue
				}
				_, err = d.Register(rr)
				if err != nil {
					log.Error("ListenIPC: Error submitting registration request")
				}
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
func listenClients(port string, ochan chan<- Command, ichan <-chan escrow.CollectedData) {
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
			go handleConnection(conn, ochan, ichan)
		}
	}()
}

// TODO terminate gracefully
// TODO determine if trimming behavior is correct
func handleConnection(c net.Conn, ochan chan<- Command, ichan <-chan escrow.CollectedData) {
	dec := json.NewDecoder(c)
	enc := json.NewEncoder(c)
	go func() {
		for d := range ichan {
			err := enc.Encode(&d)
			if err != nil {
				log.Error("control connection error encoding message: ", err)
			}
		}
	}()
	for {
		log.Debug("handleConnection running")
		comm := new(Command)
		err := dec.Decode(comm)
		if err != nil {
			if err != io.EOF {
				log.Error("control connection error decoding message: ", err)
			}
			break
		}
		ochan <- *comm
	}

	c.Close()
}
