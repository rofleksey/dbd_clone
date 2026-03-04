package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"dbd-game/internal/game"
	"dbd-game/internal/ws"
)

type PlayerSetup struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func main() {
	gameID, _ := strconv.Atoi(os.Getenv("GAME_ID"))
	gamePort := os.Getenv("GAME_PORT")
	masterURL := os.Getenv("MASTER_URL")
	killerID, _ := strconv.Atoi(os.Getenv("KILLER_ID"))

	if gamePort == "" {
		gamePort = "10000"
	}
	if masterURL == "" {
		masterURL = "http://localhost:8080"
	}

	// Setup file logging
	setupLogging(gameID)

	log.Println("Starting DBD Game Server...")
	log.Printf("Game ID: %d, Port: %s, Killer ID: %d", gameID, gamePort, killerID)

	g := game.NewGame(gameID, masterURL, killerID)

	// Setup endpoint - master server calls this to setup players
	http.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		var players []PlayerSetup
		if err := json.NewDecoder(r.Body).Decode(&players); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		for _, p := range players {
			g.AddPlayer(p.UserID, p.Username, p.Role)
			log.Printf("Added player %s (%d) as %s", p.Username, p.UserID, p.Role)
		}

		// Start game loop in goroutine
		go g.Run()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"game_id":   gameID,
			"game_over": g.IsGameOver(),
			"players":   g.PlayerCount(),
		})
	})

	// WebSocket endpoint
	wsHandler := ws.NewHandler(g)
	http.HandleFunc("/ws", wsHandler.HandleWS)

	addr := fmt.Sprintf(":%s", gamePort)
	log.Printf("Game server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func setupLogging(gameID int) {
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/logs"
	}

	// Create log directory if it doesn't exist
	os.MkdirAll(logDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("%s/game_%d_%s.log", logDir, gameID, timestamp)

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("WARNING: Could not create log file %s: %v (logging to stdout only)", logFileName, err)
		return
	}

	// Write to both stdout and the log file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	log.Printf("Log file: %s", logFileName)
}
