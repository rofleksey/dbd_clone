package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type PlayerStats struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	GamesPlayed     int    `json:"games_played"`
	GamesWon        int    `json:"games_won"`
	Kills           int    `json:"kills"`
	Escapes         int    `json:"escapes"`
	GeneratorsDone  int    `json:"generators_done"`
	SurvivorsHooked int    `json:"survivors_hooked"`
	GamesAsKiller   int    `json:"games_as_killer"`
	GamesAsSurvivor int    `json:"games_as_survivor"`
}

type Game struct {
	ID          int          `json:"id"`
	StartedAt   time.Time    `json:"started_at"`
	EndedAt     *time.Time   `json:"ended_at,omitempty"`
	Status      string       `json:"status"`
	KillerID    int          `json:"killer_id"`
	KillerName  string       `json:"killer_name,omitempty"`
	ContainerID string       `json:"-"`
	Port        int          `json:"port"`
	Result      string       `json:"result,omitempty"`
	Players     []GamePlayer `json:"players,omitempty"`
}

type GamePlayer struct {
	GameID   int    `json:"game_id"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Survived bool   `json:"survived"`
	Kills    int    `json:"kills"`
	GensDone int    `json:"gens_done"`
}

type Lobby struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	HostID     int           `json:"host_id"`
	HostName   string        `json:"host_name"`
	Players    []LobbyPlayer `json:"players"`
	MaxPlayers int           `json:"max_players"`
	CreatedAt  time.Time     `json:"created_at"`
}

type LobbyPlayer struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Ready    bool   `json:"ready"`
	Ping     int    `json:"ping"`
}

// API request/response types

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

type CreateLobbyRequest struct {
	Name string `json:"name"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// WebSocket message types for lobby
type LobbyMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Game report from game server to master
type GameReport struct {
	GameID  int            `json:"game_id"`
	Result  string         `json:"result"`
	Players []PlayerReport `json:"players"`
}

type PlayerReport struct {
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
	Survived bool   `json:"survived"`
	Kills    int    `json:"kills"`
	GensDone int    `json:"gens_done"`
}

// Game progress for live games list
type GameProgress struct {
	GameID         int      `json:"game_id"`
	GensCompleted  int      `json:"gens_completed"`
	SurvivorsAlive int      `json:"survivors_alive"`
	ElapsedSeconds float64  `json:"elapsed_seconds"`
	PlayerNames    []string `json:"player_names"`
	KillerName     string   `json:"killer_name"`
}
