// TODO fix the index rollover issue
// TODO deadlock in pushing to full timer, check others as well

package escrow

import (
	"os"
	"path"
	"time"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
)

// CollectedData is what is provided when a file has finished transferring to
// our node.
type CollectedData struct {
	DataType  string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}
type CollectedMeta struct {
	DataType  string
	Size      uint64
	Timestamp time.Time
}

// TODO maybe use an application datatype and a routing data type

type Collector struct {
	metaPath    string
	outPath     string
	storePath   string
	router      pipeline.Router
	out         chan<- CollectedData
	timers      map[string]*time.Timer
	scanRequest chan<- string // source.id
}

func (c *Collector) mkdirs() error {
	// Make the directories (this will build the path dependency for counterPath
	// as well since it just needs the path prefix, which is recursively made
	// here)
	err := os.MkdirAll(c.metaPath, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(c.outPath, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(c.storePath, 0755)
	if err != nil {
		return err
	}

}

func (c *Collector) processPacket(p packet.Packet) error {
	// if entry exists for this file add data and send ACK, check if
	// complete, is so, push to output
	// TODO check headers exist
	if p.Header == nil || p.Header.DataHeader == nil || p.Header.DataHeader.FileHeader == nil {
		return fmt.Errorf("packet did not include required headers to collect")
	}

	dh := p.Header.DataHeader.FileHeader

	source := util.Uint32toa(*p.Header.Source)
	fileID := util.Uint32toa(*dh.FileId)
	// TODO Sprintf?
	filePath := strings.Join([]string{source, ".", fileID}, "")

	// save metadata to file
	metaFilePath := path.Join(c.metaPath, filePath)
	outFilePath := path.Join(c.storePath, filePath)

	_, err := os.Stat(metaFilePath)
	if err != nil && os.IsNotExist(err) {
		// Save meta
		meta := CollectedMeta{
			DataType:  *dh.Type,
			Size:      *dh.FileSize,
			Timestamp: time.Unix(*dh.Timestamp, 0),
		}
		mfile, err := os.Create(metaFilePath)
		menc := gob.NewEncoder(mfile)
		menc.Encode(meta)
	}

	f, err := os.OpenFile(outFilePath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	// copy data to file
	_, err = f.WriteAt(p.Payload, p.Preheader.Offset)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	// send ACK
	ack := ACK{
		FileID: *dh.FileID,
		Offset: uint64(p.Preheader.Offset),
		Size:   len(p.Payload),
	}
	ack.send(*p.Header.Source, c.router)

	// check if file is correct length, if so set timeout to scan file
	fi, err = f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == *dh.FileSize {
		// set or reset timeout
		timer, ok := c.timers[fileID]
		if ok {
			timer.Reset(COL_FILE_SCAN_TIMEOUT)
		} else {
			c.timers[fileID] = time.AfterFunc(COL_FILE_SCAN_TIMEOUT, func() {
				c.scanRequest <- fileID
			})

		}
	}
}

func (c *Collector) scanFile(fileID string) {
}

func Collect(pathPrefix string, incomingPackets <-chan packet.Packet, out chan<- CollectedData, router pipeline.Router) (c *Collector, err error) {
	c = new(Collector)

	c.metaPath = path.Join(pathPrefix, COL_META_PATH)
	c.outPath = path.Join(pathPrefix, COL_OUT_PATH)
	c.storePath = path.Join(pathPrefix, COL_STORE_PATH)
	c.router = router
	c.out = out
	// TODO magic number
	c.scanRequest = make(chan string, 1000)

	err := c.mkdirs()
	if err != nil {
		return err
	}

	// Main loop
	go func() {
		for {
			select {
			case p := <-incomingPackets:
				c.processPacket(p)
			case r := <-c.scanRequest:
				c.scanFile(r)
			}
		}
	}()
	return
}
