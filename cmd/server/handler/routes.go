package handler

import (
	"database/sql"
	"net/http"

	"github.com/DhruvikDonga/wordsbattle/pkg/db"
	"github.com/DhruvikDonga/wordsbattle/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

var dbInstance db.Database

type App struct {
	db     *sql.DB
	config util.Config
}

func NewApp(pgdb *sql.DB, conf util.Config) *App {
	return &App{
		db:     pgdb,
		config: conf,
	}
}

func RouteService(app *App) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// start websocket server
	wsServer := NewWebSocketServer()
	go wsServer.Run()

	// initialize websocket connection
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(wsServer, w, r)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("home"))
	})
	r.Get("/api/v1/home", app.HelloHandler)
	r.Post("/api/v1/register", app.RegisterHandler)
	r.Post("/api/v1/login", app.LoginHandler)

	r.Get("/api/v1/isauthorized", app.IsAuthorized(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("authorized"))
	}))
	return r
}
