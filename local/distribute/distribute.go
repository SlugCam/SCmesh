package distribute

// Fix when IDs are rolled over, create will fail and drop data

// TODO how can we cleanup the file system?
// TODO prioritize resends for messages?

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
)

// These path constants will be relative to the prefix provided to the
// distribute function.
const (
	//PATH_PREFIX          = "/var/SlugCam/SCmesh"
	COUNTER_PATH = "count"
	STORE_PATH   = "store"
	OUT_PATH     = "out"
	META_PATH    = "meta"
)

const (
	MAX_PAYLOAD_SIZE    = 512
	REQUEST_BUFFER_SIZE = 100
	TIMEOUT_BUFFER_SIZE = 10
)

// TODO priorities
const (
	FIRST_TIMEOUT = 30 * time.Second // Time to wa
	ACK_TIMEOUT   = 10 * time.Second
)

// RegistrationRequest contains the data necessary to register a file to send.
type RegistrationRequest struct {
	DataType    string           `json:"type"`
	Destination uint32           `json:"destination"`
	Timestamp   time.Time        `json:"timestamp"`
	JSON        *json.RawMessage `json:"json"`
	Path        string           `json:"path"`
	Save        bool             `json:"save"`
}

// meta contains the metadata for an outbound request.
type meta struct {
	ID   uint32
	Size uint32

	DataType    string    `json:"type"`
	Destination uint32    `json:"destination"`
	Timestamp   time.Time `json:"timestamp"`
	Path        string    `json:"path"`
	Save        bool      `json:"save"`
}

type file struct {
	meta    meta
	timeout time.Time
}

type Distributor struct {
	storePath string
	outPath   string
	metaPath  string
	requests  chan<- meta
	timeouts  chan<- uint32    // send filenumber to check
	files     map[uint32]*file // A map from file entries to metadata
	messageID <-chan uint32
	router    pipeline.Router
}

// Register signals a desire to send a piece of data. It will return with the
// file ID for the new file so we can keep track of it later. This function
// should not return until the request is in a state that we can recover from in
// the event of a system failure. In other words, once this function returns,
// this request should be able to be forgotten by the caller.
func (d *Distributor) Register(r RegistrationRequest) (id uint32, err error) {
	// In order to satisfy the condition that after running this function, a
	// power failure should not cause us to fail at transmitting the data.

	// First build meta
	id = <-d.messageID

	// If request is JSON save the json and then treat this request as a regular
	// file.
	if r.JSON != nil {
		r.Path = path.Join(d.storePath, util.Utoa(id))
		r.Save = false
		// save the JSON to the store directory and update the path
		var b []byte
		b, err = json.Marshal(r.JSON)
		if err != nil {
			return
		}
		err = ioutil.WriteFile(r.Path, b, 0660)
		if err != nil {
			return
		}
	}

	if len(r.Path) == 0 {
		err = fmt.Errorf("no data included in registration request")
		return
	}

	// Encode the file, since the ID should be unique we are safe to edit these
	// files from go routines.
	reqOutPath := path.Join(d.outPath, util.Utoa(id))
	FileToWire(r.Path, reqOutPath)
	var fi os.FileInfo
	fi, err = os.Stat(reqOutPath)
	if err != nil {
		return
	}
	encodedSize := fi.Size()

	// Make new metadata entry
	m := meta{
		ID:   id,
		Size: uint32(encodedSize),

		DataType:    r.DataType,
		Destination: r.Destination,
		Timestamp:   r.Timestamp,
		Path:        r.Path,
		Save:        r.Save,
	}

	// Write metadata to file
	reqMetaPath := path.Join(d.metaPath, util.Utoa(id))
	mfile, err := os.Create(reqMetaPath)
	menc := gob.NewEncoder(mfile)
	menc.Encode(m)

	d.requests <- m
	return
}

// loadMeta is used to bring a new piece of metadata into the system. The
// register function is goroutine safe and accessing the files is ok there,
// however writing to the distributors file record is not. This function just
// adds the metadata and triggers a resend of the file.
func (d *Distributor) loadMetadata(m meta) {
	_, ok := d.files[m.ID]
	if ok {
		// Metadata already loaded
		return
	}
	d.files[m.ID] = &file{
		meta:    m,
		timeout: time.Now(),
	}
	d.timeouts <- m.ID
}

// TODO should error check come after processing
// TODO Check for errors
// Returns whether eof was encountered and an error if there is one. Note that
// in go n is initialized to 0, eof to false, and err to nil.
func scanNull(r *bufio.Reader) (n int, eof bool, err error) {
	var c byte
	for {
		c, err = r.ReadByte()
		n++
		if err != nil {
			if err == io.EOF {
				err = nil
				eof = true
				return
			}
			return
		}
		if c != byte(0) {
			err = r.UnreadByte()
			n--
			return
		}
	}
}

func scanBytes(r *bufio.Reader) (data []byte, n int, eof bool, err error) {
	var c byte
	for {
		c, err = r.ReadByte()
		n++
		if err != nil {
			if err == io.EOF {
				err = nil
				eof = true
				return
			}
			return
		}
		if c == byte(0) {
			err = r.UnreadByte()
			n--
			return
		}
		data = append(data, c)
		if n >= MAX_PAYLOAD_SIZE {
			return
		}
	}
}

// TODO better error checking
func (d *Distributor) sendRemaining(id uint32) {
	f := d.files[id]

	// TODO check if we are complete here, not at ACK
	// make packet, send it off

	dh := header.DataHeader{
		FileId:       proto.Uint32(id),
		Destinations: []uint32{f.meta.Destination},
		Type:         header.DataHeader_MESSAGE.Enum(),
		// TODO timestamp, type ->string, size
	}
	offset := 0
	dataFile, _ := os.Open(f.meta.Path)
	// TODO check err
	r := bufio.NewReader(dataFile)
	first := true
	for {
		n, eof, err := scanNull(r)
		if err != nil || eof {
			break
		}
		offset += n
		if first {
			first = false
			if uint32(n) == f.meta.Size {
				break
				// TODO file is done according to us
			}
		}
		b, n, eof, err := scanBytes(r)

		d.router.OriginateDSR(f.meta.Destination, uint32(offset), dh, b)
		if err != nil || eof {
			break
		}
		offset += n
	}

	// set timeout
	f.timeout = time.Now().Add(FIRST_TIMEOUT)
	time.AfterFunc(FIRST_TIMEOUT, func() {
		d.timeouts <- id
	})

}

func (d *Distributor) checkTimeout(id uint32) {
	f, ok := d.files[id]
	if !ok || time.Now().Before(f.timeout) {
		return
	}
	d.sendRemaining(id)
}

func (d *Distributor) receiveACK(ack ACK) {
	// If ACK denotates completed data delete the entry

	// update bitmap

	// check if fully acknowledged

	// if not reset timeout timer

	/* set timeout
	f.timeout = time.Now().Add(ACK_TIMEOUT)
	time.AfterFunc(ACK_TIMEOUT, func() {
		d.timeouts <- id
	})
	*/
}

func Distribute(pathPrefix string, router pipeline.Router, incomingACKs <-chan ACK) (d *Distributor, err error) {
	d = new(Distributor)

	requests := make(chan meta, REQUEST_BUFFER_SIZE)
	timeouts := make(chan uint32, TIMEOUT_BUFFER_SIZE)

	d.requests = requests
	d.timeouts = timeouts
	d.files = make(map[uint32]*file)

	d.metaPath = path.Join(pathPrefix, META_PATH)
	d.outPath = path.Join(pathPrefix, OUT_PATH)
	d.storePath = path.Join(pathPrefix, STORE_PATH)

	d.router = router

	counterPath := path.Join(pathPrefix, COUNTER_PATH)

	// Make the directories (this will build the path dependency for counterPath
	// as well since it just needs the path prefix, which is recursively made
	// here)
	err = os.MkdirAll(d.metaPath, 0755)
	if err != nil {
		return
	}
	err = os.MkdirAll(d.outPath, 0755)
	if err != nil {
		return
	}
	err = os.MkdirAll(d.storePath, 0755)
	if err != nil {
		return
	}

	// Start the counter
	d.messageID = util.RunCounterUint32(counterPath)

	// Main loop
	go func() {
		for {
			select {
			case rr := <-requests:
				// Registration requests
				d.loadMetadata(rr)

			case ack := <-incomingACKs:
				// Acknowledgement received
				d.receiveACK(ack)

			case fileID := <-timeouts:
				d.checkTimeout(fileID)
				// Timeout checks
			}

		}

	}()
	return
}
