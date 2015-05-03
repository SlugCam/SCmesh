package pipeline

import (
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

const BUFFER_SIZE = 1000

type Router interface {
	LocalID() uint32
	OriginateDSR(dest uint32, offset int64, dataHeader header.DataHeader, data []byte)
	OriginateFlooding(TTL int, dataHeader header.DataHeader, data []byte)
}

type Config struct {
	LocalID            uint32
	Serial             io.ReadWriter
	WiFlyResetInterval time.Duration
	Prefilter          func(in io.Reader, out chan<- packet.RawPacket)
	ParsePackets       func(in <-chan packet.RawPacket, out chan<- packet.Packet)
	RoutePackets       func(localID uint32, toForward <-chan packet.Packet, destLocal chan<- packet.Packet, out chan<- packet.Packet) Router
	LocalProcessing    func(destLocal <-chan packet.Packet, router Router)
	PackPackets        func(in <-chan packet.Packet, out chan<- []byte)
	WritePackets       func(in <-chan []byte, out io.Writer, WiFlyResetInterval time.Duration)
}

func Start(c Config) {
	log.Info("Starting SCmesh")

	// Make channels
	rawPackets := make(chan packet.RawPacket, BUFFER_SIZE)
	toRouter := make(chan packet.Packet, BUFFER_SIZE)
	destLocal := make(chan packet.Packet, BUFFER_SIZE)
	fromRouter := make(chan packet.Packet, BUFFER_SIZE)
	packedPackets := make(chan []byte, BUFFER_SIZE)

	// Setup pipeline
	c.Prefilter(c.Serial, rawPackets)

	c.ParsePackets(rawPackets, toRouter)

	r := c.RoutePackets(c.LocalID, toRouter, destLocal, fromRouter)

	c.LocalProcessing(destLocal, r)

	c.PackPackets(fromRouter, packedPackets)

	c.WritePackets(packedPackets, c.Serial, c.WiFlyResetInterval)
}
