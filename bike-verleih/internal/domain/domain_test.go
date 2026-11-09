package domain

import (
	"errors"
	"testing"
	"time"
)

var t0 = time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC)

func TestBikePrice(t *testing.T) { // SPEC R11
	s := Bike{PreisStundeCent: 800, PreisTagCent: 3500}
	cases := []struct {
		hours, want int
	}{
		{1, 800},
		{4, 3200},
		{5, 3500},         // gedeckelt auf Tagespreis
		{24, 3500},        // genau ein Tag
		{25, 3500 + 800},  // Tag + 1 h
		{29, 3500 + 3500}, // Tag + gedeckelter Rest
		{48, 7000},
		{0, 0},
	}
	for _, c := range cases {
		if got := s.Price(c.hours); got != c.want {
			t.Errorf("Price(%d) = %d, want %d", c.hours, got, c.want)
		}
	}
	noDay := Bike{PreisStundeCent: 500}
	if got := noDay.Price(30); got != 15000 {
		t.Errorf("ohne Tagespreis: %d, want 15000", got)
	}
}

func TestBikeValidate(t *testing.T) {
	ok := Bike{Marke: "Atomic", Modell: "X", RahmenCm: 54, PreisStundeCent: 500}
	if err := ok.Validate(); err != nil {
		t.Fatalf("gültiger Bike: %v", err)
	}
	bad := []Bike{
		{Modell: "X", RahmenCm: 54, PreisStundeCent: 500},
		{Marke: "A", Modell: "X", RahmenCm: 10, PreisStundeCent: 500},
		{Marke: "A", Modell: "X", RahmenCm: 54, PreisStundeCent: 0},
		{Marke: "A", Modell: "X", RahmenCm: 54, PreisStundeCent: 500, PreisTagCent: 100},
		{Marke: "A", Modell: "X", RahmenCm: 54, PreisStundeCent: 500, Kategorie: "Unsinn"},
	}
	for i, s := range bad {
		if err := s.Validate(); !errors.Is(err, ErrInvalid) {
			t.Errorf("Fall %d: erwartet ErrInvalid, got %v", i, err)
		}
	}
}

func TestPersonID(t *testing.T) { // SPEC R7, R8
	p := Person{Vorname: "A", Nachname: "B"}
	if p.HasID() {
		t.Error("ohne Ausweis darf HasID false sein")
	}
	p.AusweisTyp, p.AusweisNr = "Reisepass", "P1"
	if !p.HasID() {
		t.Error("Art + Nummer reicht")
	}
	q := Person{Vorname: "A", Nachname: "B", Dokument: &Document{}}
	if !q.HasID() {
		t.Error("Upload reicht")
	}
	half := Person{Vorname: "A", Nachname: "B", AusweisNr: "123"}
	if err := half.Validate(); !errors.Is(err, ErrInvalid) {
		t.Errorf("Nummer ohne Art: %v", err)
	}
}

func TestBikeIsFree(t *testing.T) { // SPEC R14
	r := Rental{ID: 1, BikeID: 1, Start: t0, Hours: 2, Status: StatusReserved}
	rentals := []Rental{r}
	now := t0.Add(-time.Hour)
	if BikeIsFree(rentals, 1, t0.Add(time.Hour), 2, now) {
		t.Error("Überschneidung muss blockieren")
	}
	if !BikeIsFree(rentals, 1, t0.Add(2*time.Hour), 1, now) {
		t.Error("direkt anschließend muss frei sein")
	}
	if !BikeIsFree(rentals, 1, t0.Add(-2*time.Hour), 2, now) {
		t.Error("direkt davor muss frei sein")
	}
	if !BikeIsFree(rentals, 2, t0, 2, now) {
		t.Error("anderer Bike muss frei sein")
	}
	rentals[0].Status = StatusReturned
	if !BikeIsFree(rentals, 1, t0, 2, now) {
		t.Error("zurückgegeben blockiert nicht")
	}
}

func TestOverdueStaysBlocked(t *testing.T) {
	r := Rental{BikeID: 1, Start: t0, Hours: 1, Status: StatusActive}
	now := t0.Add(5 * time.Hour)
	if BikeIsFree([]Rental{r}, 1, now, 1, now) {
		t.Error("überfälliger Bike muss belegt bleiben")
	}
}

func TestNewRental(t *testing.T) { // SPEC R8, R12, R13
	bike := Bike{ID: 1, PreisStundeCent: 800, PreisTagCent: 3500}
	person := Person{ID: 1, Vorname: "A", Nachname: "B", AusweisTyp: "Reisepass", AusweisNr: "1"}
	now := t0.Add(5 * time.Minute) // 10:05

	r, err := NewRental(bike, person, t0, 3, now)
	if err != nil || r.Status != StatusActive || r.PriceCent != 2400 {
		t.Errorf("laufende Stunde: %+v, %v", r, err)
	}
	r, err = NewRental(bike, person, t0.Add(3*time.Hour), 2, now)
	if err != nil || r.Status != StatusReserved {
		t.Errorf("Zukunft: %+v, %v", r, err)
	}
	r, _ = NewRental(bike, person, t0.Add(3*time.Hour+30*time.Minute), 1, now)
	if !r.Start.Equal(t0.Add(3 * time.Hour)) {
		t.Errorf("Beginn muss auf volle Stunde fallen: %v", r.Start)
	}
	if _, err = NewRental(bike, person, t0.Add(-time.Hour), 1, now); !errors.Is(err, ErrInvalid) {
		t.Errorf("Vergangenheit: %v", err)
	}
	for _, h := range []int{0, MaxHours + 1} {
		if _, err = NewRental(bike, person, t0, h, now); !errors.Is(err, ErrInvalid) {
			t.Errorf("Dauer %d: %v", h, err)
		}
	}
	if _, err = NewRental(bike, Person{ID: 2, Vorname: "X", Nachname: "Y"}, t0, 1, now); !errors.Is(err, ErrInvalid) {
		t.Errorf("ohne Ausweis: %v", err)
	}
}

func TestRentalTransitions(t *testing.T) { // SPEC R16
	now := t0
	r := Rental{Status: StatusReserved, Start: t0.Add(time.Hour), Hours: 1}
	if err := r.Pickup(now); !errors.Is(err, ErrConflict) {
		t.Errorf("Pickup vor Beginn: %v", err)
	}
	if err := r.Pickup(now.Add(time.Hour)); err != nil || r.Status != StatusActive {
		t.Errorf("Pickup: %v %v", err, r.Status)
	}
	if err := r.Cancel(); !errors.Is(err, ErrConflict) {
		t.Errorf("Cancel nach Pickup: %v", err)
	}
	if err := r.Return(now.Add(2 * time.Hour)); err != nil || r.Status != StatusReturned || r.ReturnedAt == nil {
		t.Errorf("Return: %v %+v", err, r)
	}
}
