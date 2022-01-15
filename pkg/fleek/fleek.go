package fleek

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wabarc/helper"
	httpretry "github.com/wabarc/ipfs-pinner/http"
)

const api = "https://storageapi.fleek.co"

// Fleek represents a Fleek configuration.
type Fleek struct {
	Apikey string
	Secret string
}

// PinFile pins content to Fleek by providing a file path, it returns an IPFS
// hash and an error.
func (f *Fleek) PinFile(fp string) (string, error) {
	file, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer file.Close()

	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer m.Close()

		part, err := m.CreateFormFile("file", filepath.Base(file.Name()))
		if err != nil {
			return
		}

		if _, err = io.Copy(part, file); err != nil {
			return
		}
	}()

	return f.pinFile(r, m)
}

// PinWithReader pins content to Fleek by given io.Reader, it returns an IPFS hash and an error.
func (f *Fleek) PinWithReader(rd io.Reader) (string, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	fn := helper.RandString(6, "lower")

	go func() {
		defer w.Close()
		defer m.Close()

		part, err := m.CreateFormFile("file", fn)
		if err != nil {
			return
		}

		if _, err = io.Copy(part, rd); err != nil {
			return
		}
	}()

	return f.pinFile(r, m)
}

// PinWithBytes pins content to Infura by given byte slice, it returns an IPFS hash and an error.
func (f *Fleek) PinWithBytes(buf []byte) (string, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	fn := helper.RandString(6, "lower")

	go func() {
		defer w.Close()
		defer m.Close()

		part, err := m.CreateFormFile("file", fn)
		if err != nil {
			return
		}

		if _, err = part.Write(buf); err != nil {
			return
		}
	}()

	return f.pinFile(r, m)
}

func (f *Fleek) pinFile(r *io.PipeReader, m *multipart.Writer) (string, error) {
	endpoint := api + "/upload"
	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", m.FormDataContentType())
	req.Header.Add("apiKey", f.Secret)
	req.Header.Add("iapiSecret", f.Apikey)

	client := httpretry.NewClient(nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(data, &dat); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return "", fmt.Errorf("json syntax error at byte offset %d", e.Offset)
		}
		return "", err
	}

	if out, err := dat["error"].(string); err {
		return "", fmt.Errorf(out)
	}
	if hash, ok := dat["IpfsHash"].(string); ok {
		return hash, nil
	}

	return "", fmt.Errorf("Pin file to Fleek failure.")
}

// PinHash pins content to Fleek by giving an IPFS hash, it returns the result and an error.
// Note: unsupported
func (f *Fleek) PinHash(hash string) (bool, error) {
	return true, nil
}
