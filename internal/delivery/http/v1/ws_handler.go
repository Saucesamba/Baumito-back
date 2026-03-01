package v1

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WsHandler struct {
	clients map[string]*websocket.Conn
	mu      sync.Mutex
}

func NewWsHandler() *WsHandler {
	return &WsHandler{
		clients: make(map[string]*websocket.Conn),
	}
}

// NotifyUser реализует интерфейс из domain
func (h *WsHandler) NotifyUser(userID string, message interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, ok := h.clients[userID]; ok {
		// Отправляем JSON напрямую в сокет студента
		_ = conn.WriteJSON(message)
	}
}

func (h *WsHandler) HandleWS(c *gin.Context) {
	userID, _ := c.Get("userId") // Берем из JWT middleware

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.clients[userID.(string)] = conn
	h.mu.Unlock()

	// Читаем из сокета, чтобы держать соединение открытым
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			h.mu.Lock()
			delete(h.clients, userID.(string))
			h.mu.Unlock()
			break
		}
	}
}
