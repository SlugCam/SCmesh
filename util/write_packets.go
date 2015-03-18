package util

import "io"

// WritePackets writes byte slices to an io.Writer. Used as the last stage in
// the pipeline.
func WritePackets(in <-chan []byte, out io.Writer) {
	out.Write([]byte{'\x04'}) // Send any extraneous data
	go func() {
		for c := range in {
			out.Write(c)
		}
	}()
}
