# SPEC – Ski-Verleih (Spec-Driven Development)

Diese Spec wurde **vor** dem Code geschrieben. Jede Anforderung hat eine ID (R1 …),
die Tests und der Code verweisen darauf.

## Aufgabe

| Teil | Inhalt |
|------|--------|
| a) | Ski-Verleih wie in der Vorlesung, aber mit **Dashboard**; möglichst benutzerfreundlich; Ausweis **hochladbar** oder **eingebbar** (Nr. + Art) |
| b) | **Flexibilisierung** der Preise sowie der Verleihdauer und des Verleihbeginns (**stundenweise**) |

## Begriffe

- **Ski**: Verleihgegenstand (Marke, Modell, Länge, Kategorie, Preise).
- **Person**: Kunde mit Kontaktdaten und Ausweis.
- **Ausleihe** (Reservierung): Ski + Person + Beginn + Dauer in Stunden.
- Status einer Ausleihe: `reserviert` → `ausgeliehen` → `zurückgegeben`; `reserviert` → `storniert`.

## Anforderungen

### Dashboard (a)
- **R1** Das Dashboard zeigt vier Kennzahlen: *Verfügbare Ski*, *Ausgeliehene Ski*, *Gesamt Ski*, *Registrierte Personen*.
- **R2** Verfügbar = Gesamt − Ausgeliehen. Ausgeliehen = Ski mit einer Ausleihe im Status `ausgeliehen`.
- **R3** Schnellzugriff: „Ski verwalten“, „Personen verwalten“, „Neue Ausleihe“.
- **R4** Das Dashboard listet die aktuell ausgeliehenen Ski mit Rückgabezeit; überfällige sind markiert und direkt zurückgebbar.

### Verwaltung (a)
- **R5** Ski anlegen, ändern, löschen (Löschen nur ohne offene Ausleihe).
- **R6** Personen anlegen, ändern, löschen (Löschen nur ohne offene Ausleihe).
- **R7** Ausweis einer Person: Art (Personalausweis, Reisepass, Führerschein, Aufenthaltstitel) + Nummer **eingeben** und/oder Bilddatei/PDF **hochladen** (max. 5 MB; PNG, JPEG, PDF; Typ wird am Inhalt geprüft, nicht am Dateinamen).
- **R8** Eine Ausleihe ist nur möglich, wenn die Person einen Ausweis hat (Art + Nummer **oder** Upload).
- **R9** Ausleihen ansehen, filtern nach Status, Ski ausgeben, zurücknehmen, Reservierung stornieren.

### Flexibilität (b)
- **R10** *Preise*: Jeder Ski hat einen eigenen **Stundenpreis** und einen optionalen **Tagespreis** (Preis für 24 h). Beide sind jederzeit änderbar; bestehende Ausleihen behalten ihren Preis.
- **R11** *Preisformel*: `Tage = Stunden / 24` (ganzzahlig), `Rest = Stunden mod 24`.
  `Preis = Tage × Tagespreis + min(Rest × Stundenpreis, Tagespreis)`.
  Ohne Tagespreis: `Stunden × Stundenpreis`.
- **R12** *Dauer*: 1 bis 168 Stunden (7 Tage), in Stundenschritten.
- **R13** *Beginn*: volle Stunde, heute oder in der Zukunft. Beginn in der laufenden Stunde ⇒ sofort `ausgeliehen`; später ⇒ `reserviert`.
- **R14** *Verfügbarkeit*: Ein Ski darf sich zeitlich nicht überschneiden (`[Beginn, Ende)`, direkt aufeinanderfolgend ist erlaubt). Ein überfälliger, nicht zurückgegebener Ski bleibt blockiert.
- **R15** Beim Anlegen einer Ausleihe sieht man nur Ski, die im gewählten Zeitraum frei sind, jeweils mit Preis; Gesamtpreis wird live angezeigt.
- **R16** Ausgeben einer Reservierung ist erst ab Beginn möglich.

### Technik
- **R17** Go-Backend, In-Memory-Store (thread-safe), HTML/JS-Frontend, keine externen Abhängigkeiten.
- **R18** Beim Start werden Demodaten geladen (20 Ski, 10 Personen, 3 ausgeliehen); abschaltbar mit `-seed=false`.

## Abnahmekriterien (Auszug)

| ID | Prüfung |
|----|---------|
| R2 | Demodaten: Dashboard zeigt 17 / 3 / 20 / 10 |
| R11 | 1 h → Stundenpreis; 24 h → Tagespreis; 25 h → Tag + Stunde; 10 h mit teurem Stundenpreis → gedeckelt auf Tagespreis |
| R14 | Zweite Ausleihe desselben Ski im überlappenden Zeitraum → HTTP 409; unmittelbar danach → 201 |
| R14 | 20 parallele Anfragen auf denselben Ski → genau eine gelingt |
| R7/R8 | Person ohne Ausweis → Ausleihe HTTP 400; Upload einer `.exe` mit Dateiendung `.png` → HTTP 400 |
