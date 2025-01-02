package errortypes

import (
	"fmt"

	"github.com/google/uuid"
)

type ErrImageNotFound struct {
	BusinessError
}

func NewErrImageNotFound(id uuid.UUID) ServiceError {
	return &ErrImageNotFound{
		BusinessError: BusinessError{
			StatusCode: 404,
			Code:       "IMAGE_NOT_FOUND",
			Detail:     fmt.Sprintf("Image with id %s not found", id),
		},
	}
}

type ErrNoImageAvailable struct {
	BusinessError
}

func NewErrNoImageAvailable(modelName string) ServiceError {
	return &ErrNoImageAvailable{
		BusinessError: BusinessError{
			StatusCode: 404,
			Code:       "NO_IMAGE_AVAILABLE",
			Detail:     fmt.Sprintf("No image available for model %s", modelName),
		},
	}
}

type ErrSearchQueryNotFound struct {
	BusinessError
}

func NewErrSearchQueryNotFound(queryID uuid.UUID) ServiceError {
	return &ErrSearchQueryNotFound{
		BusinessError: BusinessError{
			StatusCode: 404,
			Code:       "SEARCH_QUERY_NOT_FOUND",
			Detail:     fmt.Sprintf("Search query with id %s not found", queryID),
		},
	}
}

type ErrSearchFeedbackAlreadyExists struct {
	BusinessError
}

func NewErrSearchFeedbackAlreadyExists(queryID uuid.UUID) ServiceError {
	return &ErrSearchFeedbackAlreadyExists{
		BusinessError: BusinessError{
			StatusCode: 400,
			Code:       "SEARCH_FEEDBACK_ALREADY_EXISTS",
			Detail:     fmt.Sprintf("Search feedback for query with id %s already exists", queryID),
		},
	}
}
