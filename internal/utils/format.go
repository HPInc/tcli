// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"openapi/internal/common"
	"strings"

	"github.com/itchyny/gojq"
)

const (
	varParams = "$params"
)

func DoFormat(bytes []byte, v common.Values, f string) {
	var i gojq.Iter
	// check if variables processing needed
	if strings.Contains(f, varParams) {
		i = getIterWithVariables(bytes, v, f)
	} else {
		i = getIter(bytes, f)
	}
	printFormat(i)
}

// run gojq without variables, return iterator
func getIter(bytes []byte, f string) gojq.Iter {
	query, err := gojq.Parse(f)
	if err != nil {
		log.Fatalln(err)
	}
	out, err := unmarshalAny(bytes)
	if err != nil {
		log.Fatalln(err)
	}
	return query.Run(out)
}

// apply variables and run gojq, return iterator
func getIterWithVariables(bytes []byte, v common.Values, f string) gojq.Iter {
	query, err := gojq.Parse(f)
	if err != nil {
		log.Fatalln(err)
	}
	code, err := gojq.Compile(
		query,
		gojq.WithVariables([]string{
			"$params",
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	out, err := unmarshalAny(bytes)
	if err != nil {
		log.Fatalln(err)
	}
	return code.Run(out, mapJsonParams(v))
}

func printFormat(iter gojq.Iter) {
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			log.Fatalln(err)
		}
		fmt.Printf("%s", common.GetJsonString(v))
	}
}

func mapJsonParams(i common.Values) map[string]any {
	r := make(map[string]any)
	for k, v := range i {
		js, _ := common.TryUnmarshalJson(v)
		if js != nil {
			r[k] = js
		} else {
			r[k] = v
		}
	}
	return r
}

// helper method to specifically unmarshal to an any
func unmarshalAny(bytes []byte) (any, error) {
	var result any
	err := json.Unmarshal(bytes, &result)
	return result, err
}
