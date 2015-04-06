package ipc

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/local/gateway"
	"github.com/SlugCam/SCmesh/packet/header"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/routing"
)

type PingOptions struct {
	Destination uint32
	TTL         int
}

const (
	PING_FLOOD = iota
	PING_DSR
)

func makePingPacket(localID uint32) (dh header.DataHeader, b []byte, err error) {

	dh = header.DataHeader{
		Destinations: []uint32{routing.BroadcastID},
		Type:         header.DataHeader_MESSAGE.Enum(),
	}
	pingM := &gateway.OutboundMessage{
		Id:   0,
		Cam:  fmt.Sprintf("%d", localID),
		Time: time.Now(),
		Type: "ping",
	}
	b, err = json.Marshal(pingM)
	log.Debug("pingpacket:", string(b))

	return
}

func ping(r pipeline.Router, pingOptions json.RawMessage, style int) {
	var po PingOptions
	err := json.Unmarshal(pingOptions, &po)
	if err != nil {
		log.Error("error parsing control message:", err)
	}
	log.Info("ping request received.")

	dh, d, err := makePingPacket(r.LocalID())
	if err != nil {
		log.Error("error creating ping packet:", err)
	}

	if style == PING_FLOOD {
		r.OriginateFlooding(po.TTL, dh, d)
	} else if style == PING_DSR {
		r.OriginateDSR(po.Destination, int64(0), dh, d)
	}
}
