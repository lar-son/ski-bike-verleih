// Package store hält alle Daten im Speicher (Prototyp, keine Datenbank).
package store

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"skiverleih/internal/domain"
)

// Stats sind die Kennzahlen für das Dashboard (SPEC R1, R2).
type Stats struct {
	Verfuegbar   int `json:"verfuegbar"`
	Ausgeliehen  int `json:"ausgeliehen"`
	Gesamt       int `json:"gesamt"`
	Personen     int `json:"personen"`
	Reserviert   int `json:"reserviert"`
	Ueberfaellig int `json:"ueberfaellig"`
}

// Store ist die Schnittstelle zum Datenspeicher; später z.B. durch eine DB ersetzbar.
type Store interface {
	ListSkis() []domain.Ski
	GetSki(id int) (domain.Ski, error)
	CreateSki(s domain.Ski) (domain.Ski, error)
	UpdateSki(s domain.Ski) (domain.Ski, error)
	DeleteSki(id int) error

	ListPersons() []domain.Person
	GetPerson(id int) (domain.Person, error)
	CreatePerson(p domain.Person) (domain.Person, error)
	UpdatePerson(p domain.Person) (domain.Person, error)
	SetDocument(personID int, d *domain.Document) (domain.Person, error)
	DeletePerson(id int) error

	ListRentals() []domain.Rental
	CreateRental(r domain.Rental, now time.Time) (domain.Rental, error)
	UpdateRental(id int, change func(*domain.Rental) error) (domain.Rental, error)

	Stats(now time.Time) Stats
}

// MemoryStore ist ein thread-safe In-Memory-Store (Maps + RWMutex).
type MemoryStore struct {
	mu         sync.RWMutex
	skis       map[int]domain.Ski
	persons    map[int]domain.Person
	rentals    map[int]domain.Rental
	nextSki    int
	nextPerson int
	nextRental int
}

var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		skis:       map[int]domain.Ski{},
		persons:    map[int]domain.Person{},
		rentals:    map[int]domain.Rental{},
		nextSki:    1,
		nextPerson: 1,
		nextRental: 1,
	}
}

// --- Ski ---

func (m *MemoryStore) ListSkis() []domain.Ski {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Ski, 0, len(m.skis))
	for _, s := range m.skis {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (m *MemoryStore) GetSki(id int) (domain.Ski, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.skis[id]
	if !ok {
		return domain.Ski{}, domain.NotFound("Ski", id)
	}
	return s, nil
}

func (m *MemoryStore) CreateSki(s domain.Ski) (domain.Ski, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s.ID = m.nextSki
	m.nextSki++
	m.skis[s.ID] = s
	return s, nil
}

func (m *MemoryStore) UpdateSki(s domain.Ski) (domain.Ski, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.skis[s.ID]; !ok {
		return domain.Ski{}, domain.NotFound("Ski", s.ID)
	}
	m.skis[s.ID] = s
	return s, nil
}

func (m *MemoryStore) DeleteSki(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.skis[id]; !ok {
		return domain.NotFound("Ski", id)
	}
	for _, r := range m.rentals {
		if r.SkiID == id && r.Blocks() {
			return domain.Conflict("Ski hat noch eine offene Ausleihe oder Reservierung")
		}
	}
	delete(m.skis, id)
	return nil
}

// --- Personen ---

func (m *MemoryStore) ListPersons() []domain.Person {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Person, 0, len(m.persons))
	for _, p := range m.persons {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (m *MemoryStore) GetPerson(id int) (domain.Person, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.persons[id]
	if !ok {
		return domain.Person{}, domain.NotFound("Person", id)
	}
	return p, nil
}

// CreatePerson ignoriert ein mitgeschicktes Dokument; Uploads laufen über SetDocument.
func (m *MemoryStore) CreatePerson(p domain.Person) (domain.Person, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.ID = m.nextPerson
	m.nextPerson++
	p.Dokument = nil
	m.persons[p.ID] = p
	return p, nil
}

// UpdatePerson behält das vorhandene Dokument.
func (m *MemoryStore) UpdatePerson(p domain.Person) (domain.Person, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.persons[p.ID]
	if !ok {
		return domain.Person{}, domain.NotFound("Person", p.ID)
	}
	p.Dokument = old.Dokument
	m.persons[p.ID] = p
	return p, nil
}

// SetDocument setzt oder entfernt (d == nil) den hochgeladenen Ausweis.
func (m *MemoryStore) SetDocument(personID int, d *domain.Document) (domain.Person, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.persons[personID]
	if !ok {
		return domain.Person{}, domain.NotFound("Person", personID)
	}
	p.Dokument = d
	m.persons[personID] = p
	return p, nil
}

func (m *MemoryStore) DeletePerson(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.persons[id]; !ok {
		return domain.NotFound("Person", id)
	}
	for _, r := range m.rentals {
		if r.PersonID == id && r.Blocks() {
			return domain.Conflict("Person hat noch eine offene Ausleihe oder Reservierung")
		}
	}
	delete(m.persons, id)
	return nil
}

// --- Ausleihen ---

func (m *MemoryStore) ListRentals() []domain.Rental {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rentalsLocked()
}

func (m *MemoryStore) rentalsLocked() []domain.Rental {
	out := make([]domain.Rental, 0, len(m.rentals))
	for _, r := range m.rentals {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// CreateRental prüft Existenz und Verfügbarkeit atomar unter dem Schreib-Lock (SPEC R14).
func (m *MemoryStore) CreateRental(r domain.Rental, now time.Time) (domain.Rental, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.skis[r.SkiID]; !ok {
		return domain.Rental{}, domain.NotFound("Ski", r.SkiID)
	}
	if _, ok := m.persons[r.PersonID]; !ok {
		return domain.Rental{}, domain.NotFound("Person", r.PersonID)
	}
	if !domain.SkiIsFree(m.rentalsLocked(), r.SkiID, r.Start, r.Hours, now) {
		return domain.Rental{}, domain.Conflict("Der Ski ist im gewählten Zeitraum nicht verfügbar")
	}
	r.ID = m.nextRental
	m.nextRental++
	m.rentals[r.ID] = r
	return r, nil
}

// UpdateRental wendet eine Zustandsänderung (Pickup/Return/Cancel) unter Lock an.
func (m *MemoryStore) UpdateRental(id int, change func(*domain.Rental) error) (domain.Rental, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rentals[id]
	if !ok {
		return domain.Rental{}, domain.NotFound("Ausleihe", id)
	}
	if err := change(&r); err != nil {
		return domain.Rental{}, fmt.Errorf("ausleihe %d: %w", id, err)
	}
	m.rentals[id] = r
	return r, nil
}

// --- Dashboard ---

func (m *MemoryStore) Stats(now time.Time) Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	active := map[int]bool{}
	st := Stats{Gesamt: len(m.skis), Personen: len(m.persons)}
	for _, r := range m.rentals {
		switch r.Status {
		case domain.StatusActive:
			active[r.SkiID] = true
			if r.Overdue(now) {
				st.Ueberfaellig++
			}
		case domain.StatusReserved:
			st.Reserviert++
		}
	}
	st.Ausgeliehen = len(active)
	st.Verfuegbar = st.Gesamt - st.Ausgeliehen
	return st
}
