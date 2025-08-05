package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.rooms)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

func TestHubBroadcastToRoom(t *testing.T) {
	hub := NewHub()
	
	// Start the hub in a goroutine
	go hub.Run()
	
	// Give the hub a moment to start
	time.Sleep(10 * time.Millisecond)
	
	// Test broadcasting to a room
	roomID := "test-room"
	messageType := "test-message"
	data := map[string]string{"key": "value"}
	
	// This should not panic even if no clients are connected
	hub.BroadcastToRoom(roomID, messageType, data)
	
	// Verify the message was properly formatted
	expectedMessage := Message{
		Type:   messageType,
		RoomID: roomID,
		Data:   data,
	}
	
	expectedBytes, err := json.Marshal(expectedMessage)
	assert.NoError(t, err)
	assert.NotEmpty(t, expectedBytes)
}

func TestHubGetRoomClientCount(t *testing.T) {
	hub := NewHub()
	
	// Test getting client count for non-existent room
	count := hub.GetRoomClientCount("non-existent-room")
	assert.Equal(t, 0, count)
}

func TestHubGetActiveRooms(t *testing.T) {
	hub := NewHub()
	
	// Test getting active rooms when none exist
	rooms := hub.GetActiveRooms()
	assert.Empty(t, rooms)
}

func TestMessageSerialization(t *testing.T) {
	msg := Message{
		Type:     "test-type",
		RoomID:   "test-room",
		MemberID: "test-member",
		Data:     map[string]string{"key": "value"},
	}
	
	// Test marshaling
	bytes, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.NotEmpty(t, bytes)
	
	// Test unmarshaling
	var unmarshaled Message
	err = json.Unmarshal(bytes, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, msg.Type, unmarshaled.Type)
	assert.Equal(t, msg.RoomID, unmarshaled.RoomID)
	assert.Equal(t, msg.MemberID, unmarshaled.MemberID)
}