package escrow

import "time"

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

type Distributor struct {
	requests chan<- sendRequest
	timeouts chan<- int // send filenumber to check
	active   []File     // Active entries
}

type File struct {
	data   []byte
	bitmap []bool     // contains acknowledged bitmap
	timer  time.Timer // Timer contains the resend timeout
}

type sendRequest struct {
	reqType int    // Either MESSAGE_REQUEST or VIDEO_REQUEST
	data    []byte // JSON or path for message or video respectively
	dest    uint32 // The node to send to
}

func (d *Distributor) RegisterMessage(dest uint32, message []byte) {
	d.requests <- sendRequest{
		reqType: MESSAGE_REQUEST,
		data:    message,
		dest:    dest,
	}

}

func (d *Distributor) processSendRequest(r sendRequest) {
	// Make new file entry

	// Run send remaining on entry

}

func (d *Distributor) RegisterFile() {

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

func Distribute(incomingACKs <-chan ACK) *Distributor {
	d := new(Distributor)

	requests := make(chan sendRequest)
	d.requests = requests

	timeouts := make(chan int)
	d.timeouts = timeouts
	// load entries from storage

	go func() {
		for {
			select {
			case rr := <-requests:
				d.processSendRequest(rr)
			// Registration requests

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
