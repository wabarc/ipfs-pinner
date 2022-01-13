package web3storage

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

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
		_, _ = w.Write([]byte(uploadJSON))
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

	tests := []struct {
		name string
		file interface{}
	}{
		{"os.File", tmpfile},
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

func TestPinHash(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	hash := "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

	web3 := &Web3Storage{Apikey: "fake-web3-storage-apikey", client: httpClient}
	if ok, err := web3.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
