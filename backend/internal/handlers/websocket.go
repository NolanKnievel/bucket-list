package handlers

import (
	"collaborative-bucket-list/internal/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *websocket.Hub
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get room ID (group ID) from URL parameter
	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_ROOM_ID",
				"message": "Room ID is required",
			},
		})
		return
	}

	// Get member ID from query parameter
	memberID := c.Query("memberId")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_MEMBER_ID",
				"message": "Member ID is required",
			},
		})
		return
	}

	// Upgrade the HTTP connection to WebSocket
	websocket.ServeWS(h.hub, c.Writer, c.Request, roomID, memberID)
}

// GetRoomStats returns statistics about active rooms and connections
func (h *WebSocketHandler) GetRoomStats(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_ROOM_ID",
				"message": "Room ID is required",
			},
		})
		return
	}

	clientCount := h.hub.GetRoomClientCount(roomID)
	
	c.JSON(http.StatusOK, gin.H{
		"roomId":      roomID,
		"clientCount": clientCount,
	})
}

// GetAllRoomStats returns statistics about all active rooms
func (h *WebSocketHandler) GetAllRoomStats(c *gin.Context) {
	activeRooms := h.hub.GetActiveRooms()
	roomStats := make([]gin.H, 0, len(activeRooms))

	for _, roomID := range activeRooms {
		clientCount := h.hub.GetRoomClientCount(roomID)
		roomStats = append(roomStats, gin.H{
			"roomId":      roomID,
			"clientCount": clientCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"activeRooms": roomStats,
		"totalRooms":  len(activeRooms),
	})
}