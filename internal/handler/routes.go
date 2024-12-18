package handler

import (
	"database/sql"
	"net/http"

	"github.com/DhruvikDonga/simplysocket"
	"github.com/DhruvikDonga/wordsbattle/internal/modules/game"
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
	//messagehandler := cowgameclient.ClientCustomMessage{}
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

	roomdata := &game.RoomData{RandomRooms: []string{}}
	ms := simplysocket.NewMeshServer("cowgame", &simplysocket.MeshServerConfig{DirectBroadCast: false}, roomdata)
	// initialize websocket link cowgame connection clash of words
	r.HandleFunc("/wsmesh", func(w http.ResponseWriter, r *http.Request) {
		simplysocket.ServeWs(ms, w, r)
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
