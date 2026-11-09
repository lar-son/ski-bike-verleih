package handler

import (
	"net/http"
	"strconv"
	"time"

	"bikeverleih/internal/domain"
)

// bikeView ergänzt einen Bike um den aktuellen Ausleihstatus.
type bikeView struct {
	domain.Bike
	Ausgeliehen bool `json:"ausgeliehen"`
}

// bikeOffer ist ein im Zeitraum freier Bike samt Preis (SPEC R15).
type bikeOffer struct {
	domain.Bike
	PriceCent int `json:"priceCent"`
}

func (h *Handler) listBikes(w http.ResponseWriter, r *http.Request) {
	active := map[int]bool{}
	for _, rt := range h.Store.ListRentals() {
		if rt.Status == domain.StatusActive {
			active[rt.BikeID] = true
		}
	}
	bikes := h.Store.ListBikes()
	out := make([]bikeView, 0, len(bikes))
	for _, s := range bikes {
		out = append(out, bikeView{Bike: s, Ausgeliehen: active[s.ID]})
	}
	writeJSON(w, http.StatusOK, out)
}

// availableBikes: GET /api/bikes/available?start=<RFC3339>&hours=<n>
func (h *Handler) availableBikes(w http.ResponseWriter, r *http.Request) {
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
	out := []bikeOffer{}
	for _, s := range h.Store.ListBikes() {
		if domain.BikeIsFree(rentals, s.ID, start, hours, now) {
			out = append(out, bikeOffer{Bike: s, PriceCent: s.Price(hours)})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) createBike(w http.ResponseWriter, r *http.Request) {
	var s domain.Bike
	if err := decode(r, &s); err != nil {
		writeErr(w, err)
		return
	}
	if err := s.Validate(); err != nil {
		writeErr(w, err)
		return
	}
	created, err := h.Store.CreateBike(s)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateBike(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var s domain.Bike
	if err := decode(r, &s); err != nil {
		writeErr(w, err)
		return
	}
	s.ID = id
	if err := s.Validate(); err != nil {
		writeErr(w, err)
		return
	}
	updated, err := h.Store.UpdateBike(s)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteBike(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := h.Store.DeleteBike(id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
