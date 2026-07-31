package main

import (
	"collections-go-htmx-2/handlers"
	"collections-go-htmx-2/repo"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env:", err)
	}

	connString := os.Getenv("CONN_STRING")
	if connString == "" {
		log.Fatal("CONN_STRING env var must be set")
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("goose set dialect:", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatal("goose up:", err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Store = postgresstore.New(db)

	app := &handlers.App{
		Queries:        repo.New(db),
		SessionManager: sessionManager,
	}

	r := chi.NewRouter()
	r.Use(sessionManager.LoadAndSave)

	// Public routes
	r.Get("/register", app.GetRegisterHandler)
	r.Post("/register", app.PostRegisterHandler)
	r.Get("/login", app.GetLoginHandler)
	r.Post("/login", app.PostLoginHandler)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(app.RequireAuth)

		r.Get("/profile", app.GetProfileHandler)
		r.Post("/logout", app.PostLogoutHandler)

		r.Get("/collections", app.GetCollectionsHandler)
		r.Get("/collections/add-form", app.GetAddFormHandler)
		r.Get("/collections/cancel", app.GetCancelFormHandler)
		r.Post("/collections", app.PostCollectionHandler)
		r.Delete("/collections/{collection_id}", app.DeleteCollectionHandler)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
