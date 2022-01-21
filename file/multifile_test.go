package file

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/wabarc/helper"
)

func TestCreateMultiForm(t *testing.T) {
	dir, err := ioutil.TempDir("", "ipfs-pinner-dir-")
	if err != nil {
		t.Fatalf("Unexpected create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create files
	for i := 1; i <= 2; i++ {
		f, err := ioutil.TempFile(dir, "file-")
		if err != nil {
			t.Fatal("Unexpected create file")
		}
		content := []byte(helper.RandString(6, "lower"))
		if _, err := f.Write(content); err != nil {
			t.Fatal("Unexpected write content to file")
		}
	}

	node, err := NewSerialFile(dir)
	if err != nil {
		t.Fatalf("Unexpected new a serial file: %v", err)
	}
	node.MapDirectory("a-dir-name-show-in-pinning-service")

	body, err := CreateMultiForm(node, true)
	if err != nil {
		t.Fatalf("Unexpected creates multipart form data: %v", err)
	}

	boundary := body.Boundary()
	if boundary == "" {
		t.Fatal("Unexpected multipart boundary")
	}
}
