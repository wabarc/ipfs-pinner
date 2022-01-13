package nftstorage

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

const api = "https://api.nft.storage"

// NFTStorage represents an NFTStorage configuration.
type NFTStorage struct {
	Apikey string

	client *http.Client
}

// PinFile pins content to NFTStorage by providing a file path, it returns an IPFS
// hash and an error.
func (nft *NFTStorage) PinFile(fp string) (string, error) {
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

	return nft.pinFile(r, m)
}

// PinWithReader pins content to NFTStorage by given io.Reader, it returns an IPFS hash and an error.
func (nft *NFTStorage) PinWithReader(rd io.Reader) (string, error) {
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

	return nft.pinFile(r, m)
}

// PinWithBytes pins content to NFTStorage by given byte slice, it returns an IPFS hash and an error.
func (nft *NFTStorage) PinWithBytes(buf []byte) (string, error) {
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

	return nft.pinFile(r, m)
}

func (nft *NFTStorage) pinFile(r *io.PipeReader, m *multipart.Writer) (string, error) {
	endpoint := api + "/upload"

	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", m.FormDataContentType())
	if nft.Apikey != "" {
		req.Header.Add("Authorization", "Bearer "+nft.Apikey)
	}
	client := httpretry.NewClient(nft.client)
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

	if ok := dat["ok"].(bool); !ok {
		return "", fmt.Errorf("Pin file to NFTStorage failed.")
	}
	value, ok := dat["value"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Pin file to NFTStorage failed: no value")
	}
	if cid, ok := value["cid"].(string); ok {
		return cid, nil
	}

	return "", fmt.Errorf("Pin file to NFTStorage failure.")
}

// PinHash pins content to NFTStorage by giving an IPFS hash, it returns the result and an error.
// Note: unsupported
func (nft *NFTStorage) PinHash(hash string) (bool, error) {
	return true, nil
}
