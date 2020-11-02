package pinner // import "github.com/wabarc/ipfs-pinner"

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"testing"
)

func TestPinNonExistFile(t *testing.T) {
	fakeFilepath := base64.StdEncoding.EncodeToString([]byte("this is a fake filepath"))
	handle := Config{}
	if _, err := handle.Pin(fakeFilepath); err == nil {
		t.Fail()
	}
}

func TestPinFile(t *testing.T) {
	content := []byte("Hello, IPFS!")
	tmpfile, err := ioutil.TempFile("", "ipfs-pinner-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	handle := Config{Pinner: "infura"}
	if cid, err := handle.Pin(tmpfile.Name()); err != nil {
		t.Error(err)
	} else {
		t.Log(cid)
	}
}
