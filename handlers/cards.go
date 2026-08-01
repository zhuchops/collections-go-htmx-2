package handlers

import (
	"collections-go-htmx-2/repo"
	"collections-go-htmx-2/templates"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func (app *App) GetCardsHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.SessionManager.GetInt64(r.Context(), "user_id")
	collectionIdStr := chi.URLParam(r, "collection_id")
	collectionId, err := strconv.ParseInt(collectionIdStr, 10, 64)
	if err != nil {
		render(w, r, templates.ErrorPage("Cannot parse collection id. Try another one"))
		return
	}
	getCollectionParams := repo.GetCollectionParams{ID: collectionId, UserID: userId}
	collection, err := app.Queries.GetCollection(r.Context(), getCollectionParams)
	if err != nil {
		render(w, r, templates.ErrorPage("No collection with such id found"))
		return
	}
	getCardsParams := repo.GetCardsParams{UserID: userId, CollectionID: collectionId}
	cards, err := app.Queries.GetCards(r.Context(), getCardsParams)
	if err != nil {
		cards = []repo.Card{}
	}
	render(w, r, templates.CardsPage(collection, cards))
}

func (app *App) GetCardAddFormHandler(w http.ResponseWriter, r *http.Request) {
	collectionId, err := strconv.ParseInt(chi.URLParam(r, "collection_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	render(w, r, templates.AddCardForm(int(collectionId), ""))
}

func (app *App) PostCardHandler(w http.ResponseWriter, r *http.Request) {
	collectionId, err := strconv.ParseInt(chi.URLParam(r, "collection_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	userId := app.SessionManager.GetInt64(r.Context(), "user_id")
	r.ParseForm()
	title := r.FormValue("title")		
	description := r.FormValue("description")
	descriptionNullable := sql.NullString{String: description, Valid: description != ""};
	params := repo.CreateCardParams{Title: title, Description: descriptionNullable, UserID: userId, CollectionID: collectionId}
	card, err := app.Queries.CreateCard(r.Context(), params)
	if err != nil {
		render(w, r, templates.CardCreateError("Cannot create card. Interanl error"))
		return
	}
	render(w, r, templates.CardCreated(card))
}

func (app *App) DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	// collectionId, err := strconv.ParseInt(chi.URLParam(r, "collection_id"), 10, 64)
	// if err != nil {
	// 	http.Error(w, "bad request", http.StatusBadRequest)
	// 	return
	// }
	userId := app.SessionManager.GetInt64(r.Context(), "user_id")
	cardId, err := strconv.ParseInt(chi.URLParam(r, "card_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}	
	params := repo.DeleteCardParams{UserID: userId, ID: cardId}
	err = app.Queries.DeleteCard(r.Context(), params)
	if err != nil {
		render(w, r, templates.CardDeleteError("Cannot delete card. Internal error"))
		return
	}
	render(w, r, templates.CardDeleted(int(cardId)))
}