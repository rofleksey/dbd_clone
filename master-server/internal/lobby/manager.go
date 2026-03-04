package lobby

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"dbd-master/internal/models"

	"github.com/google/uuid"
)

type Manager struct {
	mu      sync.RWMutex
	lobbies map[string]*Lobby
}

type Lobby struct {
	mu         sync.RWMutex
	ID         string
	Name       string
	HostID     int
	HostName   string
	Players    map[int]*Player
	MaxPlayers int
	CreatedAt  time.Time
	OnStart    func(lobby *Lobby) // callback when game starts
}

type Player struct {
	UserID   int
	Username string
	Ready    bool
	Ping     int
	Send     chan []byte
	Done     chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		lobbies: make(map[string]*Lobby),
	}
}

func (m *Manager) CreateLobby(name string, hostID int, hostName string) *Lobby {
	m.mu.Lock()
	defer m.mu.Unlock()

	lobby := &Lobby{
		ID:         uuid.New().String()[:8],
		Name:       name,
		HostID:     hostID,
		HostName:   hostName,
		Players:    make(map[int]*Player),
		MaxPlayers: 5,
		CreatedAt:  time.Now(),
	}

	m.lobbies[lobby.ID] = lobby
	return lobby
}

func (m *Manager) GetLobby(id string) *Lobby {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lobbies[id]
}

func (m *Manager) RemoveLobby(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.lobbies, id)
}

func (m *Manager) ListLobbies() []models.Lobby {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []models.Lobby
	for _, l := range m.lobbies {
		list = append(list, l.ToModel())
	}
	return list
}

func (l *Lobby) AddPlayer(userID int, username string) (*Player, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Players) >= l.MaxPlayers {
		return nil, fmt.Errorf("lobby is full")
	}

	if _, exists := l.Players[userID]; exists {
		return nil, fmt.Errorf("already in lobby")
	}

	player := &Player{
		UserID:   userID,
		Username: username,
		Ready:    false,
		Ping:     0,
		Send:     make(chan []byte, 256),
		Done:     make(chan struct{}),
	}

	l.Players[userID] = player
	return player, nil
}

func (l *Lobby) RemovePlayer(userID int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if p, exists := l.Players[userID]; exists {
		close(p.Done)
		delete(l.Players, userID)
	}
}

func (l *Lobby) SetReady(userID int, ready bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if p, exists := l.Players[userID]; exists {
		p.Ready = ready
	}
}

func (l *Lobby) AllReady() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.Players) < 2 {
		return false
	}

	for _, p := range l.Players {
		if !p.Ready {
			return false
		}
	}
	return true
}

func (l *Lobby) PlayerCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Players)
}

func (l *Lobby) ChooseRandomKiller() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	playerIDs := make([]int, 0, len(l.Players))
	for id := range l.Players {
		playerIDs = append(playerIDs, id)
	}

	return playerIDs[rand.Intn(len(playerIDs))]
}

func (l *Lobby) GetPlayerIDs() []int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	ids := make([]int, 0, len(l.Players))
	for id := range l.Players {
		ids = append(ids, id)
	}
	return ids
}

func (l *Lobby) GetPlayers() map[int]*Player {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[int]*Player, len(l.Players))
	for k, v := range l.Players {
		result[k] = v
	}
	return result
}

func (l *Lobby) Broadcast(msg []byte) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, p := range l.Players {
		select {
		case p.Send <- msg:
		default:
		}
	}
}

func (l *Lobby) ToModel() models.Lobby {
	l.mu.RLock()
	defer l.mu.RUnlock()

	players := make([]models.LobbyPlayer, 0, len(l.Players))
	for _, p := range l.Players {
		players = append(players, models.LobbyPlayer{
			UserID:   p.UserID,
			Username: p.Username,
			Ready:    p.Ready,
			Ping:     p.Ping,
		})
	}

	return models.Lobby{
		ID:         l.ID,
		Name:       l.Name,
		HostID:     l.HostID,
		HostName:   l.HostName,
		Players:    players,
		MaxPlayers: l.MaxPlayers,
		CreatedAt:  l.CreatedAt,
	}
}
