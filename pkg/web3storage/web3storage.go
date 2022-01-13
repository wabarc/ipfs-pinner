package web3storage

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

const api = "https://api.web3.storage"

// Web3Storage represents a Web3Storage configuration.
type Web3Storage struct {
	Apikey string

	client *http.Client
}

// PinFile pins content to Web3Storage by providing a file path, it returns an IPFS
// hash and an error.
func (web3 *Web3Storage) PinFile(fp string) (string, error) {
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

	return web3.pinFile(r, m)
}

// PinWithReader pins content to Web3Storage by given io.Reader, it returns an IPFS hash and an error.
func (web3 *Web3Storage) PinWithReader(rd io.Reader) (string, error) {
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

	return web3.pinFile(r, m)
}

// PinWithBytes pins content to Web3Storage by given byte slice, it returns an IPFS hash and an error.
func (web3 *Web3Storage) PinWithBytes(buf []byte) (string, error) {
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

	return web3.pinFile(r, m)
}

func (web3 *Web3Storage) pinFile(r *io.PipeReader, m *multipart.Writer) (string, error) {
	endpoint := api + "/upload"

	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", m.FormDataContentType())
	if web3.Apikey != "" {
		req.Header.Add("Authorization", "Bearer "+web3.Apikey)
	}
	client := httpretry.NewClient(web3.client)
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

	if cid, ok := dat["cid"].(string); ok {
		return cid, nil
	}

	return "", fmt.Errorf("Pin file to Web3Storage failure.")
}

// PinHash pins content to Web3Storage by giving an IPFS hash, it returns the result and an error.
// Note: unsupported
func (web3 *Web3Storage) PinHash(hash string) (bool, error) {
	return true, nil
}
