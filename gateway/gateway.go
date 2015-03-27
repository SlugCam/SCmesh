package gateway

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
)

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

type OutboundMessage struct {
	Id   int          `json:"id"`
	Cam  string       `json:"cam"`
	Time time.Time    `json:"time"`
	Type string       `json:"type"`
	Data *interface{} `json:"data"`
}

type Gateway struct {
	MessageAddress string
	VideoAddress   string
}

func (g *Gateway) LocalProcessing(in <-chan packet.Packet, router pipeline.Router) {
	mchan := make(chan []byte)

	/*
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
						Type:         header.DataHeader_MESSAGE.Enum(),
					}
					router.OriginateFlooding(20, dh, []byte("Ping!!!"))
				}
				fmt.Println(m)

			}
			log.Debug("no longer scanning mchan")
		}()
	*/
	go func() {
		conn, err := net.Dial("tcp", g.MessageAddress)
		if err != nil {
			log.Errorf("error opening message  server connection: %s\n", err)
			return
		}

		//status, err := bufio.NewReader(conn).ReadBytes('\r')

		for m := range mchan {
			log.Debug("message sending to server: ", m)
			fmt.Fprintf(conn, "%s\r\n", m) // NOTE could change to \n?
		}

	}()

	go func() {
		for c := range in {
			log.Info("GW Packet received:", c)
			if c.Header == nil || c.Header.DataHeader == nil {
				continue
			}

			if *c.Header.DataHeader.Type == header.DataHeader_MESSAGE {
				mchan <- c.Payload
				// Forward this to server
			}
		}
	}()
}
