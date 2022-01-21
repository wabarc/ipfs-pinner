package web3storage

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner/file"

	httpretry "github.com/wabarc/ipfs-pinner/http"
)

const api = "https://api.web3.storage"

// Web3Storage represents a Web3Storage configuration.
type Web3Storage struct {
	Apikey string

	client *http.Client
}

type addEvent struct {
	Cid string
}

// PinFile pins content to Web3Storage by providing a file path, it returns an IPFS
// hash and an error.
func (web3 *Web3Storage) PinFile(fp string) (string, error) {
	f, err := file.NewSerialFile(fp)
	if err != nil {
		return "", err
	}
	f.MapDirectory(helper.RandString(32, "lower"))

	mfr, err := file.CreateMultiForm(f, true)
	if err != nil {
		return "", err
	}
	boundary := "multipart/form-data; boundary=" + mfr.Boundary()

	return web3.pinFile(mfr, boundary)
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

	return web3.pinFile(r, m.FormDataContentType())
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

	return web3.pinFile(r, m.FormDataContentType())
}

func (web3 *Web3Storage) pinFile(r io.Reader, boundary string) (string, error) {
	endpoint := api + "/upload"

	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", boundary)
	req.Header.Add("Authorization", "Bearer "+web3.Apikey)
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

	var out addEvent
	if err := json.Unmarshal(data, &out); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return "", fmt.Errorf("json syntax error at byte offset %d", e.Offset)
		}
		return "", err
	}

	return out.Cid, nil
}

// PinHash pins content to Web3Storage by giving an IPFS hash, it returns the result and an error.
// Note: unsupported
func (web3 *Web3Storage) PinHash(hash string) (bool, error) {
	return false, fmt.Errorf("not yet supported")
}

// PinDir pins a directory to the Pinata pinning service.
// It alias to PinFile.
func (web3 *Web3Storage) PinDir(name string) (string, error) {
	return web3.PinFile(name)
}
