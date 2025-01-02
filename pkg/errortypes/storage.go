package errortypes

import "fmt"

// var ErrStorageProviderNotFound = errors.New("storage provider not found")
// var ErrStorageFileNotFound = errors.New("storage file not found")

type ErrStorageFileNotFound struct {
	BusinessError
	Provider string `json:"-"`
	Key      string `json:"-"`
}

func NewErrStorageFileNotFound(provider string, key string) ServiceError {
	return &ErrStorageFileNotFound{
		BusinessError: BusinessError{
			StatusCode: 404,
			Code:       "OBJECT_NOT_FOUND",
			Detail:     fmt.Sprintf("Object not found in %s: %s", provider, key),
		},
		Provider: provider,
		Key:      key,
	}
}
