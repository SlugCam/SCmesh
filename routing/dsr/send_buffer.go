package dsr

import (
	"container/list"

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
