// Package uimsg carries user-visible backend messages to the UI.
//
// A message travels as a stable Code (a key of
// cmd/launcher/web/js/i18n/translations.js, by convention "msg_*") with
// Params filling the key's {placeholders}. The Polish Text goes along as the
// log text and as the fallback the UI shows for codes it does not know.
// cmd/launcher/backend_text_test.go checks that every message the backend
// sends to the UI has a code and that every code exists in translations.js.
package uimsg

import "errors"

// Msg is one user-visible message.
type Msg struct {
	Code   string
	Params map[string]string
	Text   string
}

// New builds a Msg. kv lists placeholder name/value pairs:
// New("msg_inst_copied", "Skopiowano do "+dir, "path", dir).
func New(code, text string, kv ...string) Msg {
	m := Msg{Code: code, Text: text}
	if len(kv) > 1 {
		m.Params = make(map[string]string, len(kv)/2)
		for i := 0; i+1 < len(kv); i += 2 {
			m.Params[kv[i]] = kv[i+1]
		}
	}
	return m
}

// Error is an error that carries a Msg.
type Error struct{ Msg Msg }

func (e *Error) Error() string { return e.Msg.Text }

// Errorf returns an error carrying New(code, text, kv...).
func Errorf(code, text string, kv ...string) error {
	return &Error{Msg: New(code, text, kv...)}
}

// From returns the Msg carried by err, or a Msg without a code holding
// err.Error() (technical errors are shown as they are).
func From(err error) Msg {
	if err == nil {
		return Msg{}
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Msg
	}
	return Msg{Text: err.Error()}
}
