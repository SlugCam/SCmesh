package local

import (
	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/local/ipc"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
)

func LocalProcessing(in <-chan packet.Packet, router pipeline.Router) {

	ipc.ListenIPC(router)

	go func() {
		for c := range in {
			log.WithFields(log.Fields{
				"packet": c,
			}).Info("Packet received locally:")
		}
	}()
}
