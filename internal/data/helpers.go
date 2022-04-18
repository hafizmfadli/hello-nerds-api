package data

import (
	"encoding/json"
	"io"
	"strings"
)

type esNativeResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string `json:"_id"`
			Source     json.RawMessage `json:"_source"`
		}
	}
}

func buildQuery(query string) io.Reader {
	return strings.NewReader(query)
}