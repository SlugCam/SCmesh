package local

import (
	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/local/escrow"
	"github.com/SlugCam/SCmesh/local/ipc"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
)

const DATA_PATH = "/var/SlugCam/SCmesh"

func LocalProcessingTrackCollected(in <-chan packet.Packet, router pipeline.Router, collectedOut chan<- escrow.CollectedData) {
	// TODO ensure that this can't block other processes
	// TODO magic number
	ipcOutChan := make(chan escrow.CollectedData, 500)
	collectedData := make(chan escrow.CollectedData)

	// Non blocking transfer of collected data to ipc out channel to avoid
	// hanging local processing of packets.
	go func() {
		for c := range collectedData {
			log.WithFields(log.Fields{
				"collected_data": c,
			}).Info("finished collecting data")

			if collectedOut != nil {
				select {
				case collectedOut <- c:
				default:
				}
			}

			select {
			case ipcOutChan <- c:
			default:
			}
		}
	}()

	packetsToCollect := make(chan packet.Packet)
	packetsToDistribute := make(chan packet.Packet)

	_, err := escrow.Collect(DATA_PATH, packetsToCollect, collectedData, router)
	if err != nil {
		log.Fatal("Local process initialization error: ", err)
	}
	d, err := escrow.Distribute(DATA_PATH, packetsToDistribute, router)
	if err != nil {
		log.Fatal("Local process initialization error: ", err)
	}

	ipc.ListenIPC(router, ipcOutChan, d)

	go func() {
		for p := range in {
			log.WithFields(log.Fields{
				"packet": p.Abbreviate(),
			}).Info("Packet received locally:")

			if p.Header == nil || p.Header.DataHeader == nil {
				continue
			}

			switch *p.Header.DataHeader.Type {
			case header.DataHeader_FILE:
				packetsToCollect <- p
			case header.DataHeader_ACK:
				packetsToDistribute <- p
			case header.DataHeader_MESSAGE:
			}
		}
	}()
}

func LocalProcessing(in <-chan packet.Packet, router pipeline.Router) {
	LocalProcessingTrackCollected(in, router, nil)
}
