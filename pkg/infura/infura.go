package infura

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	INFURA_HOST = "ipfs.infura.io"
	INFURA_PORT = 5001

	INFURA_PROTOCAL = "https"
)

func PinFile(filepath string) (string, error) {
	uri := fmt.Sprintf("%s://%s:%d/api/v0/add", INFURA_PROTOCAL, INFURA_HOST, INFURA_PORT)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", filepath)
	if err != nil {
		return "", err
	}
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err = io.Copy(part, file); err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(http.MethodPost, uri, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
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
		return "", err
	}

	if hash, ok := dat["Hash"].(string); ok {
		return hash, nil
	}

	return "", fmt.Errorf("Pin file to Infura failure.")
}

func PinHash(hash string) (bool, error) {
	if hash == "" {
		return false, fmt.Errorf("HASH: %s is invalid.", hash)
	}

	uri := fmt.Sprintf("%s://%s:%d/api/v0/pin/add?arg=%s", INFURA_PROTOCAL, INFURA_HOST, INFURA_PORT, hash)
	resp, err := http.Get(uri)
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
		return false, err
	}

	if h, ok := dat["Pins"].([]interface{}); ok {
		return h[0] == hash, nil
	}

	return false, fmt.Errorf("Pin hash to Infura failure.")
}
