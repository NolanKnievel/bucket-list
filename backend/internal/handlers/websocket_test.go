package handlers

import (
	"collaborative-bucket-list/internal/websocket"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewWebSocketHandler(t *testing.T) {
	hub := websocket.NewHub()
	handler := NewWebSocketHandler(hub)
	
	assert.NotNil(t, handler)
	assert.Equal(t, hub, handler.hub)
}

func TestWebSocketHandler_GetRoomStats(t *testing.T) {
	hub := websocket.NewHub()
	handler := NewWebSocketHandler(hub)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws/rooms/:id/stats", handler.GetRoomStats)
	
	// Test with valid room ID
	req, _ := http.NewRequest("GET", "/ws/rooms/test-room/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "roomId")
	assert.Contains(t, w.Body.String(), "clientCount")
}

func TestWebSocketHandler_GetRoomStats_MissingRoomID(t *testing.T) {
	hub := websocket.NewHub()
	handler := NewWebSocketHandler(hub)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws/rooms/stats", handler.GetRoomStats)
	
	// Test without room ID
	req, _ := http.NewRequest("GET", "/ws/rooms/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_ROOM_ID")
}

func TestWebSocketHandler_GetAllRoomStats(t *testing.T) {
	hub := websocket.NewHub()
	handler := NewWebSocketHandler(hub)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws/rooms/stats", handler.GetAllRoomStats)
	
	req, _ := http.NewRequest("GET", "/ws/rooms/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "activeRooms")
	assert.Contains(t, w.Body.String(), "totalRooms")
}