package utils

import (
	"net"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
)

func ResponseWasNotFound(resp autorest.Response) bool {
	return ResponseWasStatusCode(resp, http.StatusNotFound)
}

func ResponseWasBadRequest(resp autorest.Response) bool {
	return ResponseWasStatusCode(resp, http.StatusBadRequest)
}

func ResponseWasForbidden(resp autorest.Response) bool {
	return ResponseWasStatusCode(resp, http.StatusForbidden)
}

func ResponseWasConflict(resp autorest.Response) bool {
	return ResponseWasStatusCode(resp, http.StatusConflict)
}

func ResponseErrorIsRetryable(err error) bool {
	if arerr, ok := err.(autorest.DetailedError); ok {
		err = arerr.Original
	}

	// nolint:gocritic
	switch e := err.(type) {
	case net.Error:
		if e.Temporary() || e.Timeout() { // nolint:staticcheck
			return true
		}
	}

	return false
}

func ResponseWasStatusCode(resp autorest.Response, statusCode int) bool {
	if r := resp.Response; r != nil {
		if r.StatusCode == statusCode {
			return true
		}
	}

	return false
}

// TODO THIS NEEDS TO MOVE TO HELPERS

// WasConflict returns true if the HttpResponse is non-nil and has a status code of Conflict
func WasConflict(resp *http.Response) bool {
	return responseWasStatusCode(resp, http.StatusConflict)
}

// WasNotFound returns true if the HttpResponse is non-nil and has a status code of NotFound
func WasNotFound(resp *http.Response) bool {
	return responseWasStatusCode(resp, http.StatusNotFound)
}

func responseWasStatusCode(resp *http.Response, statusCode int) bool {
	if r := resp; r != nil {
		if r.StatusCode == statusCode {
			return true
		}
	}

	return false
}
