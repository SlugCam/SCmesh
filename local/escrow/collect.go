// TODO fix the index rollover issue
// TODO deadlock in pushing to full timer, check others as well

package escrow

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
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
	Size      int64
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
	return nil
}

func (c *Collector) processPacket(p packet.Packet) error {
	// if entry exists for this file add data and send ACK, check if
	// complete, is so, push to output
	// TODO check headers exist
	if p.Header == nil || p.Header.DataHeader == nil || p.Header.DataHeader.FileHeader == nil {
		return fmt.Errorf("packet did not include required headers to collect")
	}

	dh := p.Header.DataHeader.FileHeader

	//source := util.Utoa(*p.Header.Source)
	//fileID := util.Utoa(*dh.FileId)
	// TODO Sprintf? Close files
	//filePath := strings.Join([]string{source, ".", fileID}, "")
	filePath := fmt.Sprintf("%d.%d", *p.Header.Source, *dh.FileId)
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
		if err != nil {
			return err
		}
		menc := gob.NewEncoder(mfile)
		menc.Encode(meta)
	}

	f, err := os.OpenFile(outFilePath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	// copy data to file
	log.Info("PAYLOAD:", p.Payload)
	_, err = f.WriteAt(p.Payload, p.Preheader.PayloadOffset)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	// send ACK
	ack := ACK{
		FileID: *dh.FileId,
		Offset: p.Preheader.PayloadOffset,
		Size:   len(p.Payload),
	}
	ack.send(*p.Header.Source, c.router)

	// check if file is correct length, if so set timeout to scan file
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == *dh.FileSize {
		// set or reset timeout
		timer, ok := c.timers[filePath]
		if ok {
			timer.Reset(COL_FILE_SCAN_TIMEOUT)
		} else {
			c.timers[filePath] = time.AfterFunc(COL_FILE_SCAN_TIMEOUT, func() {
				c.scanRequest <- filePath
			})

		}
	}
	return nil
}

// TODO check file length
func (c *Collector) scanFile(fileID string) (bool, error) {
	noNull := true
	file, err := os.Open(path.Join(c.storePath, fileID))
	if err != nil {
		return noNull, err
	}
	b := make([]byte, 4096) // TODO magic number
	for {
		n, err := file.Read(b)
		for i := 0; i < n; i++ {
			if b[i] == '\x00' {
				noNull = false
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return noNull, err
			}
		}

	}
	return noNull, nil
}

func Collect(pathPrefix string, incomingPackets <-chan packet.Packet, out chan<- CollectedData, router pipeline.Router) (c *Collector, err error) {
	c = new(Collector)

	c.metaPath = path.Join(pathPrefix, COL_META_PATH)
	c.outPath = path.Join(pathPrefix, COL_OUT_PATH)
	c.storePath = path.Join(pathPrefix, COL_STORE_PATH)
	c.router = router
	c.out = out
	// TODO magic number
	c.timers = make(map[string]*time.Timer)
	scanRequest := make(chan string, 1000)
	c.scanRequest = scanRequest

	err = c.mkdirs()
	if err != nil {
		return nil, err
	}

	// Main loop
	go func() {
		for {
			select {
			case p := <-incomingPackets:
				log.WithFields(log.Fields{
					"packet": p,
				}).Info("Packet read by collector")
				err := c.processPacket(p)
				if err != nil {
					log.Error("Error processing packet: ", err)
				}
			case r := <-scanRequest:
				log.Infof("checking for completion of %d", r)
				finished, err := c.scanFile(r)
				log.Infof("log.Infof = ", finished, ", ", err)
				if err != nil {
					log.Error("Error scanning collected file. ", err)
				} else if finished {
					// TODO, check if file exists
					FileFromWire(path.Join(c.storePath, r), path.Join(c.outPath, r))
					// TODO delete old one
				}
			}
		}
	}()
	return
}
