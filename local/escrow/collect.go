package escrow

import (
	"os"
	"path"
	"time"

	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
)

type CollectedData struct {
	ID        uint32
	Size      uint32
	DataType  string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// maybe use an application datatype and a routing data type

type file struct {
}

type Collector struct {
	metaPath  string
	outPath   string
	storePath string
	router    pipeline.Router
}

func Collect(pathPrefix string, incomingPackets <-chan packet.Packet, out chan<- CollectedData, router pipeline.Router) (d *Collector, err error) {
	d = new(Collector)

	d.metaPath = path.Join(pathPrefix, COL_META_PATH)
	d.outPath = path.Join(pathPrefix, COL_OUT_PATH)
	d.storePath = path.Join(pathPrefix, COL_STORE_PATH)

	d.router = router

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

	// Main loop
	go func() {
		for {
			select {
			case p := <-incomingPackets:
				// Acknowledgement received
				_ = p
			}

		}

	}()
	return
}
