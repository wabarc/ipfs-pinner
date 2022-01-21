package nftstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/wabarc/ipfs-pinner/file"

	httpretry "github.com/wabarc/ipfs-pinner/http"
)

const api = "https://api.nft.storage"

// NFTStorage represents an NFTStorage configuration.
type NFTStorage struct {
	Apikey string

	client *http.Client
}

type value struct {
	Cid     string
	Size    int64  `json:",omitempty"`
	Created string `json:",omitempty"`
	Type    string `json:",omitempty"`
}

type er struct {
	Name, Message string
}

type addEvent struct {
	Ok    bool
	Value value
	Error er
}

// PinFile pins content to NFTStorage by providing a file path, it returns an IPFS
// hash and an error.
func (nft *NFTStorage) PinFile(fp string) (string, error) {
	fi, err := os.Stat(fp)
	if err != nil {
		return "", err
	}

	// For regular file
	if fi.Mode().IsRegular() {
		f, err := os.Open(fp)
		if err != nil {
			return "", err
		}
		defer f.Close()

		return nft.pinFile(f, file.MediaType(f))
	}

	// For directory, or etc
	f, err := file.NewSerialFile(fp)
	if err != nil {
		return "", err
	}

	mfr, err := file.CreateMultiForm(f, true)
	if err != nil {
		return "", err
	}
	boundary := "multipart/form-data; boundary=" + mfr.Boundary()

	return nft.pinFile(mfr, boundary)
}

// PinWithReader pins content to NFTStorage by given io.Reader, it returns an IPFS hash and an error.
func (nft *NFTStorage) PinWithReader(rd io.Reader) (string, error) {
	return nft.pinFile(rd, file.MediaType(rd))
}

// PinWithBytes pins content to NFTStorage by given byte slice, it returns an IPFS hash and an error.
func (nft *NFTStorage) PinWithBytes(buf []byte) (string, error) {
	return nft.pinFile(bytes.NewReader(buf), file.MediaType(buf))
}

func (nft *NFTStorage) pinFile(r io.Reader, boundary string) (string, error) {
	endpoint := api + "/upload"

	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", boundary)
	req.Header.Add("Authorization", "Bearer "+nft.Apikey)
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

	var out addEvent
	if err := json.Unmarshal(data, &out); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return "", fmt.Errorf("json syntax error at byte offset %d", e.Offset)
		}
		return "", err
	}

	return out.Value.Cid, nil
}

// PinHash pins content to NFTStorage by giving an IPFS hash, it returns the result and an error.
// Note: unsupported
func (nft *NFTStorage) PinHash(hash string) (bool, error) {
	return false, fmt.Errorf("not yet supported")
}

// PinDir pins a directory to the NFT.Storage pinning service.
// It alias to PinFile.
func (nft *NFTStorage) PinDir(name string) (string, error) {
	return nft.PinFile(name)
}
