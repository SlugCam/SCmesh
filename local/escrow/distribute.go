package escrow

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

	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/util"
	"github.com/golang/protobuf/proto"
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
	FileID int64
	Size   int64

	DataType    string    `json:"type"`
	Destination uint32    `json:"destination"`
	Timestamp   time.Time `json:"timestamp"`
	Path        string    `json:"path"`
	Save        bool      `json:"save"`
}

type outgoingFile struct {
	meta            meta
	timeout         time.Time
	scanPlaceholder int64
}

type Distributor struct {
	storePath string
	outPath   string
	metaPath  string
	requests  chan<- meta
	timeouts  chan<- int64            // send filenumber to check
	files     map[int64]*outgoingFile // A map from file entries to metadata
	messageID <-chan int64
	router    pipeline.Router
}

// Register signals a desire to send a piece of data. It will return with the
// file ID for the new file so we can keep track of it later. This function
// should not return until the request is in a state that we can recover from in
// the event of a system failure. In other words, once this function returns,
// this request should be able to be forgotten by the caller.
func (d *Distributor) Register(r RegistrationRequest) (fileID int64, err error) {
	// In order to satisfy the condition that after running this function, a
	// power failure should not cause us to fail at transmitting the data.

	// First build meta
	fileID = <-d.messageID

	// If request is JSON save the json and then treat this request as a regular
	// file.
	if r.JSON != nil {
		//r.Path = path.Join(d.storePath, util.Utoa(fileID))
		r.Path = path.Join(d.storePath, fmt.Sprintf("%d", fileID))
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
	reqOutPath := path.Join(d.outPath, fmt.Sprintf("%d", fileID))
	FileToWire(r.Path, reqOutPath)
	var fi os.FileInfo
	fi, err = os.Stat(reqOutPath)
	if err != nil {
		return
	}
	encodedSize := fi.Size()

	// Make new metadata entry
	m := meta{
		FileID: fileID,
		Size:   encodedSize,

		DataType:    r.DataType,
		Destination: r.Destination,
		Timestamp:   r.Timestamp,
		Path:        r.Path,
		Save:        r.Save,
	}

	// Write metadata to file
	reqMetaPath := path.Join(d.metaPath, fmt.Sprintf("%d", fileID))
	mfile, err := os.Create(reqMetaPath)
	if err != nil {
		log.Error("Register: ", err)
		return
	}
	defer mfile.Close()
	menc := gob.NewEncoder(mfile)

	err = menc.Encode(m)
	if err != nil {
		log.Error("Register: ", err)
		return
	}

	d.requests <- m
	return
}

// loadMeta is used to bring a new piece of metadata into the system. The
// register function is goroutine safe and accessing the files is ok there,
// however writing to the distributors file record is not. This function just
// adds the metadata and triggers a resend of the file.
func (d *Distributor) loadMetadata(m meta) {
	_, ok := d.files[m.FileID]
	if ok {
		// Metadata already loaded
		return
	}
	d.files[m.FileID] = &outgoingFile{
		meta:            m,
		timeout:         time.Now(),
		scanPlaceholder: int64(0),
	}
	d.timeouts <- m.FileID
}

// TODO should error check come after processing
// TODO Check for errors
// Returns whether eof was encountered and an error if there is one. Note that
// in go n is initialized to 0, eof to false, and err to nil.
func scanNull(r *bufio.Reader) (n int64, eof bool, err error) {
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

func scanBytes(r *bufio.Reader) (data []byte, n int64, eof bool, err error) {
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
func (d *Distributor) sendChunk(fileID int64) {
	f := d.files[fileID]

	// TODO check if we are complete here, not at ACK
	// make packet, send it off

	dh := header.DataHeader{
		Destinations: []uint32{f.meta.Destination},
		Type:         header.DataHeader_FILE.Enum(),
		FileHeader: &header.FileHeader{
			FileId:    proto.Int64(fileID),
			FileSize:  proto.Int64(f.meta.Size),
			Type:      proto.String(f.meta.DataType),
			Timestamp: proto.Int64(f.meta.Timestamp.Unix()),
		},
	}
	offset := f.scanPlaceholder
	outpath := path.Join(d.outPath, fmt.Sprintf("%d", fileID))

	dataFile, err := os.Open(outpath)
	if err != nil {
		log.Error("sendChunk: ", err)
		return
	}
	defer dataFile.Close()
	offset, err = dataFile.Seek(offset, 0)
	if err != nil {
		log.Error("sendChunk: ", err)
		return
	}

	r := bufio.NewReader(dataFile)

	n, eof, err := scanNull(r)
	if eof {
		f.scanPlaceholder = 0
		if offset == 0 {
			// File done!
		}
		return
	}
	if err != nil {
		log.Error("sendChunk: ", err)
		return
	}
	offset += n

	b, n, eof, err := scanBytes(r)

	d.router.OriginateDSR(f.meta.Destination, offset, dh, b)

	if err != nil {
		log.Error("sendChunk: ", err)
		return
	}
	if eof {
		offset = 0
	} else {
		offset += n
	}
	f.scanPlaceholder = offset

	//time.Sleep(DISTRIBUTE_RELEASE_REST)

	// set timeout
	f.timeout = time.Now().Add(DISTRIBUTE_RELEASE_REST)
	time.AfterFunc(DISTRIBUTE_RELEASE_REST, func() {
		d.timeouts <- fileID
	})

}

// TODO better error checking
func (d *Distributor) sendRemaining(fileID int64) {
	f := d.files[fileID]

	// TODO check if we are complete here, not at ACK
	// make packet, send it off

	dh := header.DataHeader{
		Destinations: []uint32{f.meta.Destination},
		Type:         header.DataHeader_FILE.Enum(),
		FileHeader: &header.FileHeader{
			FileId:    proto.Int64(fileID),
			FileSize:  proto.Int64(f.meta.Size),
			Type:      proto.String(f.meta.DataType),
			Timestamp: proto.Int64(f.meta.Timestamp.Unix()),
		},
	}
	offset := int64(0)
	outpath := path.Join(d.outPath, fmt.Sprintf("%d", fileID))

	dataFile, err := os.Open(outpath)
	if err != nil {
		log.Error("sendRemaining: ", err)
		return
	}
	defer dataFile.Close()

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
			if int64(n) == f.meta.Size {
				break
				// TODO file is done according to us
			}
		}
		b, n, eof, err := scanBytes(r)

		d.router.OriginateDSR(f.meta.Destination, offset, dh, b)

		if err != nil || eof {
			break
		}
		offset += n
		time.Sleep(DISTRIBUTE_RELEASE_REST)
	}

	// set timeout
	f.timeout = time.Now().Add(FIRST_TIMEOUT)
	time.AfterFunc(FIRST_TIMEOUT, func() {
		d.timeouts <- fileID
	})

}

func (d *Distributor) checkTimeout(fileID int64) {
	f, ok := d.files[fileID]
	if !ok || time.Now().Before(f.timeout) {
		return
	}
	//d.sendRemaining(fileID)
	d.sendChunk(fileID)
}

func (d *Distributor) receiveACK(p packet.Packet) {
	ack, err := parseACK(p)
	if err != nil {
		log.Error("receiveACK: ", err)
		return
	}
	// If ACK denotates completed data delete the entry

	// update bitmap
	outPath := path.Join(d.outPath, fmt.Sprintf("%d", ack.FileID))
	f, err := os.OpenFile(outPath, os.O_WRONLY, 0660)
	if err != nil {
		log.Error("receiveACK: ", err)
		return
	}
	defer f.Close()
	// copy data to file
	o, err := f.Seek(ack.Offset, 0)
	if err != nil || o != ack.Offset {
		log.Error("receiveACK: Could not seek to proper offset in file")
		return
	}
	for i := 0; i < ack.Size; i++ {
		// TODO check n
		_, err := f.Write([]byte{0})
		if err != nil {
			log.Error("receiveACK: Could not write null byte to file: ", err)
			return
		}
	}
	// check if fully acknowledged

	// if not reset timeout timer

	// TODO check if file exists
	d.files[ack.FileID].timeout = time.Now().Add(ACK_TIMEOUT)
	time.AfterFunc(ACK_TIMEOUT, func() {
		d.timeouts <- ack.FileID
	})
}

func Distribute(pathPrefix string, incomingPackets <-chan packet.Packet, router pipeline.Router) (d *Distributor, err error) {
	d = new(Distributor)

	requests := make(chan meta, REQUEST_BUFFER_SIZE)
	timeouts := make(chan int64, TIMEOUT_BUFFER_SIZE)

	d.requests = requests
	d.timeouts = timeouts
	d.files = make(map[int64]*outgoingFile)

	d.metaPath = path.Join(pathPrefix, DIST_META_PATH)
	d.outPath = path.Join(pathPrefix, DIST_OUT_PATH)
	d.storePath = path.Join(pathPrefix, DIST_STORE_PATH)

	d.router = router

	counterPath := path.Join(pathPrefix, DIST_COUNTER_PATH)

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
	d.messageID = util.RunCounterInt64(counterPath)

	// Main loop
	go func() {
		for {
			select {
			case rr := <-requests:
				// Registration requests
				d.loadMetadata(rr)

			case p := <-incomingPackets:
				// Acknowledgement received
				d.receiveACK(p)

			case fileID := <-timeouts:
				d.checkTimeout(fileID)
				// Timeout checks
			}

		}

	}()
	return
}
