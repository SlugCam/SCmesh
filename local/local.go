package local

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

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
		for m := range mchan {
			switch m.Command {
			case "ping":
				log.Info("ping request received.")
			}
			fmt.Println(m)

		}
	}()

	go func() {
		for c := range in {
			log.Info("Packet received:", c)
		}
	}()
	go func() {
		for {
			dh := header.DataHeader{
				FileId:       proto.Uint32(0),
				Destinations: []uint32{routing.BroadcastID},
			}
			router.OriginateFlooding(20, dh, []byte("Ping!!!"))
			time.Sleep(10 * time.Second)
		}

	}()

}

// TODO should only accept from localhost
func listenClients(port string, mchan chan<- Command) {
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
}

// TODO terminate gracefully
// TODO determine if trimming behavior is correct
func handleConnection(c net.Conn, mchan chan<- Command) {
	dec := json.NewDecoder(c)
	//enc := json.NewEncoder(c)
	for {
		comm := new(Command)
		err := dec.Decode(comm)
		if err != nil {
			log.Error("control connection error: ", err)
			break
		}
		mchan <- *comm
	}

	/*
		reader := bufio.NewReader(c)
		for {
			reply, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Error in TCP command connection")
				}
			}
			// mchan <- strings.Trim(string(reply), "\n\r ")
			mchan <- reply
		}
	*/
	c.Close()
}
