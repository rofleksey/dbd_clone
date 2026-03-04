package main

import (
	"log"
	"net/http"
	"os"

	"dbd-master/internal/auth"
	"dbd-master/internal/db"
	"dbd-master/internal/docker"
	"dbd-master/internal/lobby"
	"dbd-master/internal/router"
)

func main() {
	log.Println("Starting DBD Clone Master Server...")

	// Init auth
	auth.Init()

	// Init database
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Init managers
	lobbyMgr := lobby.NewManager()
	dockerMgr := docker.NewManager()

	// Create router
	r := router.New(lobbyMgr, dockerMgr)

	port := os.Getenv("MASTER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Master server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
