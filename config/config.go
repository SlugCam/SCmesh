package config

import (
	"io"
	"time"

	"github.com/SlugCam/SCmesh/local"
	"github.com/SlugCam/SCmesh/packet"
	"github.com/SlugCam/SCmesh/pipeline"
	"github.com/SlugCam/SCmesh/prefilter"
	"github.com/SlugCam/SCmesh/routing"
	"github.com/SlugCam/SCmesh/util"
)

// DefaultConfig returns the typical default pipeline configuration for SCmesh.
func DefaultConfig(localID uint32, serial io.ReadWriter) pipeline.Config {
	return pipeline.Config{
		LocalID:            localID,
		Serial:             serial,
		WiFlyResetInterval: 30 * time.Second,
		Prefilter:          prefilter.Prefilter,
		ParsePackets:       packet.ParsePackets,
		RoutePackets:       routing.RoutePackets,
		LocalProcessing:    local.LocalProcessing,
		PackPackets:        packet.PackPackets,
		WritePackets:       util.WritePackets,
	}
}
