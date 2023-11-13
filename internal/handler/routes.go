package handler

import (
	"database/sql"
	"net/http"

	"github.com/DhruvikDonga/wordsbattle/internal/modules/cowgameclient"
	"github.com/DhruvikDonga/wordsbattle/internal/modules/game"
	"github.com/DhruvikDonga/wordsbattle/pkg/db"
	"github.com/DhruvikDonga/wordsbattle/pkg/gogamelink"
	"github.com/DhruvikDonga/wordsbattle/pkg/gogamemesh"
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

	// start websocket server
	// wsServer := cowgameclient.NewLobbyServer()
	// go wsServer.Run()

	// link new game server
	roomgameendticker := cowgameclient.RoomGameBot{}
	lobbyserv := gogamelink.RunNewLobbyServer("cowgame", roomgameendticker)
	messagehandler := cowgameclient.ClientCustomMessage{}

	// // initialize websocket connection clash of words
	// r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	cowgameclient.ServeWs(wsServer, w, r)
	// })
	roomdata := &game.RoomData{
		IsRandomGame: false,
		PlayerLimit:  10,
	}
	ms := gogamemesh.NewMeshServer("cowgame", &gogamemesh.MeshServerConfig{DirectBroadCast: true}, roomdata)
	ms.RunMeshServer()
	// // initialize websocket link cowgame connection clash of words
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		gogamelink.ServeWs(lobbyserv, w, r, messagehandler)
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
