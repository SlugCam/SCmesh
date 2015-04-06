package util

import (
	"encoding/binary"
	"io/ioutil"
	"log"
	"path"
	"testing"
)

func TestRunCounterUint32(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "SCmeshTests")
	if err != nil {
		t.Fatal("error creating temp directory in order to run tests.")
	}

	// Test a non existent file
	ch1 := RunCounterUint32(path.Join(tmpdir, "c1"))
	c := <-ch1
	if c != 1 {
		t.Fatal("initializing the counter did not provide the expected value.")
	}
	c = <-ch1
	if c != 2 {
		t.Fatal("Second read from the channel did not increment the value.")
	}

	ch2 := RunCounterUint32(path.Join(tmpdir, "c1"))
	c = <-ch2
	if c != 4 {
		t.Fatal("Building a new counter on the same file did not keep value.")
	}

	// Make a new counter file
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(4))
	ioutil.WriteFile(path.Join(tmpdir, "c2"), b, 0666)

	// Test new file
	ch3 := RunCounterUint32(path.Join(tmpdir, "c2"))
	c = <-ch3
	log.Print(c)
	if c != 5 {
		t.Fatal("Building a counter on a new file did not work.")
	}

}
func TestRunCounterInt64(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "SCmeshTests")
	if err != nil {
		t.Fatal("error creating temp directory in order to run tests.")
	}

	// Test a non existent file
	ch1 := RunCounterInt64(path.Join(tmpdir, "c1"))
	c := <-ch1
	if c != 1 {
		t.Fatal("initializing the counter did not provide the expected value.")
	}
	c = <-ch1
	if c != 2 {
		t.Fatal("Second read from the channel did not increment the value.")
	}
	ch2 := RunCounterInt64(path.Join(tmpdir, "c1"))
	c = <-ch2
	if c != 4 {
		t.Fatalf("Building a new counter on the same file did not keep value. Got %d, wanted %d.", c, 4)
	}

}
