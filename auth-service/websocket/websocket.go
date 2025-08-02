package websocket

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsConnections = struct {
	sync.RWMutex
	conns map[string]*websocket.Conn
}{conns: make(map[string]*websocket.Conn)}

func RegisterConnection(userID string, conn *websocket.Conn) {
	wsConnections.Lock()
	defer wsConnections.Unlock()
	wsConnections.conns[userID] = conn
}

func GetConnection(userID string) *websocket.Conn {
	wsConnections.RLock()
	defer wsConnections.RUnlock()
	return wsConnections.conns[userID]
}

func RemoveConnection(userID string) {
	wsConnections.Lock()
	defer wsConnections.Unlock()
	delete(wsConnections.conns, userID)
}

func WebSocketHander(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return
	}

	defer ws.Close()

	RegisterConnection(userID, ws)
	defer RemoveConnection(userID)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func NotifyRoleChange(userID, newRole, newAccessTOken, newRefreshToken string) error {
	conn := GetConnection(userID)
	if conn == nil {
		return nil
	}

	message := map[string]interface{}{
		"event":         "role_change",
		"role":          newRole,
		"access_token":  newAccessTOken,
		"refresh_token": newRefreshToken,
		"timestamp":     time.Now().Unix(),
	}

	err := conn.WriteJSON(message)
	if err != nil {
		RemoveConnection(userID)
		return err
	}

	return nil
}
