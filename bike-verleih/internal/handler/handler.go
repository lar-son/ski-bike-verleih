// Package handler enthält die HTTP-Handler (JSON-API unter /api).
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"bikeverleih/internal/domain"
	"bikeverleih/internal/store"
)

type Handler struct {
	Store store.Store
	Now   func() time.Time // austauschbar für Tests
}

func New(s store.Store) *Handler {
	return &Handler{Store: s, Now: time.Now}
}

// Routes baut den Router; static liefert das Frontend unter "/".
func (h *Handler) Routes(static http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/stats", h.stats)

	mux.HandleFunc("GET /api/bikes", h.listBikes)
	mux.HandleFunc("GET /api/bikes/available", h.availableBikes)
	mux.HandleFunc("POST /api/bikes", h.createBike)
	mux.HandleFunc("PUT /api/bikes/{id}", h.updateBike)
	mux.HandleFunc("DELETE /api/bikes/{id}", h.deleteBike)

	mux.HandleFunc("GET /api/persons", h.listPersons)
	mux.HandleFunc("POST /api/persons", h.createPerson)
	mux.HandleFunc("PUT /api/persons/{id}", h.updatePerson)
	mux.HandleFunc("DELETE /api/persons/{id}", h.deletePerson)
	mux.HandleFunc("GET /api/persons/{id}/document", h.getDocument)
	mux.HandleFunc("POST /api/persons/{id}/document", h.uploadDocument)
	mux.HandleFunc("DELETE /api/persons/{id}/document", h.deleteDocument)

	mux.HandleFunc("GET /api/rentals", h.listRentals)
	mux.HandleFunc("POST /api/rentals", h.createRental)
	mux.HandleFunc("POST /api/rentals/{id}/pickup", h.rentalAction("pickup"))
	mux.HandleFunc("POST /api/rentals/{id}/return", h.rentalAction("return"))
	mux.HandleFunc("POST /api/rentals/{id}/cancel", h.rentalAction("cancel"))

	mux.Handle("/", static)
	return mux
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.Store.Stats(h.Now()))
}

// --- Helfer ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode: %v", err)
	}
}

// writeErr mappt Domain-Fehler auf HTTP-Statuscodes.
func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
	default:
		log.Printf("interner Fehler: %v", err)
	}
	msg := err.Error()
	var de *domain.Error
	if errors.As(err, &de) {
		msg = de.Msg // ohne Wrapping-Präfixe, nutzerfreundlich
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return domain.Invalid("Ungültige JSON-Daten: %v", err)
	}
	return nil
}

func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, domain.Invalid("Ungültige ID")
	}
	return id, nil
}
