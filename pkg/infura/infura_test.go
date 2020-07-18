package infura

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestPinFile(t *testing.T) {
	content := []byte("Hello, Infura!")
	tmpfile, err := ioutil.TempFile("", "ipfs-pinner-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	if cid, err := PinFile(tmpfile.Name()); err != nil {
		t.Error(err)
	} else {
		t.Log(cid)
	}
}

func TestPinHash(t *testing.T) {
	hash := "QmdKfpnTxbfzQL9Lyw3CMXwioVBScEb887Q4L6d9Q84bVw"

	if ok, err := PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
