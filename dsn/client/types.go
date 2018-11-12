package client

import (
	"io"
)

type Client interface {
	// UploadRaw uploads raw data to dsn and returns the resulting hash. If toEncrypt is true it
	// uploads encrypted data
	UploadRaw(r io.Reader, size int64, toEncrypt bool) (string, error)
	// DownloadRaw downloads raw data from dsn and it returns a ReadCloser and a bool whether the
	// content was encrypted
	DownloadRaw(hash string) (io.ReadCloser, bool, error)
}