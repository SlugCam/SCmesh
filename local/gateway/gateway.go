/*
The gateway package provides a connection to the central server and an
interface to an SCmesh network.

The video format is as follows:

1. Camera name (null terminated string)
2. Video ID/Timestamp (uint32 LE)
3. Video Length (uint32 LE)
4. Video Data

The video server then returns an ACK that the data have successfully
been processed and stored.
*/

package gateway

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/local"
	"github.com/SlugCam/SCmesh/local/escrow"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
)

type Gateway struct {
	MessageAddress string
	VideoAddress   string
}

func (g *Gateway) LocalProcessing(in <-chan packet.Packet, router pipeline.Router) {
	collectedData := make(chan escrow.CollectedData, 1024)
	local.LocalProcessingTrackCollected(in, router, collectedData)
	messagePaths := make(chan string, 500)
	g.sendMessagesToServer(messagePaths)

	// Watches the input stream
	go func() {
		for c := range collectedData {
			log.WithFields(log.Fields{
				"collected_data": c,
			}).Info("Gateway received collected data message")

			if c.DataType == "message" {
				messagePaths <- c.Path
			} else if c.DataType == "video" {

			}

		}
	}()
}

// Channel is string of paths
func (g *Gateway) sendMessagesToServer(mchan <-chan string) {
	go func() {
		conn, err := net.Dial("tcp", g.MessageAddress)
		if err != nil {
			log.Errorf("error opening message  server connection: %s\n", err)
			return
		}

		//status, err := bufio.NewReader(conn).ReadBytes('\r')

		for p := range mchan {
			m, err := ioutil.ReadFile(p)
			if err != nil {
				log.Error("Gateway: Error reading file: ", err)
				continue
			}
			log.WithFields(log.Fields{
				"path":     p,
				"contents": string(m),
			}).Info("Gateway sending message to central server")
			fmt.Fprintf(conn, "%s\r\n", m) // NOTE could change to \n?
			// TODO remove file
		}

	}()
}

// sendVideo sends a video to the gateway.
// TODO all the casts should be checked, or more importantly server
// should accept 64bit numbers.
func (g *Gateway) sendVideo(c escrow.CollectedData) {
	go func() {
		conn, err := net.Dial("tcp", g.VideoAddress)
		if err != nil {
			log.Errorf("error opening video server connection: %s\n", err)
			return
		}
		//status, err := bufio.NewReader(conn).ReadBytes('\r')

		f, err := os.Open(c.Path)
		if err != nil {
			log.Error("Gateway: Error opening video file: ", err)
			return
		}

		fi, err := f.Stat()
		if err != nil {
			log.Error("Gateway: Error stating video file: ", err)
			return
		}

		log.WithFields(log.Fields{
			"path": c.Path,
		}).Info("Gateway sending video to central server")

		// Camera name
		fmt.Fprintf(conn, "%d\x00", c.Source)

		// Timestamp/ID
		id := make([]byte, 4)
		binary.LittleEndian.PutUint32(id, uint32(c.Timestamp.Unix()))
		conn.Write(id)

		// Data size
		size := make([]byte, 4)
		binary.LittleEndian.PutUint32(size, uint32(fi.Size()))
		conn.Write(size)

		// Send data
		n, err := io.Copy(conn, f)
		if n != fi.Size() || err != nil {
			log.WithFields(log.Fields{
				"written": n,
				"total":   fi.Size(),
				"error":   err,
			}).Error("Gateway: Error writing video data")
		}

		// TODO
		// Wait for ACK
		// Remove file

	}()
}
