package local

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/routing"
)

const SCMESH_CTRL = "/tmp/scmeshctrl.str"

type Command struct {
	Command  string
	DataType string
	Options  *json.RawMessage
	Data     *json.RawMessage
}

type PingOptions struct {
	Destination uint32
	TTL         uint32
}

func LocalProcessing(in <-chan packet.Packet, router pipeline.Router) {
	mchan := make(chan Command)
	listenClients(SCMESH_CTRL, mchan)

	go func() {
		log.Debug("scanning mchan")
		for m := range mchan {
			log.Debug("Received message in LocalProcessing")
			switch m.Command {
			case "ping":
				log.Info("ping request received.")
				dh := header.DataHeader{
					FileId:       proto.Uint32(0),
					Destinations: []uint32{routing.BroadcastID},
				}
				router.OriginateFlooding(20, dh, []byte("Ping!!!"))
			}
			fmt.Println(m)

		}
		log.Debug("no longer scanning mchan")
	}()

	go func() {
		for c := range in {
			log.Info("Packet received:", c)
		}
	}()
}

// TODO should only accept from localhost
func listenClients(port string, mchan chan<- Command) {
	go func() {
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
		log.Debug("Decoded:", comm)
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
