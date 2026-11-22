package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

// encodes value as json and sends it with given status
// encodes into a buffer first, if fails, nothing sent
// the caller respond with 500
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes()) // no error handling: write error means client disconnect, no meaning writing it
	return nil
}

// read body with MaxBytesReader and DisallowUnknownFields
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	return errors.New("Not implemented yet")
}
