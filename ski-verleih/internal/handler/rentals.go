package handler

import (
	"net/http"
	"time"

	"skiverleih/internal/domain"
)

// rentalView ergänzt eine Ausleihe um Anzeigenamen, Ende und Überfälligkeit.
type rentalView struct {
	domain.Rental
	SkiName    string    `json:"skiName"`
	PersonName string    `json:"personName"`
	End        time.Time `json:"end"`
	Overdue    bool      `json:"overdue"`
}

type createRentalRequest struct {
	SkiID    int       `json:"skiId"`
	PersonID int       `json:"personId"`
	Start    time.Time `json:"start"`
	Hours    int       `json:"hours"`
}

func (h *Handler) view(rt domain.Rental, now time.Time) rentalView {
	v := rentalView{Rental: rt, SkiName: "(gelöscht)", PersonName: "(gelöscht)", End: rt.End(), Overdue: rt.Overdue(now)}
	if s, err := h.Store.GetSki(rt.SkiID); err == nil {
		v.SkiName = s.Name()
	}
	if p, err := h.Store.GetPerson(rt.PersonID); err == nil {
		v.PersonName = p.Name()
	}
	return v
}

func (h *Handler) listRentals(w http.ResponseWriter, r *http.Request) {
	now := h.Now()
	rentals := h.Store.ListRentals()
	out := make([]rentalView, 0, len(rentals))
	for i := len(rentals) - 1; i >= 0; i-- { // neueste zuerst
		out = append(out, h.view(rentals[i], now))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) createRental(w http.ResponseWriter, r *http.Request) {
	var req createRentalRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	ski, err := h.Store.GetSki(req.SkiID)
	if err != nil {
		writeErr(w, err)
		return
	}
	person, err := h.Store.GetPerson(req.PersonID)
	if err != nil {
		writeErr(w, err)
		return
	}
	now := h.Now()
	rt, err := domain.NewRental(ski, person, req.Start, req.Hours, now)
	if err != nil {
		writeErr(w, err)
		return
	}
	created, err := h.Store.CreateRental(rt, now)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, h.view(created, now))
}

// rentalAction liefert einen Handler für pickup | return | cancel.
func (h *Handler) rentalAction(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			writeErr(w, err)
			return
		}
		now := h.Now()
		updated, err := h.Store.UpdateRental(id, func(rt *domain.Rental) error {
			switch action {
			case "pickup":
				return rt.Pickup(now)
			case "return":
				return rt.Return(now)
			default:
				return rt.Cancel()
			}
		})
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, h.view(updated, now))
	}
}
