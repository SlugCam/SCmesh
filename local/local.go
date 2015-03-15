package local

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/routing"
)

func LocalProcessing(in <-chan packet.Packet, router *routing.Router) {
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
			//h := packet.NewPacket()
			//p.Header.Type = proto.Int32(1)
			//p.Payload = []byte("Ping!!!")
			//out <- *p
			time.Sleep(10 * time.Second)
		}

	}()

}

// TODO should only accept from localhost
func listenClients(port int, mchan chan<- string) {
	// TODO could change to unix socket
	// ln, err := net.Listen("tcp", "localhost:8080")
	ln, err := net.Listen("tcp", strings.Join([]string{"localhost:", strconv.Itoa(port)}, ""))
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
func handleConnection(c net.Conn, mchan chan<- string) {
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
		mchan <- strings.Trim(string(reply), "\n\r ")
	}
	c.Close()
}
