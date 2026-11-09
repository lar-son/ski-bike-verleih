package handler

import (
	"net/http"
	"strconv"
	"time"

	"skiverleih/internal/domain"
)

// skiView ergänzt einen Ski um den aktuellen Ausleihstatus.
type skiView struct {
	domain.Ski
	Ausgeliehen bool `json:"ausgeliehen"`
}

// skiOffer ist ein im Zeitraum freier Ski samt Preis (SPEC R15).
type skiOffer struct {
	domain.Ski
	PriceCent int `json:"priceCent"`
}

func (h *Handler) listSkis(w http.ResponseWriter, r *http.Request) {
	active := map[int]bool{}
	for _, rt := range h.Store.ListRentals() {
		if rt.Status == domain.StatusActive {
			active[rt.SkiID] = true
		}
	}
	skis := h.Store.ListSkis()
	out := make([]skiView, 0, len(skis))
	for _, s := range skis {
		out = append(out, skiView{Ski: s, Ausgeliehen: active[s.ID]})
	}
	writeJSON(w, http.StatusOK, out)
}

// availableSkis: GET /api/skis/available?start=<RFC3339>&hours=<n>
func (h *Handler) availableSkis(w http.ResponseWriter, r *http.Request) {
	start, err := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	if err != nil {
		writeErr(w, domain.Invalid("Parameter start fehlt oder ist ungültig (RFC 3339)"))
		return
	}
	hours, err := strconv.Atoi(r.URL.Query().Get("hours"))
	if err != nil || hours < 1 || hours > domain.MaxHours {
		writeErr(w, domain.Invalid("Dauer muss zwischen 1 und %d Stunden liegen", domain.MaxHours))
		return
	}
	start = start.Truncate(time.Hour)
	rentals := h.Store.ListRentals()
	now := h.Now()
	out := []skiOffer{}
	for _, s := range h.Store.ListSkis() {
		if domain.SkiIsFree(rentals, s.ID, start, hours, now) {
			out = append(out, skiOffer{Ski: s, PriceCent: s.Price(hours)})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) createSki(w http.ResponseWriter, r *http.Request) {
	var s domain.Ski
	if err := decode(r, &s); err != nil {
		writeErr(w, err)
		return
	}
	if err := s.Validate(); err != nil {
		writeErr(w, err)
		return
	}
	created, err := h.Store.CreateSki(s)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateSki(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var s domain.Ski
	if err := decode(r, &s); err != nil {
		writeErr(w, err)
		return
	}
	s.ID = id
	if err := s.Validate(); err != nil {
		writeErr(w, err)
		return
	}
	updated, err := h.Store.UpdateSki(s)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteSki(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := h.Store.DeleteSki(id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
