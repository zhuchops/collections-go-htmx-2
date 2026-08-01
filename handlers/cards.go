package handlers

import (
	"collections-go-htmx-2/repo"
	"collections-go-htmx-2/templates"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func (app *App) GetCardsHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.SessionManager.GetInt64(r.Context(), "user_id")
	collectionIdStr := chi.URLParam(r, "collection_id")
	collectionId, err := strconv.ParseInt(collectionIdStr, 10, 64)
	if err != nil {
		render(w, r, templates.ErrorPage("Cannot parse collection id. Try another one", nil))
		return
	}
	getCollectionParams := repo.GetCollectionParams{ID: collectionId, UserID: userId}
	collection, err := app.Queries.GetCollection(r.Context(), getCollectionParams)
	if err != nil {
		render(w, r, templates.ErrorPage("No collection with such id found", nil))
		return
	}
	getCardsParams := repo.GetCardsParams{UserID: userId, CollectionID: collectionId}
	cards, err := app.Queries.GetCards(r.Context(), getCardsParams)
	if err != nil {
		cards = []repo.Card{}
	}
	flashes := app.PopFlashes(r)
	render(w, r, templates.CardsPage(collection, cards, flashes))
}

func (app *App) GetCardAddFormHandler(w http.ResponseWriter, r *http.Request) {
	collectionId, err := strconv.ParseInt(chi.URLParam(r, "collection_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	render(w, r, templates.AddCardForm(int(collectionId)))
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
	descriptionNullable := sql.NullString{String: description, Valid: description != ""}
	params := repo.CreateCardParams{Title: title, Description: descriptionNullable, UserID: userId, CollectionID: collectionId}
	card, err := app.Queries.CreateCard(r.Context(), params)
	if err != nil {
		slog.Error("create card", "error", err)
		renderWithToast(w, r, templates.AddCardForm(int(collectionId)), ToastError, "Cannot create card. Internal error")
		return
	}
	render(w, r, templates.CardCreated(card))
}

func (app *App) DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.SessionManager.GetInt64(r.Context(), "user_id")
	cardId, err := strconv.ParseInt(chi.URLParam(r, "card_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	params := repo.DeleteCardParams{UserID: userId, ID: cardId}
	err = app.Queries.DeleteCard(r.Context(), params)
	if err != nil {
		slog.Error("delete card", "error", err)
		renderToast(w, r, ToastError, "Cannot delete card. Internal error")
		return
	}
	render(w, r, templates.CardDeleted(int(cardId)))
}

func (app *App) GetCardHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionUserID(r)
	r.ParseForm()
	cardId, err := strconv.ParseInt(chi.URLParam(r, "card_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	params := repo.GetCardParams{ID: cardId, UserID: userId}
	card, err := app.Queries.GetCard(r.Context(), params)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	flashes := app.PopFlashes(r)
	render(w, r, templates.CardPage(card, flashes))
}

func (app *App) GetCardUpdateFormHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionUserID(r)
	cardId, err := strconv.ParseInt(chi.URLParam(r, "card_id"), 10, 64)
	if err != nil {
		renderWithToast(w, r, templates.EmptyUpdateCardForm(), ToastError, "Cannot parse card id from URL")
		return
	}
	params := repo.GetCardParams{UserID: userId, ID: cardId}
	card, err := app.Queries.GetCard(r.Context(), params)
	if err != nil {
		renderWithToast(w, r, templates.EmptyUpdateCardForm(), ToastError, "Cannot get card information")
		return
	}
	render(w, r, templates.UpdateCardForm(card))
}

func (app *App) PutCardHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")
	description := r.FormValue("description")
	descriptionNullable := sql.NullString{String: description, Valid: description != ""}
	userId := app.sessionUserID(r)
	cardId, err := strconv.ParseInt(chi.URLParam(r, "card_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	collectionId, err := strconv.ParseInt(chi.URLParam(r, "collection_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	submitted := repo.Card{ID: cardId, UserID: userId, CollectionID: collectionId, Title: title, Description: descriptionNullable}
	params := repo.UpdateCardParams{ID: cardId, UserID: userId, Title: title, Description: descriptionNullable}
	err = app.Queries.UpdateCard(r.Context(), params)
	if err != nil {
		slog.Error("update card", "error", err.Error())
		renderWithToast(w, r, templates.UpdateCardForm(submitted), ToastError, "Card failed to update")
		return
	}
	render(w, r, templates.CardInfoUpdated(submitted))
}

func (app *App) GetAddCardFormCancelHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, templates.EmptyAddCardForm())
}

func (app *App) GetCardUpdateFormCancelHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, templates.EmptyUpdateCardForm())
}