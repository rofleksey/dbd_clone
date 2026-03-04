package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"dbd-master/internal/auth"
	"dbd-master/internal/db"
	"dbd-master/internal/docker"
	"dbd-master/internal/lobby"
	"dbd-master/internal/models"
	"dbd-master/internal/ws"
)

type Router struct {
	mux        *http.ServeMux
	lobbyMgr   *lobby.Manager
	dockerMgr  *docker.Manager
	wsHandler  *ws.LobbyHandler
	gameStates map[int]*models.GameProgress
	hostIP     string // IP/hostname to reach game server containers from master
}

func New(lobbyMgr *lobby.Manager, dockerMgr *docker.Manager) *Router {
	hostIP := os.Getenv("HOST_IP")
	if hostIP == "" {
		hostIP = "host.docker.internal"
	}

	r := &Router{
		mux:        http.NewServeMux(),
		lobbyMgr:   lobbyMgr,
		dockerMgr:  dockerMgr,
		wsHandler:  ws.NewLobbyHandler(lobbyMgr),
		gameStates: make(map[int]*models.GameProgress),
		hostIP:     hostIP,
	}
	r.setupRoutes()
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	r.mux.ServeHTTP(w, req)
}

func (r *Router) setupRoutes() {
	// Auth routes
	r.mux.HandleFunc("POST /api/auth/register", r.handleRegister)
	r.mux.HandleFunc("POST /api/auth/login", r.handleLogin)

	// Lobby routes
	r.mux.HandleFunc("GET /api/lobbies", r.authMiddleware(r.handleListLobbies))
	r.mux.HandleFunc("POST /api/lobbies", r.authMiddleware(r.handleCreateLobby))

	// Leaderboard routes
	r.mux.HandleFunc("GET /api/leaderboard", r.handleLeaderboard)
	r.mux.HandleFunc("GET /api/stats/{username}", r.handlePlayerStats)

	// Active games
	r.mux.HandleFunc("GET /api/games", r.handleActiveGames)

	// Internal game server routes
	r.mux.HandleFunc("POST /api/internal/game-report", r.handleGameReport)
	r.mux.HandleFunc("POST /api/internal/game-progress", r.handleGameProgress)

	// WebSocket routes
	r.mux.HandleFunc("/ws/lobby/{id}", r.wsHandler.HandleLobbyWS)
	r.mux.HandleFunc("/ws/game/{id}", r.handleGameWS)
}

// Auth middleware
func (r *Router) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authorization required"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ValidateToken(tokenStr)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid token"})
			return
		}

		ctx := context.WithValue(req.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		next(w, req.WithContext(ctx))
	}
}

// Auth handlers
func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	var input models.RegisterRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request"})
		return
	}

	if len(input.Username) < 3 || len(input.Username) > 32 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "username must be 3-32 characters"})
		return
	}
	if len(input.Password) < 4 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "password must be at least 4 characters"})
		return
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal error"})
		return
	}

	user, err := db.CreateUser(input.Username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "username already taken"})
			return
		}
		log.Printf("create user error: %v", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal error"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, models.AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
	})
}

func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	var input models.LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request"})
		return
	}

	user, err := db.GetUserByUsername(input.Username)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	if !auth.CheckPassword(input.Password, user.PasswordHash) {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, models.AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
	})
}

// Lobby handlers
func (r *Router) handleListLobbies(w http.ResponseWriter, req *http.Request) {
	lobbies := r.lobbyMgr.ListLobbies()
	if lobbies == nil {
		lobbies = []models.Lobby{}
	}
	writeJSON(w, http.StatusOK, lobbies)
}

func (r *Router) handleCreateLobby(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("user_id").(int)
	username := req.Context().Value("username").(string)

	var input models.CreateLobbyRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request"})
		return
	}

	if input.Name == "" {
		input.Name = username + "'s lobby"
	}

	l := r.lobbyMgr.CreateLobby(input.Name, userID, username)

	// Set the game start callback
	l.OnStart = func(lby *lobby.Lobby) {
		r.startGame(lby)
	}

	writeJSON(w, http.StatusCreated, l.ToModel())
}

// Game start logic
func (r *Router) startGame(l *lobby.Lobby) {
	killerID := l.ChooseRandomKiller()
	playerIDs := l.GetPlayerIDs()
	players := l.GetPlayers()

	log.Printf("Starting game from lobby %s with %d players, killer: %d", l.ID, len(playerIDs), killerID)

	// Create game record in DB
	game, err := db.CreateGame(killerID, "", 0)
	if err != nil {
		log.Printf("failed to create game record: %v", err)
		return
	}

	// Add players to game record
	for _, pid := range playerIDs {
		role := "survivor"
		if pid == killerID {
			role = "killer"
		}
		db.AddGamePlayer(game.ID, pid, role)
	}

	// Start game server container
	containerID, port, err := r.dockerMgr.StartGameServer(context.Background(), game.ID, killerID, len(playerIDs))
	if err != nil {
		log.Printf("failed to start game server: %v", err)
		db.CancelGame(game.ID)

		// Notify lobby of failure
		msg, _ := json.Marshal(map[string]interface{}{
			"type":    "game_error",
			"payload": "Failed to start game server",
		})
		l.Broadcast(msg)
		return
	}

	// Persist port so /ws/game/{id} can reach the game server
	if err := db.UpdateGamePortAndContainer(game.ID, port, containerID); err != nil {
		log.Printf("failed to update game port in DB: %v", err)
	}

	log.Printf("Game %d started on port %d, container %s", game.ID, port, containerID[:12])

	// Build player info for game start message
	type playerInfo struct {
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	playerInfos := make([]playerInfo, 0, len(playerIDs))
	for _, pid := range playerIDs {
		role := "survivor"
		if pid == killerID {
			role = "killer"
		}
		pname := ""
		if p, ok := players[pid]; ok {
			pname = p.Username
		}
		playerInfos = append(playerInfos, playerInfo{
			UserID:   pid,
			Username: pname,
			Role:     role,
		})
	}

	// Wait for game server to be ready, then send player setup
	go func() {
		setupURL := fmt.Sprintf("http://%s:%d/setup", r.hostIP, port)

		// Retry setup call until game server is ready (max 10 seconds)
		var setupErr error
		for attempt := 0; attempt < 20; attempt++ {
			time.Sleep(500 * time.Millisecond)

			data, _ := json.Marshal(playerInfos)
			resp, err := http.Post(setupURL, "application/json", bytes.NewReader(data))
			if err != nil {
				setupErr = err
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				log.Printf("Game %d setup complete", game.ID)
				setupErr = nil
				break
			}
			setupErr = fmt.Errorf("setup returned status %d", resp.StatusCode)
		}

		if setupErr != nil {
			log.Printf("Failed to setup game %d: %v", game.ID, setupErr)
			db.CancelGame(game.ID)
			r.dockerMgr.StopGameServer(context.Background(), game.ID)

			errMsg, _ := json.Marshal(map[string]interface{}{
				"type":    "game_error",
				"payload": "Failed to initialize game server",
			})
			l.Broadcast(errMsg)
			return
		}

		// Notify all lobby players that game is starting
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "game_start",
			"payload": map[string]interface{}{
				"game_id":   game.ID,
				"port":      port,
				"killer_id": killerID,
				"players":   playerInfos,
			},
		})
		l.Broadcast(msg)

		// Remove lobby
		r.lobbyMgr.RemoveLobby(l.ID)
	}()
}

// Leaderboard handlers
func (r *Router) handleLeaderboard(w http.ResponseWriter, req *http.Request) {
	limitStr := req.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	stats, err := db.GetLeaderboard(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal error"})
		return
	}
	if stats == nil {
		stats = []models.PlayerStats{}
	}

	writeJSON(w, http.StatusOK, stats)
}

func (r *Router) handlePlayerStats(w http.ResponseWriter, req *http.Request) {
	username := req.PathValue("username")
	if username == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "username required"})
		return
	}

	stats, err := db.GetStatsByUsername(username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "player not found"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Active games handler
func (r *Router) handleActiveGames(w http.ResponseWriter, req *http.Request) {
	games, err := db.GetActiveGames()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal error"})
		return
	}
	if games == nil {
		games = []models.Game{}
	}

	// Augment with live progress data
	type gameWithProgress struct {
		models.Game
		Progress *models.GameProgress `json:"progress,omitempty"`
	}

	result := make([]gameWithProgress, len(games))
	for i, g := range games {
		result[i] = gameWithProgress{Game: g}
		if progress, ok := r.gameStates[g.ID]; ok {
			result[i].Progress = progress
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// Internal handlers
func (r *Router) handleGameReport(w http.ResponseWriter, req *http.Request) {
	var report models.GameReport
	if err := json.NewDecoder(req.Body).Decode(&report); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request"})
		return
	}

	log.Printf("Game %d ended with result: %s", report.GameID, report.Result)

	// End game
	db.EndGame(report.GameID, report.Result)

	// Update player stats
	for _, p := range report.Players {
		won := false
		if p.Role == "killer" && (report.Result == "killer_win") {
			won = true
		} else if p.Role == "survivor" && p.Survived {
			won = true
		}

		db.UpdateGamePlayer(report.GameID, p.UserID, p.Survived, p.Kills, p.GensDone)
		db.UpdatePlayerStats(p, won)
	}

	// Clean up docker container
	r.dockerMgr.StopGameServer(context.Background(), report.GameID)

	// Remove progress tracking
	delete(r.gameStates, report.GameID)

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleGameProgress(w http.ResponseWriter, req *http.Request) {
	var progress models.GameProgress
	if err := json.NewDecoder(req.Body).Decode(&progress); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request"})
		return
	}

	r.gameStates[progress.GameID] = &progress
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Game WS proxy
func (r *Router) handleGameWS(w http.ResponseWriter, req *http.Request) {
	gameIDStr := req.PathValue("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "invalid game id", http.StatusBadRequest)
		return
	}

	game, err := db.GetGameByID(gameID)
	if err != nil || game == nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	ws.HandleGameWS(w, req, game.Port, r.hostIP)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}

func (r *Router) GetGamePort(gameID int) (int, error) {
	game, err := db.GetGameByID(gameID)
	if err != nil || game == nil {
		return 0, fmt.Errorf("game not found")
	}
	return game.Port, nil
}
