package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"dbd-game/internal/game"
	"dbd-game/internal/protocol"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	game *game.Game
}

func NewHandler(g *game.Game) *Handler {
	return &Handler{game: g}
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	// First message should be auth
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var authMsg protocol.ClientMessage
	if err := json.Unmarshal(message, &authMsg); err != nil || authMsg.Type != "auth" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","payload":"auth required"}`))
		conn.Close()
		return
	}

	userID := authMsg.UserID
	if userID <= 0 {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","payload":"invalid user"}`))
		conn.Close()
		return
	}

	sendChan := h.game.GetSendChan(userID)
	if sendChan == nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","payload":"not in this game"}`))
		conn.Close()
		return
	}

	log.Printf("Player %d connected to game", userID)

	// Send initial OK
	conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"connected"}`))

	// Start write pump
	go h.writePump(conn, sendChan, userID)

	// Read pump
	h.readPump(conn, userID)
}

func (h *Handler) readPump(conn *websocket.Conn, userID int) {
	defer func() {
		h.game.RemovePlayer(userID)
		conn.Close()
		log.Printf("Player %d disconnected from game", userID)
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

		var msg protocol.ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		h.game.HandleInput(userID, msg)
	}
}

func (h *Handler) writePump(conn *websocket.Conn, sendChan chan []byte, userID int) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-sendChan:
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
		}
	}
}
