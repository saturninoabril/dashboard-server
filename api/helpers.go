package api

import (
	"encoding/json"
	"io"
)

func decodeJSON(obj interface{}, body io.ReadCloser) error {
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&obj)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}
