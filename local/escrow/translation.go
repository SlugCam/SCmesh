package escrow

import (
	"encoding/ascii85"
	"io"
	"os"
)

func FileFromWire(inPath string, outPath string) {

}

func FileToWire(inPath string, outPath string) error {
	in, err := os.Open(inPath)
	if err != nil {
		return err
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}

	enc := ascii85.NewEncoder(out)
	// TODO _ should be n and checked?
	_, err = io.Copy(enc, in) // n is bytes copied from in
	if err != nil {
		return err
	}

	enc.Close()
	return nil
}
