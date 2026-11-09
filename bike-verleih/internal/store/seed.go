package store

import (
	"time"

	"bikeverleih/internal/domain"
)

// Seed lädt Demodaten: 20 Bike, 10 Personen, 3 ausgeliehene Bike, 2 Reservierungen (SPEC R18).
func (m *MemoryStore) Seed(now time.Time) error {
	bikes := []domain.Bike{
		{Marke: "Cube", Modell: "Touring Hybrid", Kategorie: "Trekking", RahmenCm: 56, PreisStundeCent: 500, PreisTagCent: 2200},
		{Marke: "Cube", Modell: "Nature EXC", Kategorie: "Trekking", RahmenCm: 54, PreisStundeCent: 500, PreisTagCent: 2200},
		{Marke: "Kalkhoff", Modell: "Image 5", Kategorie: "Trekking", RahmenCm: 52, PreisStundeCent: 550, PreisTagCent: 2400},
		{Marke: "Gudereit", Modell: "Comfort 7.0", Kategorie: "City", RahmenCm: 50, PreisStundeCent: 400, PreisTagCent: 1800},
		{Marke: "Gudereit", Modell: "Comfort 7.0", Kategorie: "City", RahmenCm: 54, PreisStundeCent: 400, PreisTagCent: 1800},
		{Marke: "Pegasus", Modell: "Solero 8", Kategorie: "City", RahmenCm: 48, PreisStundeCent: 400, PreisTagCent: 1800},
		{Marke: "Pegasus", Modell: "Premio", Kategorie: "City", RahmenCm: 52, PreisStundeCent: 450, PreisTagCent: 2000},
		{Marke: "KTM", Modell: "Life Joy", Kategorie: "City", RahmenCm: 56, PreisStundeCent: 450, PreisTagCent: 2000},
		{Marke: "Scott", Modell: "Aspect 950", Kategorie: "Mountainbike", RahmenCm: 48, PreisStundeCent: 700, PreisTagCent: 3000},
		{Marke: "Scott", Modell: "Scale 970", Kategorie: "Mountainbike", RahmenCm: 52, PreisStundeCent: 800, PreisTagCent: 3500},
		{Marke: "Cannondale", Modell: "Trail 6", Kategorie: "Mountainbike", RahmenCm: 56, PreisStundeCent: 750, PreisTagCent: 3200},
		{Marke: "Specialized", Modell: "Rockhopper", Kategorie: "Mountainbike", RahmenCm: 50, PreisStundeCent: 750, PreisTagCent: 3200},
		{Marke: "Trek", Modell: "Marlin 7", Kategorie: "Mountainbike", RahmenCm: 54, PreisStundeCent: 700, PreisTagCent: 3000},
		{Marke: "Haibike", Modell: "Trekking 5", Kategorie: "E-Bike", RahmenCm: 52, PreisStundeCent: 1200, PreisTagCent: 5500},
		{Marke: "Cube", Modell: "Kathmandu Hybrid", Kategorie: "E-Bike", RahmenCm: 56, PreisStundeCent: 1300, PreisTagCent: 6000},
		{Marke: "KTM", Modell: "Macina Style", Kategorie: "E-Bike", RahmenCm: 50, PreisStundeCent: 1300, PreisTagCent: 6000},
		{Marke: "Bergamont", Modell: "E-Horizon", Kategorie: "E-Bike", RahmenCm: 54, PreisStundeCent: 1250, PreisTagCent: 5800},
		{Marke: "Puky", Modell: "Youster 20", Kategorie: "Kinder", RahmenCm: 30, PreisStundeCent: 300, PreisTagCent: 1200},
		{Marke: "Woom", Modell: "Original 4", Kategorie: "Kinder", RahmenCm: 34, PreisStundeCent: 350, PreisTagCent: 1400},
		{Marke: "Puky", Modell: "Skyride 24", Kategorie: "Kinder", RahmenCm: 38, PreisStundeCent: 350, PreisTagCent: 1400},
	}
	for _, s := range bikes {
		if _, err := m.CreateBike(s); err != nil {
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
		bikeID, personID, hours int
		start                   time.Time
	}
	for _, sr := range []seedRental{
		{1, 1, 4, hour.Add(-1 * time.Hour)}, // läuft noch 3 h
		{5, 2, 8, hour.Add(-2 * time.Hour)},
		{12, 3, 3, hour},
		{3, 4, 6, tomorrow9},  // Reservierung
		{8, 5, 24, tomorrow9}, // Reservierung
	} {
		bike, err := m.GetBike(sr.bikeID)
		if err != nil {
			return err
		}
		person, err := m.GetPerson(sr.personID)
		if err != nil {
			return err
		}
		r := domain.Rental{
			BikeID: bike.ID, PersonID: person.ID, Start: sr.start, Hours: sr.hours,
			PriceCent: bike.Price(sr.hours), Status: domain.StatusActive,
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
