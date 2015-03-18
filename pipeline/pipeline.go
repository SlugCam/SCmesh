package pipeline

import (
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/packet/header"
)

type Router interface {
	OriginateDSR(dest uint32, dataHeader header.DataHeader, data []byte)
	OriginateFlooding(TTL int, dataHeader header.DataHeader, data []byte)
}

type Config struct {
	LocalID         uint32
	Serial          io.ReadWriter
	Prefilter       func(in io.Reader, out chan<- packet.RawPacket)
	ParsePackets    func(in <-chan packet.RawPacket, out chan<- packet.Packet)
	RoutePackets    func(localID uint32, toForward <-chan packet.Packet, destLocal chan<- packet.Packet, out chan<- packet.Packet) Router
	LocalProcessing func(destLocal <-chan packet.Packet, router Router)
	PackPackets     func(in <-chan packet.Packet, out chan<- []byte)
	WritePackets    func(in <-chan []byte, out io.Writer)
}

func Start(c Config) {
	log.Info("Starting SCmesh")

	// Make channels
	rawPackets := make(chan packet.RawPacket)
	toRouter := make(chan packet.Packet)
	destLocal := make(chan packet.Packet)
	fromRouter := make(chan packet.Packet)
	packedPackets := make(chan []byte)

	// Setup pipeline
	c.Prefilter(c.Serial, rawPackets)

	c.ParsePackets(rawPackets, toRouter)

	r := c.RoutePackets(c.LocalID, toRouter, destLocal, fromRouter)

	c.LocalProcessing(destLocal, r)

	c.PackPackets(fromRouter, packedPackets)

	c.WritePackets(packedPackets, c.Serial)
}
