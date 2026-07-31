package handlers

import (
	"collections-go-htmx-2/repo"
	"collections-go-htmx-2/templates"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func (app *App) GetCollectionsHandler(w http.ResponseWriter, r *http.Request) {
	collections, err := app.Queries.GetCollectionsByUserId(r.Context(), app.sessionUserID(r))
	if err != nil {
		slog.Error("get collections", "error", err)
		collections = []repo.Collection{}
	}
	render(w, r, templates.CollectionsPage(collections))
}

func (app *App) GetAddFormHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, templates.AddCollectionForm())
}

func (app *App) GetCancelFormHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, templates.EmptyAddForm())
}

func (app *App) PostCollectionHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := repo.CreateCollectionParams{
		UserID: app.sessionUserID(r),
		Title:  r.FormValue("title"),
	}
	collection, err := app.Queries.CreateCollection(r.Context(), params)
	if err != nil {
		slog.Error("create collection", "error", err)
		render(w, r, templates.CollectionCreateError("Cannot create collection, please try again"))
		return
	}
	render(w, r, templates.CollectionCreated(collection))
}

func (app *App) DeleteCollectionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "collection_id"), 10, 64)
	if err != nil {
		render(w, r, templates.CollectionDeleteError("Invalid collection ID"))
		return
	}
	err = app.Queries.DeleteCollection(r.Context(), repo.DeleteCollectionParams{
		ID:     id,
		UserID: app.sessionUserID(r),
	})
	if err != nil {
		slog.Error("delete collection", "error", err)
		render(w, r, templates.CollectionDeleteError("Internal error. Try again"))
		return
	}
	render(w, r, templates.CollectionDeleted(int(id)))
}
