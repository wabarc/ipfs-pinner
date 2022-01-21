package infura

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner/file"

	files "github.com/ipfs/go-ipfs-files"
	httpretry "github.com/wabarc/ipfs-pinner/http"
)

const api = "https://ipfs.infura.io:5001"

// Infura represents an Infura configuration. If there is no ProjectID or
// ProjectSecret, it will make API calls using anonymous requests.
type Infura struct {
	ProjectID     string
	ProjectSecret string

	client *http.Client
}

type addEvent struct {
	Name  string
	Hash  string `json:",omitempty"`
	Bytes int64  `json:",omitempty"`
	Size  string `json:",omitempty"`
}

// PinFile alias to *Infura.PinFile, the purpose is to be backwards
// compatible with the original function.
func PinFile(fp string) (string, error) {
	return (&Infura{}).PinFile(fp)
}

// PinFile pins content to Infura by providing a file path, it returns an IPFS
// hash and an error.
func (inf *Infura) PinFile(fp string) (string, error) {
	mfr, err := file.NewMultiFileReader(fp, false)
	if err != nil {
		return "", fmt.Errorf("unexpected creates multipart file: %v", err)
	}
	boundary := "multipart/form-data; boundary=" + mfr.Boundary()

	return inf.pinFile(mfr, boundary)
}

// PinWithReader pins content to Infura by given io.Reader, it returns an IPFS hash and an error.
func (inf *Infura) PinWithReader(rd io.Reader) (string, error) {
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

	return inf.pinFile(r, m.FormDataContentType())
}

// PinWithBytes pins content to Infura by given byte slice, it returns an IPFS hash and an error.
func (inf *Infura) PinWithBytes(buf []byte) (string, error) {
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

	return inf.pinFile(r, m.FormDataContentType())
}

func (inf *Infura) pinFile(r io.Reader, boundary string) (string, error) {
	endpoint := api + "/api/v0/add?cid-version=1&pin=true"
	client := httpretry.NewClient(inf.client)
	// client.Timeout = 30 * time.Second

	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return "", err
	}
	if inf.ProjectID != "" && inf.ProjectSecret != "" {
		req.SetBasicAuth(inf.ProjectID, inf.ProjectSecret)
	}
	req.Header.Add("Content-Type", boundary)
	req.Header.Set("Content-Disposition", `form-data; name="files"`)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// It limits anonymous requests to 12 write requests/min.
	// https://infura.io/docs/ipfs#section/Rate-Limits/API-Anonymous-Requests
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(resp.Status)
	}

	var out addEvent
	dec := json.NewDecoder(resp.Body)
loop:
	for {
		var evt addEvent
		switch err := dec.Decode(&evt); err {
		case nil:
		case io.EOF:
			break loop
		default:
			return "", err
		}
		out = evt
	}

	return out.Hash, nil
}

// PinHash alias to *Infura.PinHash, the purpose is to be backwards
// compatible with the original function.
func PinHash(hash string) (bool, error) {
	return (&Infura{}).PinHash(hash)
}

// PinHash pins content to Infura by giving an IPFS hash, it returns the result and an error.
func (inf *Infura) PinHash(hash string) (bool, error) {
	if hash == "" {
		return false, fmt.Errorf("invalid hash: %s", hash)
	}

	endpoint := fmt.Sprintf("%s/api/v0/pin/add?arg=%s", api, hash)
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return false, err
	}
	if inf.ProjectID != "" && inf.ProjectSecret != "" {
		req.SetBasicAuth(inf.ProjectID, inf.ProjectSecret)
	}
	client := httpretry.NewClient(inf.client)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// It limits anonymous requests to 12 write requests/min.
	// https://infura.io/docs/ipfs#section/Rate-Limits/API-Anonymous-Requests
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf(resp.Status)
	}

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

	if h, ok := dat["Pins"].([]interface{}); ok && len(h) > 0 {
		return h[0] == hash, nil
	}

	return false, fmt.Errorf("pin hash to Infura failed")
}

// PinDir pins a directory to the Infura pinning service.
func (inf *Infura) PinDir(mfr *files.MultiFileReader) (string, error) {
	boundary := "multipart/form-data; boundary=" + mfr.Boundary()
	return inf.pinFile(mfr, boundary)
}
