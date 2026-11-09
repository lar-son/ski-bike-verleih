package domain

import "time"

type Status string

const (
	StatusReserved  Status = "reserviert"
	StatusActive    Status = "ausgeliehen"
	StatusReturned  Status = "zurückgegeben"
	StatusCancelled Status = "storniert"
)

// Rental ist eine Ausleihe/Reservierung: ein Ski, eine Person, Beginn + Dauer in Stunden.
// Der Preis wird beim Anlegen fixiert (SPEC R10).
type Rental struct {
	ID         int        `json:"id"`
	SkiID      int        `json:"skiId"`
	PersonID   int        `json:"personId"`
	Start      time.Time  `json:"start"`
	Hours      int        `json:"hours"`
	PriceCent  int        `json:"priceCent"`
	Status     Status     `json:"status"`
	ReturnedAt *time.Time `json:"returnedAt,omitempty"`
}

func (r Rental) End() time.Time { return r.Start.Add(time.Duration(r.Hours) * time.Hour) }

// Blocks: reservierte und ausgeliehene Ski sind belegt.
func (r Rental) Blocks() bool { return r.Status == StatusReserved || r.Status == StatusActive }

// Overdue: ausgeliehen und Ende überschritten.
func (r Rental) Overdue(now time.Time) bool { return r.Status == StatusActive && now.After(r.End()) }

// overlaps prüft Überschneidung mit [start, end). Ein überfälliger Ski bleibt bis zur Rückgabe belegt (SPEC R14).
func (r Rental) overlaps(start, end, now time.Time) bool {
	if !r.Blocks() {
		return false
	}
	until := r.End()
	if r.Overdue(now) {
		until = now.AddDate(10, 0, 0)
	}
	return start.Before(until) && r.Start.Before(end)
}

// SkiIsFree sagt, ob der Ski im Zeitraum [start, start+hours) frei ist (SPEC R14).
func SkiIsFree(rentals []Rental, skiID int, start time.Time, hours int, now time.Time) bool {
	end := start.Add(time.Duration(hours) * time.Hour)
	for _, r := range rentals {
		if r.SkiID == skiID && r.overlaps(start, end, now) {
			return false
		}
	}
	return true
}

// NewRental prüft die Regeln R8, R12, R13 und legt eine Ausleihe an (ohne ID).
// Beginn in der laufenden Stunde ⇒ sofort ausgeliehen, später ⇒ reserviert.
func NewRental(ski Ski, person Person, start time.Time, hours int, now time.Time) (Rental, error) {
	start = start.Truncate(time.Hour)
	switch {
	case hours < 1 || hours > MaxHours:
		return Rental{}, Invalid("Dauer muss zwischen 1 und %d Stunden liegen", MaxHours)
	case start.Before(now.Truncate(time.Hour)):
		return Rental{}, Invalid("Der Beginn liegt in der Vergangenheit")
	case !person.HasID():
		return Rental{}, Invalid("Für %s fehlt der Ausweis (Art + Nummer eingeben oder Datei hochladen)", person.Name())
	}
	status := StatusReserved
	if !start.After(now) {
		status = StatusActive
	}
	return Rental{
		SkiID:     ski.ID,
		PersonID:  person.ID,
		Start:     start,
		Hours:     hours,
		PriceCent: ski.Price(hours),
		Status:    status,
	}, nil
}

// Pickup gibt eine Reservierung aus (erst ab Beginn, SPEC R16).
func (r *Rental) Pickup(now time.Time) error {
	if r.Status != StatusReserved {
		return Conflict("Nur reservierte Ausleihen können ausgegeben werden")
	}
	if r.Start.After(now) {
		return Conflict("Der Beginn ist noch nicht erreicht")
	}
	r.Status = StatusActive
	return nil
}

// Return nimmt einen ausgeliehenen Ski zurück.
func (r *Rental) Return(now time.Time) error {
	if r.Status != StatusActive {
		return Conflict("Nur ausgeliehene Ski können zurückgegeben werden")
	}
	r.Status = StatusReturned
	r.ReturnedAt = &now
	return nil
}

// Cancel storniert eine Reservierung.
func (r *Rental) Cancel() error {
	if r.Status != StatusReserved {
		return Conflict("Nur Reservierungen können storniert werden")
	}
	r.Status = StatusCancelled
	return nil
}
