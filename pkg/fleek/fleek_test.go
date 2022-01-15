package fleek

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
	fleekKey = "Rl1w0kTGQJXU+3FxgYFl0w=="
	fleekSec = "h7lkcIKxH50wVMVAvpHZGTjrkYGt3LYYLkPeRMUzSKo="
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

	fleek := &Fleek{Apikey: fleekKey, Secret: fleekSec}
	if _, err := fleek.PinFile(tmpfile.Name()); err != nil {
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
			fleek := &Fleek{Apikey: fleekKey, Secret: fleekSec}
			file := test.file.(io.Reader)
			if _, err := fleek.PinWithReader(file); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPinWithBytes(t *testing.T) {
	buf := []byte(helper.RandString(6, "lower"))
	fleek := &Fleek{Apikey: fleekKey, Secret: fleekSec}
	if _, err := fleek.PinWithBytes(buf); err != nil {
		t.Error(err)
	}
}

func TestPinHash(t *testing.T) {
	hash := "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

	fleek := &Fleek{Apikey: fleekKey, Secret: fleekSec}
	if ok, err := fleek.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
