package handlers

import (
	"collections-go-htmx-2/templates"
	"net/http"

	"github.com/a-h/templ"
)

type ToastType = templates.ToastType

const (
	ToastSuccess ToastType = templates.ToastSuccess
	ToastError   ToastType = templates.ToastError
	ToastInfo    ToastType = templates.ToastInfo
	ToastWarning ToastType = templates.ToastWarning
)

type Toast = templates.Toast

// PutFlash stores a flash toast message in the session for the next page load.
func (app *App) PutFlash(r *http.Request, t ToastType, msg string) {
	app.SessionManager.Put(r.Context(), "flash_type", string(t))
	app.SessionManager.Put(r.Context(), "flash_msg", msg)
}

// PopFlashes pops any flash toast message stored in the session.
func (app *App) PopFlashes(r *http.Request) []Toast {
	msg := app.SessionManager.PopString(r.Context(), "flash_msg")
	if msg == "" {
		return nil
	}
	tStr := app.SessionManager.PopString(r.Context(), "flash_type")
	if tStr == "" {
		tStr = string(ToastInfo)
	}
	return []Toast{
		{
			Type:    ToastType(tStr),
			Message: msg,
		},
	}
}

// renderWithToast renders a component along with an out-of-band Toast message.
func renderWithToast(w http.ResponseWriter, r *http.Request, comp templ.Component, toastType ToastType, msg string) {
	render(w, r, comp)
	if msg != "" {
		render(w, r, templates.ToastOOB(Toast{Type: toastType, Message: msg}))
	}
}

// renderToast renders only an out-of-band Toast message.
func renderToast(w http.ResponseWriter, r *http.Request, toastType ToastType, msg string) {
	render(w, r, templates.ToastOOB(Toast{Type: toastType, Message: msg}))
}

func renderToasts(w http.ResponseWriter, r *http.Request, toasts []Toast) {
	render(w, r, templates.ToastContainer(toasts))
}