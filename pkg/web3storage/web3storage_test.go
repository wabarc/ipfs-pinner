package web3storage

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/wabarc/helper"
)

var (
	unauthorizedJSON = `{
  "name": "HTTP Error",
  "message": "Unauthorized"
}`
	badRequestJSON = `{
  "name": "string",
  "message": "string"
}`
	uploadJSON = `{
  "cid": "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u"
}`
)

func handleResponse(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) < 10 {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(unauthorizedJSON))
		return
	}
	switch r.URL.Path {
	case "/upload":
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
		if multipartReader != nil && len(r.MultipartForm.File["file"]) > 1 {
			_, _ = w.Write([]byte(uploadJSON))
			return
		}
		// Pin file
		if multipartReader != nil && len(r.MultipartForm.File["file"]) == 1 {
			_, _ = w.Write([]byte(uploadJSON))
			return
		}
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

	web3 := &Web3Storage{Apikey: "fake-web3-storage-apikey", client: httpClient}
	if _, err := web3.PinFile(tmpfile.Name()); err != nil {
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

	fr, _ := os.Open(tmpfile.Name())
	tests := []struct {
		name string
		file interface{}
	}{
		{"os.File", fr},
		{"strings.Reader", strings.NewReader(helper.RandString(6, "lower"))},
		{"bytes.Buffer", bytes.NewBufferString(helper.RandString(6, "lower"))},
	}

	web3 := &Web3Storage{Apikey: "fake-web3-storage-apikey", client: httpClient}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := test.file.(io.Reader)
			if _, err := web3.PinWithReader(file); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPinWithBytes(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	web3 := &Web3Storage{Apikey: "fake-web3-storage-apikey", client: httpClient}
	buf := []byte(helper.RandString(6, "lower"))
	if _, err := web3.PinWithBytes(buf); err != nil {
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

	web3 := &Web3Storage{Apikey: "fake-web3-storage-apikey", client: httpClient}
	o, err := web3.PinDir(dir)
	if err != nil {
		t.Fatalf("Unexpected pin directory: %v", err)
	}
	_, err = cid.Parse(o)
	if err != nil {
		t.Fatalf("Invalid cid: %v", o)
	}
}

func TestPinHash(t *testing.T) {
	t.Skip("Not yet supported")

	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	hash := "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

	web3 := &Web3Storage{Apikey: "fake-web3-storage-apikey", client: httpClient}
	if ok, err := web3.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
