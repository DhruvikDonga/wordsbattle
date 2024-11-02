package main

import (
	"bufio"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DhruvikDonga/wordsbattle/internal/handler"
	"github.com/DhruvikDonga/wordsbattle/internal/modules/cowgameclient"
	"github.com/DhruvikDonga/wordsbattle/internal/modules/game"
	"github.com/DhruvikDonga/wordsbattle/pkg/db"
	"github.com/DhruvikDonga/wordsbattle/util"
)

type App struct {
	db     *sql.DB
	config util.Config
	ctx    context.Context
}

func main() {
	config, err := util.LoadConfig("../../")
	if err != nil {
		log.Fatalf("cannot load config %v", err)
	}
	dbSource := config.DBSource
	database, err := db.Initialize(dbSource)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	defer database.Conn.Close()

	// Run db migration
	db.RunDBMigration(config.MigrationURL, config.DBSource)

	//creating word dicitionary
	log.Println("creating a dictionary")
	readFile, err := os.Open("data.txt")
	if err != nil {
		log.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		cowgameclient.Worddictionary[fileScanner.Text()] = true
		game.Worddictionary[fileScanner.Text()] = true
	}
	readFile.Close()

	newapp := handler.NewApp(database.Conn, config)
	//HTTP server
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: handler.RouteService(newapp),
	}

	// Server run context
	log.Println("Server started on :- 0.0.0.0:8080")
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		log.Println("shutting down")
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err = server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
