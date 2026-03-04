package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"dbd-master/internal/auth"
	"dbd-master/internal/lobby"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LobbyHandler struct {
	lobbyMgr *lobby.Manager
}

func NewLobbyHandler(lobbyMgr *lobby.Manager) *LobbyHandler {
	return &LobbyHandler{lobbyMgr: lobbyMgr}
}

type wsMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func (h *LobbyHandler) HandleLobbyWS(w http.ResponseWriter, r *http.Request) {
	// Get lobby ID from URL
	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id required", http.StatusBadRequest)
		return
	}

	// Get token from query
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	l := h.lobbyMgr.GetLobby(lobbyID)
	if l == nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	player, err := l.AddPlayer(claims.UserID, claims.Username)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","payload":"lobby full"}`))
		conn.Close()
		return
	}

	log.Printf("Player %s joined lobby %s", claims.Username, lobbyID)

	// Broadcast player joined
	h.broadcastLobbyState(l)

	// Start write pump
	go h.writePump(conn, player)

	// Read pump
	h.readPump(conn, l, player, claims.UserID)
}

func (h *LobbyHandler) readPump(conn *websocket.Conn, l *lobby.Lobby, player *lobby.Player, userID int) {
	defer func() {
		l.RemovePlayer(userID)
		conn.Close()

		// If lobby is empty, remove it
		if l.PlayerCount() == 0 {
			h.lobbyMgr.RemoveLobby(l.ID)
			log.Printf("Lobby %s removed (empty)", l.ID)
		} else {
			h.broadcastLobbyState(l)
		}
	}()

	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("ws error: %v", err)
			}
			break
		}

		var msg wsMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "ready":
			var ready bool
			json.Unmarshal(msg.Payload, &ready)
			l.SetReady(userID, ready)
			h.broadcastLobbyState(l)

			// Check if all ready
			if l.AllReady() && l.OnStart != nil {
				l.OnStart(l)
			}

		case "ping":
			resp, _ := json.Marshal(wsMessage{Type: "pong"})
			select {
			case player.Send <- resp:
			default:
			}
		}
	}
}

func (h *LobbyHandler) writePump(conn *websocket.Conn, player *lobby.Player) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-player.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			conn.WriteMessage(websocket.TextMessage, msg)

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-player.Done:
			return
		}
	}
}

func (h *LobbyHandler) broadcastLobbyState(l *lobby.Lobby) {
	state := l.ToModel()
	msg, _ := json.Marshal(wsMessage{
		Type:    "lobby_state",
		Payload: mustMarshal(state),
	})
	l.Broadcast(msg)
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// HandleGameWS proxies WebSocket connections to game server containers
func HandleGameWS(w http.ResponseWriter, r *http.Request, gamePort int, hostIP string) {
	// Get token from query
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	_, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Connect to game server
	gameURL := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", hostIP, gamePort),
		Path:   "/ws",
	}

	// Upgrade client connection
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("client ws upgrade error: %v", err)
		return
	}
	defer clientConn.Close()

	// Connect to game server
	gameConn, _, err := websocket.DefaultDialer.Dial(gameURL.String(), nil)
	if err != nil {
		log.Printf("game server dial error: %v", err)
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","payload":"game server unavailable"}`))
		return
	}
	defer gameConn.Close()

	// Proxy messages bidirectionally
	done := make(chan struct{})

	// Client -> Game server
	go func() {
		defer close(done)
		for {
			msgType, msg, err := clientConn.ReadMessage()
			if err != nil {
				return
			}
			if err := gameConn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}()

	// Game server -> Client
	for {
		select {
		case <-done:
			return
		default:
			msgType, msg, err := gameConn.ReadMessage()
			if err != nil {
				return
			}
			if err := clientConn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}
}

// HandleGameReportProxy forwards game reports to master
func HandleGameReportProxy(masterURL string, body io.Reader) error {
	_, err := http.Post(masterURL+"/api/internal/game-report", "application/json", body)
	return err
}
