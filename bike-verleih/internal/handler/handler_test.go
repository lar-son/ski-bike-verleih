package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bikeverleih/internal/store"
)

var now = time.Date(2026, 1, 10, 10, 30, 0, 0, time.UTC)

func newServer(t *testing.T) http.Handler {
	t.Helper()
	st := store.NewMemoryStore()
	if err := st.Seed(now); err != nil {
		t.Fatal(err)
	}
	h := New(st)
	h.Now = func() time.Time { return now }
	return h.Routes(http.NotFoundHandler())
}

func do(t *testing.T, srv http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rd)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestStatsEndpoint(t *testing.T) { // SPEC R1, R2
	rec := do(t, newServer(t), "GET", "/api/stats", nil)
	var st store.Stats
	if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || st.Verfuegbar != 17 || st.Ausgeliehen != 3 || st.Gesamt != 20 || st.Personen != 10 {
		t.Errorf("%d %+v", rec.Code, st)
	}
}

func TestRentalFlow(t *testing.T) { // SPEC R13, R14, R9
	srv := newServer(t)
	start := now.Truncate(time.Hour).Add(5 * time.Hour).Format(time.RFC3339)
	body := map[string]any{"bikeId": 2, "personId": 6, "start": start, "hours": 2}

	rec := do(t, srv, "POST", "/api/rentals", body)
	if rec.Code != 201 {
		t.Fatalf("anlegen: %d %s", rec.Code, rec.Body)
	}
	var created rentalView
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.Status != "reserviert" || created.PriceCent != 1000 {
		t.Errorf("created = %+v", created)
	}

	if rec := do(t, srv, "POST", "/api/rentals", body); rec.Code != 409 { // Überschneidung
		t.Errorf("Überschneidung: %d", rec.Code)
	}
	body["start"] = now.Truncate(time.Hour).Add(7 * time.Hour).Format(time.RFC3339) // direkt anschließend
	if rec := do(t, srv, "POST", "/api/rentals", body); rec.Code != 201 {
		t.Errorf("anschließend: %d %s", rec.Code, rec.Body)
	}

	if rec := do(t, srv, "POST", fmt.Sprintf("/api/rentals/%d/pickup", created.ID), nil); rec.Code != 409 {
		t.Errorf("Pickup vor Beginn: %d", rec.Code)
	}
	if rec := do(t, srv, "POST", fmt.Sprintf("/api/rentals/%d/cancel", created.ID), nil); rec.Code != 200 {
		t.Errorf("Cancel: %d", rec.Code)
	}
}

func TestRentalWithoutID(t *testing.T) { // SPEC R8: Person 9 hat keinen Ausweis
	body := map[string]any{"bikeId": 2, "personId": 9, "start": now.Format(time.RFC3339), "hours": 1}
	rec := do(t, newServer(t), "POST", "/api/rentals", body)
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "Ausweis") {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestAvailableBikes(t *testing.T) { // SPEC R15
	srv := newServer(t)
	start := now.Truncate(time.Hour).Format(time.RFC3339)
	rec := do(t, srv, "GET", "/api/bikes/available?start="+start+"&hours=2", nil)
	var offers []bikeOffer
	_ = json.Unmarshal(rec.Body.Bytes(), &offers)
	if rec.Code != 200 || len(offers) != 17 {
		t.Errorf("%d, %d Angebote", rec.Code, len(offers))
	}
	if rec := do(t, srv, "GET", "/api/bikes/available?start="+start+"&hours=0", nil); rec.Code != 400 {
		t.Errorf("hours=0: %d", rec.Code)
	}
}

func upload(t *testing.T, srv http.Handler, personID int, name string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", name)
	_, _ = fw.Write(data)
	_ = mw.Close()
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/persons/%d/document", personID), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestDocumentUpload(t *testing.T) { // SPEC R7, R8
	srv := newServer(t)
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	if rec := upload(t, srv, 9, "ausweis.png", png); rec.Code != 200 {
		t.Fatalf("PNG: %d %s", rec.Code, rec.Body)
	}
	rec := do(t, srv, "GET", "/api/persons/9/document", nil)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" {
		t.Errorf("Download: %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	// Person 9 hat nun einen Ausweis (Upload) ⇒ Ausleihe möglich
	body := map[string]any{"bikeId": 2, "personId": 9, "start": now.Format(time.RFC3339), "hours": 1}
	if rec := do(t, srv, "POST", "/api/rentals", body); rec.Code != 201 {
		t.Errorf("Ausleihe nach Upload: %d %s", rec.Code, rec.Body)
	}
	// ausführbare Datei mit .png-Endung wird am Inhalt erkannt und abgelehnt
	if rec := upload(t, srv, 8, "virus.png", []byte("MZ\x90\x00\x03\x00\x00\x00")); rec.Code != 400 {
		t.Errorf("Fake-PNG: %d", rec.Code)
	}
	if rec := upload(t, srv, 8, "gross.pdf", append([]byte("%PDF-1.4\n"), make([]byte, 6<<20)...)); rec.Code != 400 {
		t.Errorf("zu groß: %d", rec.Code)
	}
}

func TestBikeCRUDAndPriceChange(t *testing.T) { // SPEC R5, R10
	srv := newServer(t)
	rec := do(t, srv, "POST", "/api/bikes", map[string]any{"marke": "Test", "modell": "T1", "rahmenCm": 54, "preisStundeCent": 500, "preisTagCent": 2000})
	if rec.Code != 201 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if rec := do(t, srv, "POST", "/api/bikes", map[string]any{"marke": "", "modell": "T1", "rahmenCm": 54, "preisStundeCent": 500}); rec.Code != 400 {
		t.Errorf("ungültig: %d", rec.Code)
	}
	if rec := do(t, srv, "PUT", "/api/bikes/21", map[string]any{"marke": "Test", "modell": "T1", "rahmenCm": 54, "preisStundeCent": 600, "preisTagCent": 2400}); rec.Code != 200 {
		t.Errorf("Preis ändern: %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, srv, "DELETE", "/api/bikes/1", nil); rec.Code != 409 { // Bike 1 ist ausgeliehen
		t.Errorf("Löschen ausgeliehen: %d", rec.Code)
	}
	if rec := do(t, srv, "DELETE", "/api/bikes/21", nil); rec.Code != 204 {
		t.Errorf("Löschen: %d", rec.Code)
	}
}
