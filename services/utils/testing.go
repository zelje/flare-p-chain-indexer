package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func JsonToReader(t *testing.T, json string) io.Reader {
	return bytes.NewReader([]byte(json))
}

func StructToReader(t *testing.T, s any) io.Reader {
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(data)
}

func DecodeStruct(t *testing.T, r io.Reader, v any) {
	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)
	if err != nil {
		t.Fatal(err)
	}
}
