package infura

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/wabarc/helper"
)

var (
	projectID     = "fake-project-id"
	projectSecret = "fake-project-secret"
	addJSON       = `{
  "Bytes": 0,
  "Hash": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a",
  "Name": "name",
  "Size": "string"
}`
	pinHashJSON = `{
  "Pins": [
    "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"
  ],
  "Progress": 0
}`
	badRequestJSON      = `{}`
	unauthorizedJSON    = `{}`
	tooManyRequestsJSON = `{}`
)

func handleResponse(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) < 10 {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(unauthorizedJSON))
		return
	}
	switch r.URL.Path {
	case "/api/v0/add":
		_, _ = w.Write([]byte(addJSON))
	case "/api/v0/pin/add":
		_, _ = w.Write([]byte(pinHashJSON))
	default:
		_, _ = w.Write([]byte(badRequestJSON))
	}
}

func TestPinFile(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	content := []byte(helper.RandString(6, "lower"))
	tmpfile, err := ioutil.TempFile("", "ipfs-pinner-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	inf := &Infura{projectID, projectSecret, httpClient}
	if _, err := inf.PinFile(tmpfile.Name()); err != nil {
		t.Error(err)
	}
}

func TestPinWithReader(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

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

	inf := &Infura{projectID, projectSecret, httpClient}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := test.file.(io.Reader)
			if _, err := inf.PinWithReader(file); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPinWithBytes(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	inf := &Infura{projectID, projectSecret, httpClient}
	buf := []byte(helper.RandString(6, "lower"))
	if _, err := inf.PinWithBytes(buf); err != nil {
		t.Error(err)
	}
}

func TestPinHash(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	hash := "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

	inf := &Infura{projectID, projectSecret, httpClient}
	if ok, err := inf.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}

func TestRateLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skip in short mode")
	}

	httpClient, mux, server := helper.MockServer()
	var retries int32
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Retry one times
		if retries < 1 {
			atomic.AddInt32(&retries, 1)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(tooManyRequestsJSON))
		} else {
			_, _ = w.Write([]byte(addJSON))
		}
	})
	defer server.Close()

	content := []byte(helper.RandString(6, "lower"))
	tmpfile, err := ioutil.TempFile("", "ipfs-pinner-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	inf := &Infura{projectID, projectSecret, httpClient}
	if _, err := inf.PinFile(tmpfile.Name()); err != nil {
		t.Error(err)
	}
}
