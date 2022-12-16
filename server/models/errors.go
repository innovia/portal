package models

import (
	"encoding/json"
	"fmt"
)

// ClientError is an error whose details to be shared with client.
// the reason for this being an interface instead of a simple struct
// is to avoid type assertions, since this will require the error handler to know
// every custom error in the package that can be returned and assert them
// it also decouples the handler function from each error
type ClientError interface {
	Error() string

	// ResponseBody returns JSON response body of the error (title, message, error code…) in bytes
	ResponseBody() ([]byte, error)

	// ResponseHeaders returns http status code and headers.
	ResponseHeaders() (int, map[string]string)
}

// HTTPError implements ClientError interface.
type HTTPError struct {
	Cause  error  `json:"-"`
	Detail string `json:"detail"`
	Status int    `json:"-"`
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}

// ResponseBody returns JSON response body.
func (e *HTTPError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("error while parsing response body: %v", err)
	}
	return body, nil
}

// ResponseHeaders returns http status code and headers.
func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

// NewHTTPError will hold the http errors with status code and details
func NewHTTPError(err error, status int, detail string) error {
	return &HTTPError{
		Cause:  err,
		Detail: detail,
		Status: status,
	}
}
