package pinata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner/file"

	httpretry "github.com/wabarc/ipfs-pinner/http"
)

const (
	PIN_FILE_URL = "https://api.pinata.cloud/pinning/pinFileToIPFS"
	PIN_HASH_URL = "https://api.pinata.cloud/pinning/pinByHash"
)

// Pinata represents a Pinata configuration.
type Pinata struct {
	*http.Client
	HttpClientFactory func(client *http.Client) *http.Client

	Apikey string
	Secret string
}

type addEvent struct {
	IpfsHash  string
	PinSize   int64  `json:",omitempty"`
	Timestamp string `json:",omitempty"`
}

// PinFile pins content to Pinata by providing a file path, it returns an IPFS
// hash and an error.
func (p *Pinata) PinFile(fp string) (string, error) {
	f, err := file.NewSerialFile(fp)
	if err != nil {
		return "", err
	}
	f.MapDirectory(filepath.Base(fp))

	mfr, err := file.CreateMultiForm(f, true)
	if err != nil {
		return "", err
	}
	boundary := "multipart/form-data; boundary=" + mfr.Boundary()

	return p.pinFile(mfr, boundary)
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

	return p.pinFile(r, m.FormDataContentType())
}

// PinWithBytes pins content to Infura by given byte slice, it returns an IPFS hash and an error.
func (p *Pinata) PinWithBytes(buf []byte) (string, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	fn := helper.RandString(6, "lower")

	go func() {
		defer w.Close()
		defer m.Close()

		// m.WriteField("pinataOptions", `{cidVersion: 1}`)
		part, err := m.CreateFormFile("file", fn)
		if err != nil {
			return
		}

		if _, err = part.Write(buf); err != nil {
			return
		}
	}()

	return p.pinFile(r, m.FormDataContentType())
}

func (p *Pinata) pinFile(r io.Reader, boundary string) (string, error) {
	// if fr, ok := r.(*file.MultiFileReader); ok {
	// 	// Metadata part.
	// 	metadataHeader := textproto.MIMEHeader{}
	// 	metadataHeader.Set("Content-Disposition", `form-data; name="pinataMetadata"`)
	// 	// Metadata content.
	// 	metadata := fmt.Sprintf(`{"name":"%s"}`, "adsasdfa")
	// 	fr.Write(metadataHeader, []byte(metadata))

	// 	// options part.
	// 	optsHeader := textproto.MIMEHeader{}
	// 	optsHeader.Set("Content-Disposition", `form-data; name="pinataOptions"`)
	// 	// options content.
	// 	opts := `{"cidVersion":"1","wrapWithDirectory":false}`
	// 	fr.Write(optsHeader, []byte(opts))
	// }
	req, err := http.NewRequest(http.MethodPost, PIN_FILE_URL, r)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", boundary)
	if p.Secret != "" && p.Apikey != "" {
		req.Header.Add("pinata_secret_api_key", p.Secret)
		req.Header.Add("pinata_api_key", p.Apikey)
	} else {
		req.Header.Add("Authorization", "Bearer "+p.Apikey)
	}

	if p.HttpClientFactory == nil {
		p.HttpClientFactory = httpretry.NewClient
	}
	client := p.HttpClientFactory(p.Client)
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

	return out.IpfsHash, nil
}

// PinHash pins content to Pinata by giving an IPFS hash, it returns the result and an error.
func (p *Pinata) PinHash(hash string) (bool, error) {
	if hash == "" {
		return false, fmt.Errorf("invalid hash: %s", hash)
	}

	jsonValue, _ := json.Marshal(map[string]string{"hashToPin": hash})

	req, err := http.NewRequest(http.MethodPost, PIN_HASH_URL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.Secret != "" && p.Apikey != "" {
		req.Header.Add("pinata_secret_api_key", p.Secret)
		req.Header.Add("pinata_api_key", p.Apikey)
	} else {
		req.Header.Add("Authorization", "Bearer "+p.Apikey)
	}

	if p.HttpClientFactory == nil {
		p.HttpClientFactory = httpretry.NewClient
	}
	client := p.HttpClientFactory(p.Client)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

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

	if h, ok := dat["hashToPin"].(string); ok {
		return h == hash, nil
	}

	return false, fmt.Errorf("pin hash to Pinata failed")
}

// PinDir pins a directory to the Pinata pinning service.
// It alias to PinFile.
func (p *Pinata) PinDir(name string) (string, error) {
	return p.PinFile(name)
}
