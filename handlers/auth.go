package handlers

import (
	"collections-go-htmx-2/repo"
	"collections-go-htmx-2/templates"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func (app *App) GetRegisterHandler(w http.ResponseWriter, r *http.Request) {
	flashes := app.PopFlashes(r)
	render(w, r, templates.RegisterPage(flashes))
}

func (app *App) PostRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordAgain := r.FormValue("password_again")
	slog.Info("registration attempt", "username", username, "email", email)

	if password != passwordAgain {
		renderToast(w, r, ToastError, "Passwords do not match")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		slog.Error("bcrypt", "error", err)
		renderToast(w, r, ToastError, "Internal error, please try again")
		return
	}

	_, err = app.Queries.CreateUser(r.Context(), repo.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			renderToast(w, r, ToastError, "Email or username already taken")
			return
		}
		slog.Error("create user", "error", err)
		renderToast(w, r, ToastError, "Internal error, please try again")
		return
	}

	renderToast(w, r, ToastSuccess, "Registered! You can now log in.")
}

func (app *App) GetLoginHandler(w http.ResponseWriter, r *http.Request) {
	flashes := app.PopFlashes(r)
	render(w, r, templates.LoginPage(flashes))
}

func (app *App) PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	slog.Info("login attempt", "email", email)

	user, err := app.Queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			renderToast(w, r, ToastError, "Invalid email or password")
			return
		}
		slog.Error("get user by email", "error", err)
		renderToast(w, r, ToastError, "Internal error, please try again")
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		renderToast(w, r, ToastError, "Invalid email or password")
		return
	}

	app.SessionManager.Put(r.Context(), "user_id", user.ID)
	app.PutFlash(r, ToastSuccess, "Welcome back!")
	w.Header().Set("HX-Redirect", "/profile")
	w.WriteHeader(http.StatusOK)
}

func (app *App) PostLogoutHandler(w http.ResponseWriter, r *http.Request) {
	app.SessionManager.Destroy(r.Context())
	app.PutFlash(r, ToastInfo, "Logged out successfully")
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}

func (app *App) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := app.Queries.GetUserById(r.Context(), app.sessionUserID(r))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	flashes := app.PopFlashes(r)
	render(w, r, templates.ProfilePage(user.Email, user.Username, flashes))
}
