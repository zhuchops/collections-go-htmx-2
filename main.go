package main

import (
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load environment")
	}

	connString := os.Getenv("CONN_STRING")
	if connString == "" {
		log.Fatal("Cannot read CONN_STRING env. MUST BE SET")
	}
	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatal("Failed to connect to database. Error: ", err.Error())
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Store = postgresstore.New(db)

	queries := repo.New(db)
	app := App{
		Queries: queries,
		SessionManager: sessionManager,
	}

	r := chi.NewRouter()

	r.Use(sessionManager.LoadAndSave)

	r.Get("/register", app.GetRegisterHandler)
	r.Get("/login", app.GetLoginHandler)
	r.Get("/collections/add-form", app.GetAddFormHandler)

	r.Post("/login", app.PostLoginHandler)
	r.Post("/register", app.PostRegisterHandler)

	r.Group(func(r chi.Router) {
		r.Use(app.RequireAuth)

		r.Get("/profile", app.GetProfileHandler)
		r.Get("/collections", app.GetCollectionsHandler)
		
		r.Post("/logout", app.PostLogoutHandler)
		r.Post("/collections/new", app.PostCollectionHandler)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
