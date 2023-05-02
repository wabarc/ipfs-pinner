package infura

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner/file"
)

var (
	apikey  = "fake-project-id"
	secret  = "fake-project-secret"
	addJSON = `{
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
		_ = r.ParseMultipartForm(32 << 20)
		_, params, parseErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(badRequestJSON))
			return
		}

		multipartReader := multipart.NewReader(r.Body, params["boundary"])
		defer r.Body.Close()

		// Pin directory
		if len(r.MultipartForm.File) == 0 && multipartReader != nil {
			_, _ = w.Write([]byte(addJSON))
			return
		}
		// Pin file
		if len(r.MultipartForm.File["file"]) > 0 {
			_, _ = w.Write([]byte(addJSON))
			return
		}
	case "/api/v0/pin/add":
		_, _ = w.Write([]byte(pinHashJSON))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(badRequestJSON))
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

	inf := &Infura{httpClient, apikey, secret}
	o, err := inf.PinFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	_, err = cid.Parse(o)
	if err != nil {
		t.Fatalf("Invalid cid: %v", o)
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

	fr, _ := os.Open(tmpfile.Name())
	tests := []struct {
		name string
		file interface{}
	}{
		{"os.File", fr},
		{"strings.Reader", strings.NewReader(helper.RandString(6, "lower"))},
		{"bytes.Buffer", bytes.NewBufferString(helper.RandString(6, "lower"))},
	}

	inf := &Infura{httpClient, apikey, secret}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := test.file.(io.Reader)
			o, err := inf.PinWithReader(file)
			if err != nil {
				t.Fatal(err)
			}
			_, err = cid.Parse(o)
			if err != nil {
				t.Fatalf("Invalid cid: %v", o)
			}
		})
	}
}

func TestPinWithBytes(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	inf := &Infura{httpClient, apikey, secret}
	buf := []byte(helper.RandString(6, "lower"))
	o, err := inf.PinWithBytes(buf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cid.Parse(o)
	if err != nil {
		t.Fatalf("Invalid cid: %v", o)
	}
}

func TestPinHash(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	hash := "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

	inf := &Infura{httpClient, apikey, secret}
	if ok, err := inf.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}

func TestPinDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "ipfs-pinner-dir-")
	if err != nil {
		t.Fatalf("Unexpected create directory: %v", err)
	}
	defer os.RemoveAll(dir)
	subdir, err := ioutil.TempDir(dir, "ipfs-pinner-subdir-")

	// Create files
	for i := 1; i <= 2; i++ {
		f, err := ioutil.TempFile(dir, "file-")
		if err != nil {
			t.Fatal("Unexpected create file")
		}
		content := []byte(helper.RandString(6, "lower"))
		if _, err := f.Write(content); err != nil {
			t.Fatal("Unexpected write content to file")
		}
	}
	// Write file to subdirectory
	f, err := ioutil.TempFile(subdir, "file-in-subdir-")
	if err != nil {
		t.Fatal("Unexpected create file")
	}
	content := []byte(helper.RandString(6, "lower"))
	if _, err := f.Write(content); err != nil {
		t.Fatal("Unexpected write content to file")
	}

	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	body, err := file.NewMultiFileReader(dir, false)
	if err != nil {
		t.Fatalf("Unexpected creates multipart file")
	}
	inf := &Infura{httpClient, apikey, secret}
	o, err := inf.PinDir(body)
	if err != nil {
		t.Fatalf("Unexpected pin directory: %v", err)
	}
	_, err = cid.Parse(o)
	if err != nil {
		t.Fatalf("Invalid cid: %v", o)
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

	inf := &Infura{httpClient, apikey, secret}
	o, err := inf.PinFile(tmpfile.Name())
	if err != nil {
		t.Error(err)
	}
	_, err = cid.Parse(o)
	if err != nil {
		t.Fatalf("Invalid cid: %v", o)
	}
}
