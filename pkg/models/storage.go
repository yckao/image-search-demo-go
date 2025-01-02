package models

import "io"

type StorageFile struct {
	Provider string
	Key      string
}

type StorageFileStream struct {
	Reader        io.Reader
	ContentType   string
	ContentLength int64
	Filename      string
}
