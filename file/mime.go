package file

import (
	"bufio"
	"io"
	"net/http"
	"os"
	// "github.com/gabriel-vasile/mimetype"
)

// MediaType returns the file's mime type. If the mime type cannot be
// determined, it returns "application/octet-stream".
//
// The i should be a *os.File, io.Reader, or byte slice.
func MediaType(i interface{}) string {
	defaultType := "application/octet-stream"

	// var err error
	// var mtype *mimetype.MIME
	// switch v := i.(type) {
	// case *os.File:
	// 	mtype, err = mimetype.DetectFile(v.Name())
	// case io.Reader:
	// 	mtype, err = mimetype.DetectReader(v)
	// case []byte:
	// 	mtype = mimetype.Detect(v)
	// default:
	// 	return defaultType
	// }

	// if err != nil {
	// 	return defaultType
	// }
	// return mtype.String()

	switch v := i.(type) {
	case *os.File:
		scanner := bufio.NewScanner(v)
		return http.DetectContentType(scanner.Bytes())
	case io.Reader:
		rd := bufio.NewReader(v)
		buf, err := rd.Peek(8192)
		if err != nil {
			return defaultType
		}
		return http.DetectContentType(buf)
	case []byte:
		return http.DetectContentType(v)
	}

	return defaultType
}
