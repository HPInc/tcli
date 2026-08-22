// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// Package petstoretest provides an in-memory HTTP server that implements
// the subset of the Swagger Petstore API tcli's example pipelines exercise.
// Intended for CI integration tests replaces the external
// petstore.swagger.io dependency with a fast, deterministic local server.
package petstoretest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
)

// Pet mirrors the minimal shape addPet accepts and getPetById returns.
// Fields not used by tcli pipelines are not included.
type Pet struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	PhotoUrls []string `json:"photoUrls"`
	Status    string   `json:"status,omitempty"`
	Tags      []any    `json:"tags"`
	Category  any      `json:"category,omitempty"`
}

// apiResponse mirrors the real petstore's ApiResponse envelope, used for
// delete confirmation and 404 replies.
type apiResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewServer returns a running httptest.Server serving the pet endpoints
// under `/v2/`. Callers must call Close(). Each server has its own
// in-memory store, so tests are isolated from one another.
func NewServer() *httptest.Server {
	st := &store{pets: make(map[int64]*Pet)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v2/pet", st.addPet)
	mux.HandleFunc("GET /v2/pet/findByStatus", st.findByStatus)
	mux.HandleFunc("GET /v2/pet/{petId}", st.getPetByID)
	mux.HandleFunc("DELETE /v2/pet/{petId}", st.deletePet)
	return httptest.NewServer(mux)
}

type store struct {
	mu   sync.Mutex
	pets map[int64]*Pet
}

func (s *store) addPet(w http.ResponseWriter, r *http.Request) {
	var pet Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Default an empty status so findByStatus behaves predictably when the
	// caller omits it (the real petstore does the same coercion in practice).
	if pet.Status == "" {
		pet.Status = "available"
	}
	if pet.Tags == nil {
		pet.Tags = []any{}
	}
	s.mu.Lock()
	s.pets[pet.ID] = &pet
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, pet)
}

func (s *store) getPetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("petId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid petId", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	pet, ok := s.pets[id]
	s.mu.Unlock()
	if !ok {
		writeJSON(w, http.StatusNotFound, apiResponse{
			Code: 1, Type: "error", Message: "Pet not found",
		})
		return
	}
	writeJSON(w, http.StatusOK, pet)
}

// deletePet mirrors the real petstore's response: the `message` field
// carries the deleted id back, which the CRUD pipeline relies on for its
// `{petId:.message}` format expression to chain into verify_gone.
func (s *store) deletePet(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("petId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid petId", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	delete(s.pets, id)
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, apiResponse{
		Code: 200, Type: "unknown", Message: idStr,
	})
}

func (s *store) findByStatus(w http.ResponseWriter, r *http.Request) {
	want := r.URL.Query().Get("status")
	s.mu.Lock()
	// Always return [] (never null) so the pipeline runner's array-flatten
	// logic sees a well-formed JSON array even when no pets match.
	out := make([]*Pet, 0, len(s.pets))
	for _, p := range s.pets {
		if want == "" || p.Status == want {
			out = append(out, p)
		}
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
