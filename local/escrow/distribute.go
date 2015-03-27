package escrow

import (
	"path"
	"time"

	"github.com/SlugCam/SCmesh/util"
)

const (
	//PATH_PREFIX          = "/var/SlugCam/SCmesh"
	MESSAGE_COUNTER_PATH = "message_counter"
)

const (
	REQUEST_BUFFER_SIZE = 100
	TIMEOUT_BUFFER_SIZE = 10
)

// TODO priorities
const (
	FIRST_TIMEOUT = 30 * time.Second // Time to wa
	ACK_TIMEOUT   = 10 * time.Second
)

const (
	MESSAGE_REQUEST = iota
	VIDEO_REQUEST
)

type request interface {
	processRequest(d *Distributor)
}

type messageRequest struct {
	fileID uint32
	data   []byte
	dest   uint32
}

func (r messageRequest) processRequest(d *Distributor) {
	// Make new file entry

	// Write message to message file

	// Encode file to a new file

	// Run send remaining on entry

}

/*
type videoRequest struct {
	path string
	dest uint32
}

func (r videoRequest) processRequest(d *Distributor) {
	// Make new file entry

	// Run send remaining on entry

}
*/

type Distributor struct {
	requests  chan<- request
	timeouts  chan<- int // send filenumber to check
	active    []file     // Active entries
	messageID <-chan uint32
}

type file struct {
	fileType int
	data     []byte
	bitmap   []bool     // contains acknowledged bitmap
	timer    time.Timer // Timer contains the resend timeout
}

// RegisterMessage signals a desire to send a given message. It will return with
// the file ID for the new message so we can keep track of it later.
func (d *Distributor) RegisterMessage(dest uint32, message []byte) uint32 {
	id := <-d.messageID
	d.requests <- messageRequest{
		data:   message,
		dest:   dest,
		fileID: id,
	}
	return id
}

// RegisterVideo signals a desire to send the video given by the path. It is
// assumed that the videos name will be a unix timestamp as has already been
// done in the SlugCam system.
// TODO deletion
func (d *Distributor) RegisterVideo(dest uint32, path string) {

}

func sendRemaining() {
	// make packet, send it off

	// set timeout
}

func (d *Distributor) timeoutCheck(fileID int) {
	// send remaining

}

func (d *Distributor) receiveACK(ack ACK) {
	// update bitmap

	// check if fully acknowledged

	// if not reset timout timer
}

func Distribute(pathPrefix string, incomingACKs <-chan ACK) *Distributor {
	d := new(Distributor)

	requests := make(chan request, REQUEST_BUFFER_SIZE)
	d.requests = requests

	timeouts := make(chan int, TIMEOUT_BUFFER_SIZE)
	d.timeouts = timeouts
	// load entries from storage

	d.messageID = util.RunCounterUint32(path.Join(pathPrefix, MESSAGE_COUNTER_PATH))

	go func() {
		for {
			select {
			case rr := <-requests:
				// Registration requests
				rr.processRequest(d)

			case ack := <-incomingACKs:
				// Acknowledgement received
				d.receiveACK(ack)

			case fileID := <-timeouts:
				d.timeoutCheck(fileID)
				// Timeout checks
			}

		}

	}()
	return d
}
