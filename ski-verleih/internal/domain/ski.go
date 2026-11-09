package domain

import (
	"fmt"
	"slices"
	"strings"
)

// MaxHours ist die maximale Verleihdauer (7 Tage).
const MaxHours = 24 * 7

var Kategorien = []string{"Carving", "Allmountain", "Freeride", "Race", "Kinder"}

// Ski ist ein Verleihgegenstand. Preise in Cent.
// PreisTagCent == 0 bedeutet: kein Tagespreis, es gilt nur der Stundenpreis.
type Ski struct {
	ID              int    `json:"id"`
	Marke           string `json:"marke"`
	Modell          string `json:"modell"`
	LaengeCm        int    `json:"laengeCm"`
	Kategorie       string `json:"kategorie"`
	PreisStundeCent int    `json:"preisStundeCent"`
	PreisTagCent    int    `json:"preisTagCent"`
}

// Name liefert eine kurze Anzeige, z.B. "Atomic Redster (165 cm)".
func (s Ski) Name() string {
	return fmt.Sprintf("%s %s (%d cm)", s.Marke, s.Modell, s.LaengeCm)
}

// Validate normalisiert und prüft die Eingaben (SPEC R5, R10).
func (s *Ski) Validate() error {
	s.Marke = strings.TrimSpace(s.Marke)
	s.Modell = strings.TrimSpace(s.Modell)
	if s.Kategorie == "" {
		s.Kategorie = "Allmountain"
	}
	switch {
	case s.Marke == "":
		return Invalid("Marke fehlt")
	case s.Modell == "":
		return Invalid("Modell fehlt")
	case s.LaengeCm < 50 || s.LaengeCm > 220:
		return Invalid("Länge muss zwischen 50 und 220 cm liegen")
	case !slices.Contains(Kategorien, s.Kategorie):
		return Invalid("Unbekannte Kategorie %q", s.Kategorie)
	case s.PreisStundeCent <= 0:
		return Invalid("Preis pro Stunde muss größer als 0 sein")
	case s.PreisTagCent < 0:
		return Invalid("Preis pro Tag darf nicht negativ sein")
	case s.PreisTagCent > 0 && s.PreisTagCent < s.PreisStundeCent:
		return Invalid("Tagespreis darf nicht unter dem Stundenpreis liegen")
	}
	return nil
}

// Price berechnet den Preis in Cent für die Dauer in Stunden (SPEC R11):
// volle 24-h-Blöcke zum Tagespreis, der Rest zum Stundenpreis, höchstens aber ein Tagespreis.
func (s Ski) Price(hours int) int {
	if hours <= 0 {
		return 0
	}
	if s.PreisTagCent <= 0 {
		return hours * s.PreisStundeCent
	}
	days, rest := hours/24, hours%24
	return days*s.PreisTagCent + min(rest*s.PreisStundeCent, s.PreisTagCent)
}
