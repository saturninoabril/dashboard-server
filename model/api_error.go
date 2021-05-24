package model

import (
	"encoding/json"
	"io"
)

// APIError is the error struct returned by REST API endpoints.
type APIError struct {
	Message string `json:"message"`
}

// APIErrorFromReader decodes a json-encoded APIError from the given io.Reader.
func APIErrorFromReader(reader io.Reader) (*APIError, error) {
	apiErr := APIError{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&apiErr)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &apiErr, nil
}
