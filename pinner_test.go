package pinner // import "github.com/wabarc/ipfs-pinner"

import (
	"bytes"
	"encoding/base64"
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

func TestPinNonExistFile(t *testing.T) {
	fakeFilepath := base64.StdEncoding.EncodeToString([]byte("this is a fake filepath"))
	handle := Config{}
	if _, err := handle.Pin(fakeFilepath); err == nil {
		t.Fail()
	}
}

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

	tests := []struct {
		pinner string
		apikey string
		secret string
		source string
		file   interface{}
	}{
		{"infura", "", "", "os.File", tmpfile},
		{"infura", "", "", "strings.Reader", strings.NewReader(helper.RandString(6, "lower"))},
		{"infura", "", "", "bytes.Buffer", bytes.NewBufferString(helper.RandString(6, "lower"))},
		{"pinata", pinataKey, pinataSec, "os.File", tmpfile},
		{"pinata", pinataKey, pinataSec, "strings.Reader", strings.NewReader(helper.RandString(6, "lower"))},
		{"pinata", pinataKey, pinataSec, "bytes.Buffer", bytes.NewBufferString(helper.RandString(6, "lower"))},
	}

	for _, test := range tests {
		name := test.pinner + "-" + test.source
		t.Run(name, func(t *testing.T) {
			file := test.file.(io.Reader)
			pinner := Config{Pinner: test.pinner, Apikey: test.apikey, Secret: test.secret}
			if _, err := pinner.Pin(file); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPinFileWithBytes(t *testing.T) {
	tests := []struct {
		pinner string
		apikey string
		secret string
		source string
		file   interface{}
	}{
		{"infura", "", "", "bytes", []byte(helper.RandString(6, "lower"))},
		{"pinata", pinataKey, pinataSec, "bytes", []byte(helper.RandString(6, "lower"))},
	}

	for _, test := range tests {
		name := test.pinner + "-" + test.source
		t.Run(name, func(t *testing.T) {
			file := test.file.([]byte)
			pinner := Config{Pinner: test.pinner, Apikey: test.apikey, Secret: test.secret}
			if _, err := pinner.Pin(file); err != nil {
				t.Error(err)
			}
		})
	}
}
