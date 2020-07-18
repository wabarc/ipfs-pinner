package pinata

import (
	"io/ioutil"
	"os"
	"testing"
)

var (
	APIKEY string
	SECRET string
)

func init() {
	APIKEY = os.Getenv("IPFS_PINNER_PINATA_API_KEY")
	SECRET = os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
}

func skip(t *testing.T) {
	if APIKEY == "" || SECRET == "" {
		t.Skip("Skipping testing in CI environment when without set secrets")
	}
}

func TestPinFile(t *testing.T) {
	skip(t)

	content := []byte("Hello, Pinata!")
	tmpfile, err := ioutil.TempFile("", "ipfs-pinner-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	pinata := &Pinata{Apikey: APIKEY, Secret: SECRET}
	if cid, err := pinata.PinFile(tmpfile.Name()); err != nil {
		t.Error(err)
	} else {
		t.Log(cid)
	}
}

func TestPinHash(t *testing.T) {
	skip(t)

	hash := "QmdKfpnTxbfzQL9Lyw3CMXwioVBScEb887Q4L6d9Q84bVw"

	pinata := &Pinata{Apikey: APIKEY, Secret: SECRET}
	if ok, err := pinata.PinHash(hash); !ok || err != nil {
		t.Error(err)
	}
}
