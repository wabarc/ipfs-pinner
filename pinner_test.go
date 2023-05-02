package pinner // import "github.com/wabarc/ipfs-pinner"

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wabarc/helper"
)

var (
	apikey           = "8864aeb47a5d4b2801c6"
	secret           = "7f70e2a3720efbfee0905fb5b3af8994c58a4a09766bca190d5259d34b03d345"
	badRequestJSON   = `{}`
	unauthorizedJSON = `{}`
	addJSON          = `{
  "Bytes": 0,
  "Hash": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a",
  "Name": "name",
  "Size": "string"
}`
	pinHashJSON = `{
    "hashToPin": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"
}`
	pinFileJSON = `{
    "IpfsHash": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a",
    "PinSize": 1234,
    "Timestamp": "1979-01-01 00:00:00Z"
}`
)

func handleResponse(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Hostname() {
	case "ipfs.infura.io":
		pinHashJSON = `{
  "Pins": [
    "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"
  ],
  "Progress": 0
}`
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
	case "api.pinata.cloud":
		authorization := r.Header.Get("Authorization")
		apiKey := r.Header.Get("pinata_api_key")
		apiSec := r.Header.Get("pinata_secret_api_key")
		switch {
		case apiKey != "" && apiSec != "":
			// access
		case authorization != "" && !strings.HasPrefix(authorization, "Bearer"):
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(unauthorizedJSON))
			return
		default:
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(unauthorizedJSON))
			return
		}

		switch r.URL.Path {
		case "/pinning/pinFileToIPFS":
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
				_, _ = w.Write([]byte(pinFileJSON))
				return
			}
			// Pin file
			if multipartReader != nil && len(r.MultipartForm.File["file"]) == 1 {
				_, _ = w.Write([]byte(pinFileJSON))
				return
			}
		case "/pinning/pinByHash":
			_, _ = w.Write([]byte(pinHashJSON))
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(badRequestJSON))
	}
}

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

	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	fr, _ := os.Open(tmpfile.Name())
	tests := []struct {
		pinner string
		apikey string
		secret string
		source string
		file   interface{}
	}{
		{"infura", apikey, secret, "os.File", fr},
		{"infura", apikey, secret, "strings.Reader", strings.NewReader(helper.RandString(6, "lower"))},
		{"infura", apikey, secret, "bytes.Buffer", bytes.NewBufferString(helper.RandString(6, "lower"))},
		{"pinata", apikey, secret, "os.File", tmpfile},
		{"pinata", apikey, secret, "strings.Reader", strings.NewReader(helper.RandString(6, "lower"))},
		{"pinata", apikey, secret, "bytes.Buffer", bytes.NewBufferString(helper.RandString(6, "lower"))},
	}

	for _, test := range tests {
		name := test.pinner + "-" + test.source
		t.Run(name, func(t *testing.T) {
			file := test.file.(io.Reader)
			pinner := Config{Pinner: test.pinner, Apikey: test.apikey, Secret: test.secret}
			if _, err := pinner.WithClient(httpClient).Pin(file); err != nil {
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
		{"infura", apikey, secret, "bytes", []byte(helper.RandString(6, "lower"))},
		{"pinata", apikey, secret, "bytes", []byte(helper.RandString(6, "lower"))},
	}

	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	for _, test := range tests {
		name := test.pinner + "-" + test.source
		t.Run(name, func(t *testing.T) {
			file := test.file.([]byte)
			pinner := Config{Pinner: test.pinner, Apikey: test.apikey, Secret: test.secret}
			if _, err := pinner.WithClient(httpClient).Pin(file); err != nil {
				t.Error(err)
			}
		})
	}
}
