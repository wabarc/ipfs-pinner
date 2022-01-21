package pinata

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
	pinataKey   = "fake-project-id"
	pinataSec   = "fake-project-secret"
	pinHashJSON = `{
    "hashToPin": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"
}`
	pinFileJSON = `{
    "IpfsHash": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a",
    "PinSize": 1234,
    "Timestamp": "1979-01-01 00:00:00Z"
}`
	badRequestJSON      = `{}`
	unauthorizedJSON    = `{}`
)

func handleResponse(w http.ResponseWriter, r *http.Request) {
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
		if len(r.MultipartForm.File) == 0 && multipartReader != nil {
			_, _ = w.Write([]byte(pinFileJSON))
			return
		}
		// Pin file
		if len(r.MultipartForm.File["file"]) > 0 {
			_, _ = w.Write([]byte(pinFileJSON))
			return
		}
	case "/pinning/pinByHash":
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

	pinata := &Pinata{pinataKey, pinataSec, httpClient}
	o, err := pinata.PinFile(tmpfile.Name())
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pinata := &Pinata{pinataKey, pinataSec, httpClient}
			file := test.file.(io.Reader)
			o, err := pinata.PinWithReader(file)
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

	buf := []byte(helper.RandString(6, "lower"))
	pinata := &Pinata{pinataKey, pinataSec, httpClient}
	o, err := pinata.PinWithBytes(buf)
	if err != nil {
		t.Errorf("Unexpected pin directory: %v", err)
	}
	_, err = cid.Parse(o)
	if err != nil {
		t.Fatalf("Invalid cid: %v", o)
	}
}

func TestPinDir(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	dir, err := ioutil.TempDir("", "ipfs-pinner-dir-")
	if err != nil {
		t.Fatalf("Unexpected create directory: %v", err)
	}
	defer os.RemoveAll(dir)

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

	pinata := &Pinata{pinataKey, pinataSec, httpClient}
	o, err := pinata.PinDir(dir)
	if err != nil {
		t.Fatalf("Unexpected pin directory: %v", err)
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

	pinata := &Pinata{pinataKey, pinataSec, httpClient}
	if ok, err := pinata.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
