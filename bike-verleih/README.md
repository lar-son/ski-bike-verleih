# Fahrrad-Verleih

Verleihsystem für Fahrräder – Aufgabe 1 c), entwickelt mit Spec-Driven Development
(`SPEC.md` → `CLAUDE.md` → Code + Tests).

- **Backend:** Go (nur Standardbibliothek), In-Memory-Daten, JSON-API
- **Frontend:** HTML/JS ohne Framework, vom Go-Server ausgeliefert
- **Sprache der Oberfläche:** Deutsch

## Starten

```bash
go run .                 # http://localhost:8080 (mit Demodaten)
go run . -addr :9000     # anderer Port
go run . -seed=false     # leer starten
go test ./...            # Tests
```

Benötigt Go ≥ 1.22. Alle Daten liegen nur im Speicher und sind nach dem Neustart weg.

## Funktionen

### Fahrrad-Verleih mit Dashboard
| Seite | Inhalt |
|-------|--------|
| **Dashboard** | Kennzahlen *Verfügbare Fahrräder*, *Ausgeliehene Fahrräder*, *Gesamt Fahrräder*, *Registrierte Personen*; Schnellzugriff; Liste der aktuell ausgeliehenen Fahrräder mit Rückgabe-Button und „überfällig“-Markierung |
| **Fahrrad-Verwaltung** | Fahrrad anlegen, bearbeiten, löschen; Status verfügbar/ausgeliehen |
| **Personen** | Personen verwalten; **Ausweis eingeben** (Art + Nummer) und/oder **hochladen** (PNG, JPEG, PDF, max. 5 MB) |
| **Reservierungen** | Alle Ausleihen, Filter nach Status; Ausgeben, Zurücknehmen, Stornieren |
| **Neue Ausleihe** | Person, Beginn, Dauer wählen → nur freie Fahrräder mit Preis; Gesamtpreis live |

Benutzerfreundlichkeit: Personen ohne Ausweis sind in der Ausleihe gesperrt (mit Hinweis), Dauer-Schnellwahl
(1 h … 2 Tage), Beginn „jetzt“ vorbelegt, Bestätigungsmeldungen und verständliche Fehlertexte, mobil bedienbar.

### Flexible Preise, Dauer und Beginn
- **Preise:** pro Fahrrad ein *Stundenpreis* und ein optionaler *Tagespreis* (24 h), jederzeit änderbar.
  Bestehende Ausleihen behalten ihren Preis.
- **Dauer:** 1 bis 168 Stunden, in Stundenschritten.
- **Beginn:** volle Stunde, heute oder später. Beginn in der laufenden Stunde ⇒ sofort *ausgeliehen*,
  sonst *reserviert*. Zeitliche Überschneidungen desselben Fahrrads sind ausgeschlossen.

**Preisformel:** `Tage = Stunden ÷ 24`, `Rest = Stunden mod 24`

```
Preis = Tage × Tagespreis + min(Rest × Stundenpreis, Tagespreis)     (ohne Tagespreis: Stunden × Stundenpreis)
```

Beispiel mit 8 €/h und 35 €/Tag: 3 h = 24 €, 5 h = 35 € (gedeckelt), 24 h = 35 €, 26 h = 51 €.

## Ablauf einer Ausleihe

```
reserviert ──Ausgeben (ab Beginn)──▶ ausgeliehen ──Zurücknehmen──▶ zurückgegeben
    └──Stornieren──▶ storniert
```

## API (Auszug)

| Methode | Pfad | Zweck |
|---------|------|-------|
| GET | `/api/stats` | Dashboard-Kennzahlen |
| GET/POST | `/api/bikes` | Fahrrad auflisten / anlegen |
| PUT/DELETE | `/api/bikes/{id}` | Fahrrad ändern / löschen |
| GET | `/api/bikes/available?start=<RFC3339>&hours=<n>` | Freie Fahrrad mit Preis im Zeitraum |
| GET/POST | `/api/persons` | Personen auflisten / anlegen |
| PUT/DELETE | `/api/persons/{id}` | Person ändern / löschen |
| GET/POST/DELETE | `/api/persons/{id}/document` | Ausweis ansehen / hochladen (`multipart`, Feld `file`) / entfernen |
| GET/POST | `/api/rentals` | Ausleihen auflisten / anlegen (`bikeId`, `personId`, `start`, `hours`) |
| POST | `/api/rentals/{id}/pickup` · `/return` · `/cancel` | Status ändern |

Fehler kommen als `{"error": "…"}` mit 400 (ungültig), 404 (nicht gefunden) oder 409 (Konflikt).
Geldbeträge sind ganze Cent.

## Projektstruktur

```
main.go                  Einstieg, Embed des Frontends
internal/domain/         Fahrrad, Person, Rental, Preisformel, Verfügbarkeit
internal/store/          Store-Interface, In-Memory-Store (thread-safe), Demodaten
internal/handler/        HTTP-Handler
web/                     index.html, app.js, style.css
SPEC.md                  Anforderungen R1–R18 (vor dem Code geschrieben)
CLAUDE.md                Projektregeln für Claude Code
```

## Annahmen und Grenzen (Prototyp)

- Keine Anmeldung, keine Persistenz, keine Zahlungsabwicklung.
- Der Ausweis wird nur gespeichert, nicht auf Echtheit geprüft; hochgeladene Dateien liegen im Arbeitsspeicher.
  Echte Ausweisdaten gehören nicht in einen Prototyp – für Tests Fantasiedaten verwenden.
- Der Preis wird bei der Anlage aus der geplanten Dauer berechnet; verspätete Rückgabe wird markiert, aber nicht nachberechnet.
- Der Beginn wird auf volle Stunden (UTC-basiert) abgeschnitten.
