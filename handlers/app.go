package handlers

import (
	"collections-go-htmx-2/repo"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type App struct {
	Queries        *repo.Queries
	SessionManager *scs.SessionManager
}

// sessionUserID is a convenience helper to read the authenticated user's ID
// from the session. Returns 0 if not authenticated.
func (app *App) sessionUserID(r *http.Request) int64 {
	return app.SessionManager.GetInt64(r.Context(), "user_id")
}
