package handlers

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

// render writes a templ component to the response writer and logs any error.
func render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	if err := component.Render(r.Context(), w); err != nil {
		slog.Error("render error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
