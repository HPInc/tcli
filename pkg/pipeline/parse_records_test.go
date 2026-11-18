// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseRecords(t *testing.T) {
	multipart := "--b1\r\nContent-Disposition: form-data; name=\"meta\"\r\nContent-Type: application/json\r\n\r\n" +
		"{\"a\":1}\r\n--b1\r\nContent-Disposition: form-data; name=\"file\"\r\n\r\nline one\r\nline two\r\n--b1--\r\n"

	tests := []struct {
		name    string
		in      string
		want    []Record
		wantErr bool
	}{
		// JSON keeps behaving as before.
		{"empty output has no records", "", nil, false},
		{"blank output has no records", " \n\t\n", nil, false},
		{"object", "{\"id\":1}\n", []Record{{"id": float64(1)}}, false},
		{"array of objects is flattened", `[{"id":1},{"id":2}]`, []Record{{"id": float64(1)}, {"id": float64(2)}}, false},
		{"array elements that are not objects are skipped", `[1,{"id":2},"x"]`, []Record{{"id": float64(2)}}, false},
		{"one object per line", "{\"id\":1}\n{\"id\":2}\n", []Record{{"id": float64(1)}, {"id": float64(2)}}, false},
		{"a JSON number has no record", "12345\n", nil, false},
		{"a JSON string has no record", "\"abc\"\n", nil, false},

		// Output that is not JSON becomes one text record.
		{"plain text", "abc123\n", []Record{{"text": "abc123"}}, false},
		{"plain text that starts with digits", "6561f0a9\n", []Record{{"text": "6561f0a9"}}, false},
		{"outer whitespace is trimmed", "  \n abc123 \r\n\n", []Record{{"text": "abc123"}}, false},
		{"inner lines are kept", "line one\nline two\n", []Record{{"text": "line one\nline two"}}, false},
		{"html", "<html><body>hi</body></html>\n", []Record{{"text": "<html><body>hi</body></html>"}}, false},
		{"multipart body", multipart + "\n", []Record{{"text": strings.TrimSpace(multipart)}}, false},

		// Output that starts like JSON but is not valid JSON is still an error.
		{"truncated object", `{"id":6,`, nil, true},
		{"truncated array", `[{"a":1},`, nil, true},
		{"object followed by text", "{\"a\":1}\nabc\n", nil, true},
		{"malformed object after whitespace", "\n  {bad", nil, true},
		{"text that starts with a bracket", "[OK] done\n", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRecords([]byte(tt.in))
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "decode tcli output") {
					t.Fatalf("parseRecords(%q) = %v, %v; want a decode error", tt.in, got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRecords(%q) error: %v", tt.in, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRecords(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}
