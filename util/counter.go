package util

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"log"
	"os"
)

// RunCounter keeps a running counter in sync with a file. It writes a new
// counter value first, this way even in a crash we will not reuse a value, this
// means IDs will also start at 0 but that we also skip the value read at
// startup. So beware, this counter will not provide entirely contiguous values,
// but it should provide unique values.
// TODO deal with ID rollover
// TODO there is probably a better way to do this. This cannot be closed and so
// if the reference to the channel is lost this routine is wasted. Which doesn't
// matter for our purposes as much because it will stay on the whole time.
func RunCounterUint32(path string) <-chan uint32 {
	ch := make(chan uint32)
	go func() {
		current := uint32(0)
		// Attempt to read the value off disk
		d, err := ioutil.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatal("Error reading message counter in escrow")
			}
		} else {
			current = binary.LittleEndian.Uint32(d)
		}
		// At this point the ID is either the read value or 0

		// Go in to a loop putting out values
		outBuf := make([]byte, 4)
		for {
			current++
			binary.LittleEndian.PutUint32(outBuf, current)
			err = ioutil.WriteFile(path, outBuf, 0600)
			ch <- current // Block until this value is read
		}
	}()
	return ch
}

// RunCounter keeps a running counter in sync with a file. It writes a new
// counter value first, this way even in a crash we will not reuse a value, this
// means IDs will also start at 0 but that we also skip the value read at
// startup. So beware, this counter will not provide entirely contiguous values,
// but it should provide unique values.
// TODO deal with ID rollover
// TODO there is probably a better way to do this. This cannot be closed and so
// if the reference to the channel is lost this routine is wasted. Which doesn't
// matter for our purposes as much because it will stay on the whole time.
func RunCounterInt64(path string) <-chan int64 {
	ch := make(chan int64)
	go func() {
		current := int64(0)
		// Attempt to read the value off disk
		d, err := ioutil.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatal("Error reading message counter in escrow")
			}
		} else {
			b := bytes.NewBuffer(d)
			err := binary.Read(b, binary.LittleEndian, &current)
			if err != nil {
				log.Fatal("Error reading counter value from disk")
			}
		}
		// At this point the ID is either the read value or 0

		// Go in to a loop putting out values
		for {
			current++
			b := new(bytes.Buffer)
			err = binary.Write(b, binary.LittleEndian, current)
			if err != nil {
				log.Fatal("Error writing counter to disk")
			}
			err = ioutil.WriteFile(path, b.Bytes(), 0600)
			if err != nil {
				log.Fatal("Error writing counter to disk")
			}
			ch <- current // Block until this value is read
		}
	}()
	return ch
}
