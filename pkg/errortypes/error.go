package errortypes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ServiceError interface {
	GetStatusCode() int
	GetErrorCode() string
	GetErrorDetail() string
	Error() string
}

type BusinessError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Detail     string `json:"detail"`
	Err        error  `json:"-"`
}

func NewBusinessError(code string, detail string, err error) ServiceError {
	return &BusinessError{
		StatusCode: 500,
		Code:       code,
		Detail:     detail,
		Err:        err,
	}
}

func (e *BusinessError) GetStatusCode() int {
	return e.StatusCode
}

func (e *BusinessError) GetErrorCode() string {
	return e.Code
}

func (e *BusinessError) GetErrorDetail() string {
	return e.Detail
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("%s: %s", e.GetErrorCode(), e.GetErrorDetail())
}

type InternalError struct {
	BusinessError
}

func NewInternalError(err error) ServiceError {
	return &InternalError{
		BusinessError: BusinessError{
			StatusCode: 500,
			Code:       "INTERNAL_ERROR",
			Detail:     err.Error(),
			Err:        err,
		},
	}
}

func ErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	var svcerror ServiceError
	if !errors.As(err, &svcerror) {
		svcerror = NewInternalError(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(svcerror.GetStatusCode())
	json.NewEncoder(w).Encode(svcerror)
}
