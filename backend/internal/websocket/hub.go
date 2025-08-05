package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients grouped by room (group ID)
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex to protect concurrent access to rooms
	mutex sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub and handles client registration, unregistration, and broadcasting
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the appropriate room
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*Client]bool)
	}
	h.rooms[client.roomID][client] = true

	log.Printf("Client registered to room %s. Room now has %d clients", 
		client.roomID, len(h.rooms[client.roomID]))
}

// unregisterClient removes a client from its room and closes the connection
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, exists := h.rooms[client.roomID]; exists {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)

			// Remove the room if it's empty
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
				log.Printf("Room %s removed (no clients remaining)", client.roomID)
			} else {
				log.Printf("Client unregistered from room %s. Room now has %d clients", 
					client.roomID, len(clients))
			}
		}
	}
}

// broadcastMessage sends a message to all clients in the appropriate room
func (h *Hub) broadcastMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error unmarshaling broadcast message: %v", err)
		return
	}

	h.mutex.RLock()
	clients, exists := h.rooms[msg.RoomID]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("No room found for ID: %s", msg.RoomID)
		return
	}

	// Broadcast to all clients in the room
	for client := range clients {
		select {
		case client.send <- message:
		default:
			// Client's send channel is full, close it and remove from room
			h.unregisterClient(client)
		}
	}

	log.Printf("Broadcasted message to %d clients in room %s", len(clients), msg.RoomID)
}

// BroadcastToRoom sends a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(roomID string, messageType string, data interface{}) {
	message := Message{
		Type:   messageType,
		RoomID: roomID,
		Data:   data,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.broadcast <- messageBytes
}

// GetRoomClientCount returns the number of clients in a specific room
func (h *Hub) GetRoomClientCount(roomID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, exists := h.rooms[roomID]; exists {
		return len(clients)
	}
	return 0
}

// GetActiveRooms returns a list of all active room IDs
func (h *Hub) GetActiveRooms() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	rooms := make([]string, 0, len(h.rooms))
	for roomID := range h.rooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}