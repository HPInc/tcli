// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package petstoretest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Sanity tests for the fake server. These run every time (no gate) because
// they are self-contained and fast no external tcli invocation involved.

func TestServer_AddThenGet(t *testing.T) {
	srv := NewServer()
	defer srv.Close()

	got := doJSON(t, srv.URL, "POST", "/v2/pet", `{"id":7,"name":"a","photoUrls":["x"]}`)
	if got["id"] != float64(7) {
		t.Errorf("addPet returned id=%v, want 7", got["id"])
	}
	got = doJSON(t, srv.URL, "GET", "/v2/pet/7", "")
	if got["name"] != "a" {
		t.Errorf("getPetById returned name=%v, want a", got["name"])
	}
}

func TestServer_GetMissingReturns404(t *testing.T) {
	srv := NewServer()
	defer srv.Close()
	resp := do(t, srv.URL, "GET", "/v2/pet/999", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServer_DeleteThenGet404(t *testing.T) {
	srv := NewServer()
	defer srv.Close()
	doJSON(t, srv.URL, "POST", "/v2/pet", `{"id":1,"name":"x","photoUrls":[]}`)
	got := doJSON(t, srv.URL, "DELETE", "/v2/pet/1", "")
	if got["message"] != "1" {
		t.Errorf("deletePet.message = %v, want \"1\" (id echoed back)", got["message"])
	}
	resp := do(t, srv.URL, "GET", "/v2/pet/1", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("after delete, GET status = %d, want 404", resp.StatusCode)
	}
}

func TestServer_FindByStatusReturnsArray(t *testing.T) {
	srv := NewServer()
	defer srv.Close()
	doJSON(t, srv.URL, "POST", "/v2/pet", `{"id":1,"name":"a","photoUrls":[],"status":"available"}`)
	doJSON(t, srv.URL, "POST", "/v2/pet", `{"id":2,"name":"b","photoUrls":[],"status":"sold"}`)

	resp := do(t, srv.URL, "GET", "/v2/pet/findByStatus?status=available", "")
	defer resp.Body.Close()
	var arr []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		t.Fatalf("decode array: %v", err)
	}
	if len(arr) != 1 || arr[0]["id"] != float64(1) {
		t.Errorf("findByStatus(available) = %+v, want single pet id=1", arr)
	}
}

func do(t *testing.T, base, method, path, body string) *http.Response {
	t.Helper()
	var b io.Reader
	if body != "" {
		b = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, base+path, b)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func doJSON(t *testing.T, base, method, path, body string) map[string]any {
	t.Helper()
	resp := do(t, base, method, path, body)
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)
	if resp.StatusCode >= 300 {
		t.Fatalf("%s %s -> %d: %s", method, path, resp.StatusCode, buf.String())
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("decode %s %s: %v (body=%q)", method, path, err, buf.String())
	}
	return out
}
