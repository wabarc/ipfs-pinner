package file

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"sync"

	files "github.com/ipfs/go-ipfs-files"
)

// MultiFileReader reads from a `commands.Node` (which can be a directory of
// files or a regular file) as HTTP multipart encoded data.
type MultiFileReader struct {
	io.Reader

	mpWriter *multipart.Writer
	mutex    *sync.Mutex
}

// NewMultiFileReader constructs a files.MultiFileReader via github.com/ipfs/go-ipfs-files.
// `path` can be any `commands.Directory`. If `form` is set to true, the Content-Disposition
// will be "form-data". Otherwise, it will be "attachment".
//
// It returns an io.Reader and error.
func NewMultiFileReader(path string, form bool) (*files.MultiFileReader, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	file, err := files.NewSerialFile(path, false, stat)
	if err != nil {
		return nil, err
	}
	d := files.NewMapDirectory(map[string]files.Node{"": file}) // unwrapped on the other side

	return files.NewMultiFileReader(d, form), nil
}

// CreateMultiForm constructs a MultiFileReader. `path` should be a Node in serialfile.
// If `form` is set to true, the Content-Disposition will be "form-data".
// Otherwise, it will be "attachment".
//
// It returns an io.Reader and error.
//
// Example:
//
// > node, err := file.NewSerialFile("directory-path")
// >
// > node.MapDirectory("a-dir-name-show-in-pinning-service")
func CreateMultiForm(node *Node, form bool) (mfr *MultiFileReader, err error) {
	if len(node.files) == 0 {
		return mfr, fmt.Errorf("node.files empty")
	}

	dispositionPrefix := "attachment"
	if form {
		dispositionPrefix = "form-data"
	}

	// New multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// // Metadata part.
	// metadataHeader := textproto.MIMEHeader{}
	// metadataHeader.Set("Content-Disposition", fmt.Sprintf(`%s; filename="folderName"`, dispositionPrefix))
	// metadataHeader.Set("Content-Type", "application/x-directory")
	// // Metadata content.
	// part, err := writer.CreatePart(metadataHeader)
	// if err != nil {
	// 	return nil, fmt.Errorf("Error writing metadata headers: %v", err)
	// }
	// part.Write([]byte(metadataHeader))

	type meta struct {
		key  string
		name string
		data string
	}
	metadata := []meta{
		// For pinata
		{
			key:  "Content-Disposition",
			name: fmt.Sprintf(`%s; name="pinataMetadata"`, dispositionPrefix),
			data: fmt.Sprintf(`{"name":"%s"}`, filepath.Base(node.base)),
		},
		{
			key:  "Content-Disposition",
			name: fmt.Sprintf(`%s; name="pinataOptions"`, dispositionPrefix),
			data: `{"cidVersion":"1","wrapWithDirectory":false}`,
		},
	}
	for _, m := range metadata {
		header := textproto.MIMEHeader{}
		header.Set(m.key, m.name)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, fmt.Errorf("error writing metadata headers: %v", err)
		}
		_, _ = part.Write([]byte(m.data))
	}

	for _, fi := range node.files {
		fn := node.path
		if node.stat.IsDir() {
			fn = path.Join(node.path, fi.Name())
		}
		f, err := os.Open(fn)
		if err != nil {
			return mfr, fmt.Errorf("error reading media file: %v", err)
		}
		defer f.Close()

		filename := path.Join(node.base, filepath.Base(fn))
		// absPath, _ := filepath.Abs(fn)
		mediaHeader := textproto.MIMEHeader{}
		// mediaHeader.Set("Abspath", absPath)
		mediaHeader.Set("Content-Disposition", fmt.Sprintf(`%s; name="file"; filename="%s"`, dispositionPrefix, filename))
		mediaHeader.Set("Content-Type", "application/octet-stream")
		part, err := writer.CreatePart(mediaHeader)
		if err != nil {
			return mfr, fmt.Errorf("error writing media headers: %v", err)
		}

		if _, err := io.Copy(part, f); err != nil {
			return mfr, fmt.Errorf("error writing media: %v", err)
		}
	}

	// Close multipart writer.
	if err := writer.Close(); err != nil {
		return mfr, fmt.Errorf("error closing multipart writer: %v", err)
	}

	return &MultiFileReader{
		bytes.NewReader(body.Bytes()),
		writer,
		&sync.Mutex{},
	}, nil
}

// TODO: create multipart via Read
// func (mfr *MultiFileReader) Read(buf []byte) (written int, err error) {
// 	return
// }

func (mfr *MultiFileReader) Write(header textproto.MIMEHeader, content []byte) error {
	part, err := mfr.mpWriter.CreatePart(header)
	if err != nil {
		return fmt.Errorf("write header failed: %v", err)
	}
	_, _ = part.Write(content)

	return nil
}

// Boundary returns the boundary string to be used to separate files in the multipart data
func (mfr *MultiFileReader) Boundary() string {
	return mfr.mpWriter.Boundary()
}
