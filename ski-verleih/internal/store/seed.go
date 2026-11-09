package store

import (
	"time"

	"skiverleih/internal/domain"
)

// Seed lädt Demodaten: 20 Ski, 10 Personen, 3 ausgeliehene Ski, 2 Reservierungen (SPEC R18).
func (m *MemoryStore) Seed(now time.Time) error {
	skis := []domain.Ski{
		{Marke: "Atomic", Modell: "Redster S9", LaengeCm: 165, Kategorie: "Carving", PreisStundeCent: 800, PreisTagCent: 3500},
		{Marke: "Atomic", Modell: "Maverick 86", LaengeCm: 170, Kategorie: "Allmountain", PreisStundeCent: 700, PreisTagCent: 3000},
		{Marke: "Atomic", Modell: "Bent 100", LaengeCm: 180, Kategorie: "Freeride", PreisStundeCent: 900, PreisTagCent: 4000},
		{Marke: "Fischer", Modell: "RC4 Pro", LaengeCm: 160, Kategorie: "Race", PreisStundeCent: 1000, PreisTagCent: 4500},
		{Marke: "Fischer", Modell: "Ranger 90", LaengeCm: 175, Kategorie: "Allmountain", PreisStundeCent: 700, PreisTagCent: 3000},
		{Marke: "Fischer", Modell: "Pro Mtn 86", LaengeCm: 170, Kategorie: "Carving", PreisStundeCent: 750, PreisTagCent: 3200},
		{Marke: "Head", Modell: "Supershape e-Magnum", LaengeCm: 170, Kategorie: "Carving", PreisStundeCent: 900, PreisTagCent: 3800},
		{Marke: "Head", Modell: "Kore 93", LaengeCm: 178, Kategorie: "Freeride", PreisStundeCent: 900, PreisTagCent: 4000},
		{Marke: "Head", Modell: "Worldcup Rebels", LaengeCm: 165, Kategorie: "Race", PreisStundeCent: 1100, PreisTagCent: 5000},
		{Marke: "Rossignol", Modell: "Experience 82", LaengeCm: 168, Kategorie: "Allmountain", PreisStundeCent: 650, PreisTagCent: 2800},
		{Marke: "Rossignol", Modell: "Forza 70", LaengeCm: 160, Kategorie: "Carving", PreisStundeCent: 700, PreisTagCent: 3000},
		{Marke: "Rossignol", Modell: "Sender 104", LaengeCm: 182, Kategorie: "Freeride", PreisStundeCent: 950, PreisTagCent: 4200},
		{Marke: "Völkl", Modell: "Deacon 80", LaengeCm: 172, Kategorie: "Allmountain", PreisStundeCent: 800, PreisTagCent: 3500},
		{Marke: "Völkl", Modell: "Racetiger SC", LaengeCm: 165, Kategorie: "Race", PreisStundeCent: 1000, PreisTagCent: 4500},
		{Marke: "Salomon", Modell: "S/Max 8", LaengeCm: 165, Kategorie: "Carving", PreisStundeCent: 650, PreisTagCent: 2800},
		{Marke: "Salomon", Modell: "QST 92", LaengeCm: 176, Kategorie: "Allmountain", PreisStundeCent: 750, PreisTagCent: 3200},
		{Marke: "Blizzard", Modell: "Brahma 82", LaengeCm: 170, Kategorie: "Allmountain", PreisStundeCent: 700, PreisTagCent: 3000},
		{Marke: "Blizzard", Modell: "Firebird Kids", LaengeCm: 120, Kategorie: "Kinder", PreisStundeCent: 400, PreisTagCent: 1800},
		{Marke: "Atomic", Modell: "Redster J2", LaengeCm: 130, Kategorie: "Kinder", PreisStundeCent: 400, PreisTagCent: 1800},
		{Marke: "Head", Modell: "Monster Kids", LaengeCm: 110, Kategorie: "Kinder", PreisStundeCent: 350, PreisTagCent: 1600},
	}
	for _, s := range skis {
		if _, err := m.CreateSki(s); err != nil {
			return err
		}
	}

	persons := []domain.Person{
		{Vorname: "Anna", Nachname: "Gruber", Email: "anna.gruber@example.com", Telefon: "+43 660 1111111", AusweisTyp: "Personalausweis", AusweisNr: "A1234567"},
		{Vorname: "Lukas", Nachname: "Huber", Email: "lukas.huber@example.com", Telefon: "+43 660 2222222", AusweisTyp: "Reisepass", AusweisNr: "P7654321"},
		{Vorname: "Sophie", Nachname: "Bauer", Email: "sophie.bauer@example.com", Telefon: "+43 660 3333333", AusweisTyp: "Führerschein", AusweisNr: "F5551234"},
		{Vorname: "Felix", Nachname: "Wagner", Email: "felix.wagner@example.com", Telefon: "+43 660 4444444", AusweisTyp: "Personalausweis", AusweisNr: "A2345678"},
		{Vorname: "Marie", Nachname: "Pichler", Email: "marie.pichler@example.com", Telefon: "+43 660 5555555", AusweisTyp: "Reisepass", AusweisNr: "P1122334"},
		{Vorname: "Jonas", Nachname: "Steiner", Email: "jonas.steiner@example.com", Telefon: "+43 660 6666666", AusweisTyp: "Personalausweis", AusweisNr: "A3456789"},
		{Vorname: "Laura", Nachname: "Moser", Email: "laura.moser@example.com", Telefon: "+43 660 7777777", AusweisTyp: "Führerschein", AusweisNr: "F9998765"},
		{Vorname: "David", Nachname: "Fischer", Email: "david.fischer@example.com", Telefon: "+43 660 8888888", AusweisTyp: "Personalausweis", AusweisNr: "A4567890"},
		{Vorname: "Julia", Nachname: "Mayer", Email: "julia.mayer@example.com", Telefon: "+43 660 9999999"}, // bewusst ohne Ausweis (Demo R8)
		{Vorname: "Tobias", Nachname: "Leitner", Email: "tobias.leitner@example.com", Telefon: "+43 660 1212121", AusweisTyp: "Reisepass", AusweisNr: "P4455667"},
	}
	for _, p := range persons {
		if _, err := m.CreatePerson(p); err != nil {
			return err
		}
	}

	hour := now.Truncate(time.Hour)
	tomorrow9 := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())
	type seedRental struct {
		skiID, personID, hours int
		start                  time.Time
	}
	for _, sr := range []seedRental{
		{1, 1, 4, hour.Add(-1 * time.Hour)}, // läuft noch 3 h
		{5, 2, 8, hour.Add(-2 * time.Hour)},
		{12, 3, 3, hour},
		{3, 4, 6, tomorrow9},  // Reservierung
		{8, 5, 24, tomorrow9}, // Reservierung
	} {
		ski, err := m.GetSki(sr.skiID)
		if err != nil {
			return err
		}
		person, err := m.GetPerson(sr.personID)
		if err != nil {
			return err
		}
		r := domain.Rental{
			SkiID: ski.ID, PersonID: person.ID, Start: sr.start, Hours: sr.hours,
			PriceCent: ski.Price(sr.hours), Status: domain.StatusActive,
		}
		if sr.start.After(now) {
			r.Status = domain.StatusReserved
		}
		if _, err := m.CreateRental(r, now); err != nil {
			return err
		}
	}
	return nil
}
