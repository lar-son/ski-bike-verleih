package domain

import (
	"errors"
	"fmt"
)

// Sentinel-Fehler; Handler mappen sie auf HTTP-Statuscodes.
var (
	ErrNotFound = errors.New("nicht gefunden")
	ErrInvalid  = errors.New("ungültige Eingabe")
	ErrConflict = errors.New("Konflikt")
)

// Error trägt eine nutzerfreundliche Meldung und bleibt per errors.Is
// mit dem Sentinel vergleichbar.
type Error struct {
	Kind error
	Msg  string
}

func (e *Error) Error() string { return e.Msg }
func (e *Error) Unwrap() error { return e.Kind }

func Invalid(format string, a ...any) error {
	return &Error{Kind: ErrInvalid, Msg: fmt.Sprintf(format, a...)}
}

func Conflict(format string, a ...any) error {
	return &Error{Kind: ErrConflict, Msg: fmt.Sprintf(format, a...)}
}

func NotFound(what string, id int) error {
	return &Error{Kind: ErrNotFound, Msg: fmt.Sprintf("%s %d nicht gefunden", what, id)}
}
