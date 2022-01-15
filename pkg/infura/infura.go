package infura

import (
	"encoding/base64"
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

const api = "https://ipfs.infura.io:5001"

// Infura represents an Infura configuration. If there is no ProjectID or
// ProjectSecret, it will make API calls using anonymous requests.
type Infura struct {
	ProjectID     string
	ProjectSecret string

	client *http.Client
}

// PinFile alias to *Infura.PinFile, the purpose is to be backwards
// compatible with the original function.
func PinFile(fp string) (string, error) {
	return (&Infura{}).PinFile(fp)
}

// PinFile pins content to Infura by providing a file path, it returns an IPFS
// hash and an error.
func (inf *Infura) PinFile(fp string) (string, error) {
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

	return inf.pinFile(r, m)
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

	return inf.pinFile(r, m)
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

	return inf.pinFile(r, m)
}

func (inf *Infura) pinFile(r *io.PipeReader, m *multipart.Writer) (string, error) {
	endpoint := api + "/api/v0/add"
	client := httpretry.NewClient(inf.client)
	// client.Timeout = 30 * time.Second

	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", m.FormDataContentType())
	if inf.ProjectID != "" && inf.ProjectSecret != "" {
		req.Header.Add("Authorization", "Basic "+basicAuth(inf.ProjectID, inf.ProjectSecret))
	}
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

	if hash, ok := dat["Hash"].(string); ok {
		return hash, nil
	}

	return "", fmt.Errorf("Pin file to Infura failure.")
}

// PinHash alias to *Infura.PinHash, the purpose is to be backwards
// compatible with the original function.
func PinHash(hash string) (bool, error) {
	return (&Infura{}).PinHash(hash)
}

// PinHash pins content to Infura by giving an IPFS hash, it returns the result and an error.
func (inf *Infura) PinHash(hash string) (bool, error) {
	if hash == "" {
		return false, fmt.Errorf("HASH: %s is invalid.", hash)
	}

	endpoint := fmt.Sprintf("%s/api/v0/pin/add?arg=%s", api, hash)
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return false, err
	}
	if inf.ProjectID != "" && inf.ProjectSecret != "" {
		req.Header.Add("Authorization", "Basic "+basicAuth(inf.ProjectID, inf.ProjectSecret))
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

	return false, fmt.Errorf("Pin hash to Infura failure.")
}

func basicAuth(projectID, projectSecret string) string {
	auth := projectID + ":" + projectSecret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
