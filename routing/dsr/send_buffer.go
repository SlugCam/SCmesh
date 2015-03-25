package dsr

import (
	"container/list"
	log "github.com/Sirupsen/logrus"

	"github.com/SlugCam/SCmesh/packet"
)

// TODO needs lots of work

type sendBuffer struct {
	l *list.List // l should only contain *packet.Packet
}

func newSendBuffer() *sendBuffer {
	b := new(sendBuffer)
	b.l = list.New()
	return b
}

func (b *sendBuffer) addPacket(p *packet.Packet) {
	b.l.PushBack(p)
}

// getSendable returns slice of packets now sendable with source route options added.
func (b *sendBuffer) getSendable(route []NodeID) []*packet.Packet {
	log.Info("getSendable route:", route)

	var sendable []*packet.Packet

	for e := b.l.Front(); e != nil; e = e.Next() {
		p, ok := e.Value.(*packet.Packet)
		if !ok || p.Header.Destination == nil {
			log.Error("Packet in send buffer was not packet or missing destination")
			b.l.Remove(e)
			continue
		}
		i := findNodeIndex(route, NodeID(*p.Header.Destination))
		// Add source route and remove from list if route found
		if i > -1 {
			b.l.Remove(e)
			err := addSourceRoute(p, route[:i])
			if err != nil {
				log.Error("getSendable:", err)
				continue
			}
			sendable = append(sendable, p)
		}
	}
	return sendable
}
