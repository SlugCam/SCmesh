package util

import (
	"io"
	"time"
)

func resetWiFly(out io.Writer) {
	out.Write([]byte("$$$"))
	time.Sleep(260 * time.Millisecond)
	out.Write([]byte("reboot\r\n"))
	time.Sleep(3 * time.Second)
}

// WritePackets writes byte slices to an io.Writer. Used as the last stage in
// the pipeline.
func WritePackets(in <-chan []byte, out io.Writer, resetTime time.Duration) {
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
					out.Write(c)
				case <-t.C:
					resetWiFly(out)
				}
			}
		}

	}()

}
