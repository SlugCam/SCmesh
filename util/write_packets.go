package util

import (
	log "github.com/Sirupsen/logrus"
	"io"
	"time"
)

func resetWiFly(out io.Writer) {
	out.Write([]byte("$$$"))
	time.Sleep(300 * time.Millisecond)
	out.Write([]byte("\rreboot\r"))
	time.Sleep(3 * time.Second)
}

// WritePackets writes byte slices to an io.Writer. Used as the last stage in
// the pipeline.
func WritePackets(in <-chan []byte, out io.Writer, resetTime time.Duration) {
	log.Info("reseting WiFly")
	resetWiFly(out)

	t := time.NewTicker(resetTime)
	out.Write([]byte{'\x04'}) // Send any extraneous data

	go func() {
		if resetTime == time.Duration(0) {
			for c := range in {
				out.Write(c)
			}
		} else {

			for {
				select {
				case c := <-in:
					log.Info("sending packet")
					out.Write(c)
				case <-t.C:
					log.Info("reseting WiFly")
					resetWiFly(out)
				}
			}
		}

	}()

}
