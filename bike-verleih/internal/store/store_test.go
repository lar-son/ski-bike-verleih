package store

import (
	"errors"
	"sync"
	"testing"
	"time"

	"bikeverleih/internal/domain"
)

func TestSeedStats(t *testing.T) { // SPEC R2, R18
	m := NewMemoryStore()
	now := time.Date(2026, 1, 10, 10, 30, 0, 0, time.UTC)
	if err := m.Seed(now); err != nil {
		t.Fatal(err)
	}
	st := m.Stats(now)
	if st.Gesamt != 20 || st.Ausgeliehen != 3 || st.Verfuegbar != 17 || st.Personen != 10 {
		t.Errorf("Stats = %+v", st)
	}
}

// 20 parallele Anfragen auf denselben Bike im selben Zeitraum: genau eine gelingt (SPEC R14).
func TestConcurrentCreateRental(t *testing.T) {
	m := NewMemoryStore()
	bike, _ := m.CreateBike(domain.Bike{Marke: "A", Modell: "B", RahmenCm: 54, PreisStundeCent: 500})
	p, _ := m.CreatePerson(domain.Person{Vorname: "A", Nachname: "B", AusweisTyp: "Reisepass", AusweisNr: "1"})
	now := time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, conflicts := 0, 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := domain.Rental{BikeID: bike.ID, PersonID: p.ID, Start: now, Hours: 2, Status: domain.StatusActive}
			_, err := m.CreateRental(r, now)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case errors.Is(err, domain.ErrConflict):
				conflicts++
			}
		}()
	}
	wg.Wait()
	if ok != 1 || conflicts != 19 {
		t.Errorf("ok=%d conflicts=%d", ok, conflicts)
	}
}

func TestDeleteBlockedByOpenRental(t *testing.T) {
	m := NewMemoryStore()
	bike, _ := m.CreateBike(domain.Bike{Marke: "A", Modell: "B", RahmenCm: 54, PreisStundeCent: 500})
	p, _ := m.CreatePerson(domain.Person{Vorname: "A", Nachname: "B"})
	now := time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC)
	r, err := m.CreateRental(domain.Rental{BikeID: bike.ID, PersonID: p.ID, Start: now, Hours: 1, Status: domain.StatusActive}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteBike(bike.ID); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("DeleteBike: %v", err)
	}
	if err := m.DeletePerson(p.ID); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("DeletePerson: %v", err)
	}
	if _, err := m.UpdateRental(r.ID, func(r *domain.Rental) error { return r.Return(now) }); err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteBike(bike.ID); err != nil {
		t.Errorf("nach Rückgabe löschbar: %v", err)
	}
}

func TestUpdatePersonKeepsDocument(t *testing.T) {
	m := NewMemoryStore()
	p, _ := m.CreatePerson(domain.Person{Vorname: "A", Nachname: "B"})
	if _, err := m.SetDocument(p.ID, &domain.Document{FileName: "x.pdf", ContentType: "application/pdf", Data: []byte("%PDF")}); err != nil {
		t.Fatal(err)
	}
	p.Vorname = "Neu"
	p.Dokument = nil // Client schickt kein Dokument mit
	got, err := m.UpdatePerson(p)
	if err != nil || got.Dokument == nil {
		t.Errorf("Dokument muss erhalten bleiben: %+v %v", got, err)
	}
}
