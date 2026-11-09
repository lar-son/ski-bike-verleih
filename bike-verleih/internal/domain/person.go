package domain

import (
	"slices"
	"strings"
)

var AusweisTypen = []string{"Personalausweis", "Reisepass", "Führerschein", "Aufenthaltstitel"}

// Document ist ein hochgeladener Ausweis (Bild oder PDF). Die Bytes werden nie als JSON ausgegeben.
type Document struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Data        []byte `json:"-"`
}

// Person ist ein Kunde. Der Ausweis kann eingegeben (Art + Nummer) und/oder hochgeladen sein.
type Person struct {
	ID         int       `json:"id"`
	Vorname    string    `json:"vorname"`
	Nachname   string    `json:"nachname"`
	Email      string    `json:"email"`
	Telefon    string    `json:"telefon"`
	AusweisTyp string    `json:"ausweisTyp"`
	AusweisNr  string    `json:"ausweisNr"`
	Dokument   *Document `json:"dokument"`
}

func (p Person) Name() string { return p.Vorname + " " + p.Nachname }

// HasID: Ausweis vorhanden = Art + Nummer eingegeben ODER Dokument hochgeladen (SPEC R8).
func (p Person) HasID() bool {
	return (p.AusweisTyp != "" && p.AusweisNr != "") || p.Dokument != nil
}

// Validate normalisiert und prüft die Eingaben (SPEC R6, R7).
func (p *Person) Validate() error {
	p.Vorname = strings.TrimSpace(p.Vorname)
	p.Nachname = strings.TrimSpace(p.Nachname)
	p.Email = strings.TrimSpace(p.Email)
	p.Telefon = strings.TrimSpace(p.Telefon)
	p.AusweisTyp = strings.TrimSpace(p.AusweisTyp)
	p.AusweisNr = strings.TrimSpace(p.AusweisNr)
	switch {
	case p.Vorname == "":
		return Invalid("Vorname fehlt")
	case p.Nachname == "":
		return Invalid("Nachname fehlt")
	case p.Email != "" && !strings.Contains(p.Email, "@"):
		return Invalid("E-Mail-Adresse ist ungültig")
	case p.AusweisTyp != "" && !slices.Contains(AusweisTypen, p.AusweisTyp):
		return Invalid("Unbekannte Ausweisart %q", p.AusweisTyp)
	case p.AusweisTyp != "" && p.AusweisNr == "":
		return Invalid("Ausweisnummer fehlt")
	case p.AusweisTyp == "" && p.AusweisNr != "":
		return Invalid("Ausweisart fehlt")
	}
	return nil
}
