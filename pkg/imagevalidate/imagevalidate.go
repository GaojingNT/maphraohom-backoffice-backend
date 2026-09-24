// Package imagevalidate validates uploaded image files by magic bytes
// (never trusting the client-supplied filename/Content-Type) — shared by
// every module that accepts an image upload (bill slips, store logo/signature).
package imagevalidate

import (
	"errors"
	"mime/multipart"

	"github.com/gabriel-vasile/mimetype"
)

// MaxFileSize is the upload size cap shared by every image upload endpoint.
const MaxFileSize = 10 * 1024 * 1024

// ErrUnsupported is returned by DetectExtension when the file's magic bytes
// don't match any of jpeg/png/webp.
var ErrUnsupported = errors.New("unsupported image type")

// extByMIME maps a magic-byte-detected MIME type to the file extension its
// object key is stored under — only these three are accepted.
var extByMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// DetectExtension reads the file's magic bytes and returns the extension its
// object key should be stored under, or ErrUnsupported if it isn't
// jpeg/png/webp.
func DetectExtension(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	mtype, err := mimetype.DetectReader(f)
	if err != nil {
		return "", err
	}

	// mimetype.Is walks the detected type's parent chain (e.g. some jpeg
	// variants), so match by string on the parents too.
	for m := mtype; m != nil; m = m.Parent() {
		if ext, ok := extByMIME[m.String()]; ok {
			return ext, nil
		}
	}

	return "", ErrUnsupported
}
