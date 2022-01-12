package pinata

import (
	"bytes"
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

const (
	PIN_FILE_URL = "https://api.pinata.cloud/pinning/pinFileToIPFS"
	PIN_HASH_URL = "https://api.pinata.cloud/pinning/pinByHash"
)

// Pinata represents a Pinata configuration.
type Pinata struct {
	Apikey string
	Secret string
}

// PinFile pins content to Pinata by providing a file path, it returns an IPFS
// hash and an error.
func (p *Pinata) PinFile(fp string) (string, error) {
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

	return p.pinFile(r, m)
}

// PinWithReader pins content to Pinata by given io.Reader, it returns an IPFS hash and an error.
func (p *Pinata) PinWithReader(rd io.Reader) (string, error) {
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

	return p.pinFile(r, m)
}

// PinWithBytes pins content to Infura by given byte slice, it returns an IPFS hash and an error.
func (p *Pinata) PinWithBytes(buf []byte) (string, error) {
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

	return p.pinFile(r, m)
}

func (p *Pinata) pinFile(r *io.PipeReader, m *multipart.Writer) (string, error) {
	req, err := http.NewRequest(http.MethodPost, PIN_FILE_URL, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", m.FormDataContentType())
	req.Header.Add("pinata_secret_api_key", p.Secret)
	req.Header.Add("pinata_api_key", p.Apikey)

	client := httpretry.NewClient(nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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

	return "", fmt.Errorf("Pin file to Pinata failure.")
}

// PinHash pins content to Pinata by giving an IPFS hash, it returns the result and an error.
func (p *Pinata) PinHash(hash string) (bool, error) {
	if hash == "" {
		return false, fmt.Errorf("HASH: %s is invalid.", hash)
	}

	jsonValue, _ := json.Marshal(map[string]string{"hashToPin": hash})

	req, err := http.NewRequest(http.MethodPost, PIN_HASH_URL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("pinata_secret_api_key", p.Secret)
	req.Header.Add("pinata_api_key", p.Apikey)

	client := httpretry.NewClient(nil)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(data, &dat); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return false, fmt.Errorf("json syntax error at byte offset %d", e.Offset)
		}
		return false, err
	}

	if h, ok := dat["ipfsHash"].(string); ok {
		return h == hash, nil
	}

	return false, fmt.Errorf("Pin hash to Pinata failure.")
}
