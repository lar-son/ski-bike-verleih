package handler

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"slices"

	"skiverleih/internal/domain"
)

const maxUpload = 5 << 20 // 5 MB (SPEC R7)

// Erlaubt sind nur Typen, die am Dateiinhalt erkannt wurden – nicht am Namen.
var allowedDocTypes = []string{"image/png", "image/jpeg", "application/pdf"}

func (h *Handler) listPersons(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.Store.ListPersons())
}

func (h *Handler) createPerson(w http.ResponseWriter, r *http.Request) {
	var p domain.Person
	if err := decode(r, &p); err != nil {
		writeErr(w, err)
		return
	}
	if err := p.Validate(); err != nil {
		writeErr(w, err)
		return
	}
	created, err := h.Store.CreatePerson(p)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updatePerson(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var p domain.Person
	if err := decode(r, &p); err != nil {
		writeErr(w, err)
		return
	}
	p.ID = id
	if err := p.Validate(); err != nil {
		writeErr(w, err)
		return
	}
	updated, err := h.Store.UpdatePerson(p)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deletePerson(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := h.Store.DeletePerson(id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// uploadDocument: POST /api/persons/{id}/document, multipart-Feld "file" (SPEC R7).
func (h *Handler) uploadDocument(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload+1<<20)
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, domain.Invalid("Datei fehlt oder ist zu groß (max. 5 MB)"))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxUpload+1))
	if err != nil {
		writeErr(w, fmt.Errorf("upload lesen: %w", err))
		return
	}
	if len(data) > maxUpload {
		writeErr(w, domain.Invalid("Datei ist zu groß (max. 5 MB)"))
		return
	}
	contentType := http.DetectContentType(data)
	if !slices.Contains(allowedDocTypes, contentType) {
		writeErr(w, domain.Invalid("Nur PNG, JPEG oder PDF erlaubt"))
		return
	}
	doc := &domain.Document{FileName: filepath.Base(hdr.Filename), ContentType: contentType, Data: data}
	person, err := h.Store.SetDocument(id, doc)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, person)
}

func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	p, err := h.Store.GetPerson(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	if p.Dokument == nil {
		writeErr(w, domain.NotFound("Dokument von Person", id))
		return
	}
	w.Header().Set("Content-Type", p.Dokument.ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": p.Dokument.FileName}))
	if _, err := w.Write(p.Dokument.Data); err != nil {
		return
	}
}

func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	person, err := h.Store.SetDocument(id, nil)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, person)
}
