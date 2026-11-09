# Prototype Project – Ski-Verleih

Nur Prototypen. Priorität: schnell verständlich, leicht änderbar.
Vorgehen: Spec-Driven Development – erst `SPEC.md` anpassen, dann Code und Tests.

## Stack
- Backend: Go ('objektorientiert': Structs + Methoden, Interfaces wo sinnvoll)
- DB: In-Memory (kein Postgres/SQLite, kein Object-Relational Mapping (ORM) nötig)
- Frontend: reines HTML/JS (`web/`, wird per `go:embed` in die Binary eingebettet)
- Versionierung: lokales Git, bei Bedarf GitHub

## Commands
- Backend starten: `go run .` (Standard: http://localhost:8080, Port ändern mit `-addr :9000`)
- Ohne Demodaten starten: `go run . -seed=false`
- Tests: `go test ./...`
- Format: `gofmt -w .` bzw. `go fmt ./...`
- Lint (falls eingerichtet): `golangci-lint run`
- Frontend (HTML): wird vom Go-Server ausgeliefert, keine separate Build-Stufe

## Architektur (Go)
- `main.go` → Einstieg, HTTP-Server, Embed des Frontends
- `internal/domain/` → Kern-Structs + Methoden (Business-Logik: Ski, Person, Rental, Preisformel, Verfügbarkeit)
- `internal/store/` → In-Memory-Store (Maps + RWMutex), `Store`-Interface, Demodaten (`seed.go`)
- `internal/handler/` → HTTP-Handler (JSON-API unter `/api`), rufen Domain-Objekte auf
- `web/` → `index.html`, `app.js`, `style.css` (Dashboard, Ski, Personen, Reservierungen, Neue Ausleihe)
- Keine globalen Variablen außer dem Store-Singleton wenn nötig
- Fehler mit `fmt.Errorf("%w", err)` wrappen, nicht verschlucken
- Domain-Fehler: `domain.ErrNotFound` (404), `ErrInvalid` (400), `ErrConflict` (409)

## Konventionen
- Methoden an Structs hängen statt freier Funktionen, wo es Sinn ergibt
- Interfaces nur bei echtem Bedarf (z.B. Store-Interface für spätere Austauschbarkeit)
- Kein Dependency Injection Framework
- Keine neuen Dependencies ohne Nachfrage
- In-Memory-Store muss thread-safe sein
- Keine Secrets im Code (auch bei Prototypen nicht)
- Geldbeträge immer als ganze Cent (`int`), nie als Float
- Zeiten im API als RFC 3339; Beginn wird auf volle Stunden abgeschnitten
- UI-Texte auf Deutsch

## Frontend
- HTML-Variante: möglichst minimales JS, klare Trennung von Backend-API
- API-Calls immer gegen die Go-Endpoints, nie direkt gegen In-Memory-Strukturen
- Nutzereingaben beim Rendern immer escapen (`esc()`)

## Regeln
- Keine generierten Dateien von Hand editieren
- Commit-Messages: kurz und imperativ
- Bei Unklarheit fragen statt raten
- Prototyp-Fokus: lieber einfach und klar als elegant und komplex
