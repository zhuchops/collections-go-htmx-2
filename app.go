package main

import (
	"collections-go-htmx-2/repo"
	"collections-go-htmx-2/templates"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Queries *repo.Queries
	SessionManager *scs.SessionManager
}

func (app *App) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := app.SessionManager.GetInt64(r.Context(), "user_id")
		if userId == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *App) GetRegisterHandler(w http.ResponseWriter, r *http.Request) {
	templates.RegisterPage().Render(r.Context(), w)
}

func (app *App) PostRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordAgain := r.FormValue("password_again")
	slog.Info("Registration attempt", "username", username, "email", email)

	if password != passwordAgain {
		templates.ErrorPartial("Registration failed. Passwords do not match").Render(r.Context(), w)
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		templates.ErrorPartial("Registration failed due to internal error. Please try again").Render(r.Context(), w)
		slog.Error("Cannot generate password")
		return
	}
	params := repo.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: string(passwordHash),
	}
	_, err = app.Queries.CreateUser(r.Context(), params)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			if pgxErr.Code == "23505" {
				templates.ErrorPartial("Registration failed. User with such email or username already exists").Render(r.Context(), w)
				return
			}
		}
		templates.ErrorPartial("Registration failed due to internal error. Please try again").Render(r.Context(), w)
		slog.Error("Datbase error", "error", err.Error())
		return
	}
	templates.SuccessPartial("Registration succeed. You can now login").Render(r.Context(), w)
}

func (app *App) GetLoginHandler(w http.ResponseWriter, r *http.Request) {
	templates.LoginPage().Render(r.Context(), w)
}

func (app *App) PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	slog.Info("Login attempt", "email", email)

	user, err := app.Queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			templates.ErrorPartial("Login failed. Bad credentials").Render(r.Context(), w)
			return
		}
		templates.ErrorPartial("Login failed due to internal error. Please try again").Render(r.Context(), w)
		slog.Error("Database error", "error", err.Error())
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		templates.ErrorPartial("Login failed. Bad credentials").Render(r.Context(), w)
		return
	}
	app.SessionManager.Put(r.Context(), "user_id", user.ID)
	w.Header().Set("HX-Redirect", "/profile")
	w.WriteHeader(http.StatusOK)
}

func (app *App) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.SessionManager.GetInt64(r.Context(), "user_id")
	user, err := app.Queries.GetUserById(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	templates.ProfilePage(user.Email, user.Username).Render(r.Context(), w)
}

func (app *App) PostLogoutHandler(w http.ResponseWriter, r *http.Request) {
	app.SessionManager.Destroy(r.Context())
	w.Header().Set("Hx-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}

func (app *App) GetCollectionsHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.SessionManager.GetInt64(r.Context(), "user_id")
	nullUserId := sql.NullInt64{Int64: userId, Valid: userId != 0}
	collections, err := app.Queries.GetCollectionsByUserId(r.Context(), nullUserId)
	if err != nil {
		slog.Error("Database error", "error", err.Error())
	}
	templates.CollectionsPage(collections).Render(r.Context(), w)
}

func (app *App) GetAddFormHandler(w http.ResponseWriter, r *http.Request) {
	templates.AddCollectionForm().Render(r.Context(), w)
}

func (app *App) PostCollectionHandler(w http.ResponseWriter, r *http.Request) {
	userId := app.SessionManager.GetInt(r.Context(), "user_id")
	nullUserId := sql.NullInt64{Int64: int64(userId), Valid: userId != 0}
	r.ParseForm()
	title := r.FormValue("title")
	params := repo.CreateCollectionParams{UserID: nullUserId, Title: title}
	collection, _ := app.Queries.CreateCollection(r.Context(), params)
	templates.CollectionItem(collection).Render(r.Context(), w)
}