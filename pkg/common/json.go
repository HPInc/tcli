// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package common

import (
	"bytes"
	"encoding/json"
	"log"
)

func GetJsonString(r interface{}) string {
	buffer := &bytes.Buffer{}
	e := json.NewEncoder(buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(r)
	if err != nil {
		log.Println("could not encode json", err)
		return ""
	}
	return buffer.String()
}

// check if a given string is valid json
// if valid, returns raw bytes of unmarshalled json
func TryUnmarshalJson(str *string) (*json.RawMessage, error) {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(*str), &js); err != nil {
		return nil, err
	}
	return &js, nil
}
