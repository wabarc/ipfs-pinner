package pinata

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
)

var (
	// This key is only for testing purposes.
	pinataKey = "8864aeb47a5d4b2801c6"
	pinataSec = "7f70e2a3720efbfee0905fb5b3af8994c58a4a09766bca190d5259d34b03d345"
)

func TestPinFile(t *testing.T) {
	content := []byte(helper.RandString(6, "lower"))
	tmpfile, err := ioutil.TempFile("", "ipfs-pinner-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	pinata := &Pinata{Apikey: pinataKey, Secret: pinataSec}
	if _, err := pinata.PinFile(tmpfile.Name()); err != nil {
		t.Error(err)
	}
}

func TestPinWithReader(t *testing.T) {
	content := []byte(helper.RandString(6, "lower"))
	tmpfile, err := ioutil.TempFile("", "ipfs-pinner-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		file interface{}
	}{
		{"os.File", tmpfile},
		{"strings.Reader", strings.NewReader(helper.RandString(6, "lower"))},
		{"bytes.Buffer", bytes.NewBufferString(helper.RandString(6, "lower"))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pinata := &Pinata{Apikey: pinataKey, Secret: pinataSec}
			file := test.file.(io.Reader)
			if _, err := pinata.PinWithReader(file); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPinWithBytes(t *testing.T) {
	buf := []byte(helper.RandString(6, "lower"))
	pinata := &Pinata{Apikey: pinataKey, Secret: pinataSec}
	if _, err := pinata.PinWithBytes(buf); err != nil {
		t.Error(err)
	}
}

func TestPinHash(t *testing.T) {
	hash := "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

	pinata := &Pinata{Apikey: pinataKey, Secret: pinataSec}
	if ok, err := pinata.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
