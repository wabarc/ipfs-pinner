package nftstorage

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
  "ok": false,
  "error": {
    "name": "string",
    "message": "string"
  }
}`
	badRequestJSON = `{
  "ok": false,
  "error": {
    "name": "string",
    "message": "string"
  }
}`
	uploadJSON = `{
  "ok": true,
  "value": {
    "cid": "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
    "size": 132614,
    "created": "2021-03-12T17:03:07.787Z",
    "type": "image/jpeg",
    "scope": "default",
    "pin": {
      "cid": "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
      "name": "pin name",
      "meta": {},
      "status": "queued",
      "created": "2021-03-12T17:03:07.787Z",
      "size": 132614
    },
    "files": [
      {
        "name": "logo.jpg",
        "type": "image/jpeg"
      }
    ],
    "deals": [
      {
        "batchRootCid": "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
        "lastChange": "2021-03-18T11:46:50.000Z",
        "miner": "f05678",
        "network": "nerpanet",
        "pieceCid": "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
        "status": "queued",
        "statusText": "miner rejected my data",
        "chainDealID": 138,
        "dealActivation": "2021-03-18T11:46:50.000Z",
        "dealExpiration": "2021-03-18T11:46:50.000Z"
      }
    ]
  }
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

	nft := &NFTStorage{Apikey: "fake-nft-storage-apikey", client: httpClient}
	if _, err := nft.PinFile(tmpfile.Name()); err != nil {
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

	nft := &NFTStorage{Apikey: "fake-nft-storage-apikey", client: httpClient}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := test.file.(io.Reader)
			if _, err := nft.PinWithReader(file); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPinWithBytes(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	nft := &NFTStorage{Apikey: "fake-nft-storage-apikey", client: httpClient}
	buf := []byte(helper.RandString(6, "lower"))
	if _, err := nft.PinWithBytes(buf); err != nil {
		t.Error(err)
	}
}

func TestPinHash(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	hash := "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

	nft := &NFTStorage{Apikey: "fake-nft-storage-apikey", client: httpClient}
	if ok, err := nft.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
