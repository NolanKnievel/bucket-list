package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"collaborative-bucket-list/internal/models"
	"collaborative-bucket-list/internal/repositories"
)

// HubInterface defines the interface for WebSocket hub operations
type HubInterface interface {
	BroadcastToRoom(roomID string, messageType string, data interface{})
}

// EventHandler handles WebSocket events and business logic
type EventHandler struct {
	hub   HubInterface
	repos repositories.RepositoryManager
}

// NewEventHandler creates a new WebSocket event handler
func NewEventHandler(hub HubInterface, repos repositories.RepositoryManager) *EventHandler {
	return &EventHandler{
		hub:   hub,
		repos: repos,
	}
}

// WebSocket event types
const (
	// Client to Server events
	EventJoinGroup        = "join-group"
	EventAddItem          = "add-item"
	EventToggleCompletion = "toggle-completion"

	// Server to Client events
	EventMemberJoined = "member-joined"
	EventItemAdded    = "item-added"
	EventItemUpdated  = "item-updated"
	EventError        = "error"
)

// Event payload structures
type JoinGroupPayload struct {
	GroupID  string `json:"groupId"`
	MemberID string `json:"memberId"`
}

type AddItemPayload struct {
	GroupID string                    `json:"groupId"`
	Item    models.CreateItemRequest  `json:"item"`
}

type ToggleCompletionPayload struct {
	GroupID   string `json:"groupId"`
	ItemID    string `json:"itemId"`
	Completed bool   `json:"completed"`
	MemberID  string `json:"memberId"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProcessMessage processes incoming WebSocket messages and routes them to appropriate handlers
func (eh *EventHandler) ProcessMessage(client *Client, messageBytes []byte) {
	var msg Message
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		log.Printf("Error unmarshaling WebSocket message: %v", err)
		eh.sendError(client, "INVALID_MESSAGE", "Invalid message format", err.Error())
		return
	}

	// Validate that the message is for the correct room
	if msg.RoomID != client.roomID {
		log.Printf("Message room ID mismatch: expected %s, got %s", client.roomID, msg.RoomID)
		eh.sendError(client, "ROOM_MISMATCH", "Message room ID does not match client room", "")
		return
	}

	// Validate that the message is from the correct member
	if msg.MemberID != client.memberID {
		log.Printf("Message member ID mismatch: expected %s, got %s", client.memberID, msg.MemberID)
		eh.sendError(client, "MEMBER_MISMATCH", "Message member ID does not match client member", "")
		return
	}

	ctx := context.Background()

	// Route message to appropriate handler
	switch msg.Type {
	case EventJoinGroup:
		eh.handleJoinGroup(ctx, client, msg.Data)
	case EventAddItem:
		eh.handleAddItem(ctx, client, msg.Data)
	case EventToggleCompletion:
		eh.handleToggleCompletion(ctx, client, msg.Data)
	default:
		log.Printf("Unknown WebSocket event type: %s", msg.Type)
		eh.sendError(client, "UNKNOWN_EVENT", "Unknown event type", msg.Type)
	}
}

// handleJoinGroup handles join-group events
func (eh *EventHandler) handleJoinGroup(ctx context.Context, client *Client, data interface{}) {
	var payload JoinGroupPayload
	if err := eh.parsePayload(data, &payload); err != nil {
		eh.sendError(client, "INVALID_PAYLOAD", "Invalid join-group payload", err.Error())
		return
	}

	// Validate group ID matches client room
	if payload.GroupID != client.roomID {
		eh.sendError(client, "GROUP_MISMATCH", "Group ID does not match client room", "")
		return
	}

	// Validate member ID matches client member
	if payload.MemberID != client.memberID {
		eh.sendError(client, "MEMBER_MISMATCH", "Member ID does not match client member", "")
		return
	}

	// Verify the member exists and belongs to the group
	member, err := eh.repos.Members().GetByID(ctx, payload.MemberID)
	if err != nil {
		log.Printf("Error fetching member %s: %v", payload.MemberID, err)
		eh.sendError(client, "MEMBER_NOT_FOUND", "Member not found", "")
		return
	}

	if member.GroupID != payload.GroupID {
		log.Printf("Member %s does not belong to group %s", payload.MemberID, payload.GroupID)
		eh.sendError(client, "MEMBER_GROUP_MISMATCH", "Member does not belong to this group", "")
		return
	}

	// Broadcast member-joined event to all clients in the room
	eh.hub.BroadcastToRoom(payload.GroupID, EventMemberJoined, member)
	log.Printf("Member %s joined group %s via WebSocket", member.Name, payload.GroupID)
}

// handleAddItem handles add-item events
func (eh *EventHandler) handleAddItem(ctx context.Context, client *Client, data interface{}) {
	var payload AddItemPayload
	if err := eh.parsePayload(data, &payload); err != nil {
		eh.sendError(client, "INVALID_PAYLOAD", "Invalid add-item payload", err.Error())
		return
	}

	// Validate group ID matches client room
	if payload.GroupID != client.roomID {
		eh.sendError(client, "GROUP_MISMATCH", "Group ID does not match client room", "")
		return
	}

	// Validate member ID matches client member
	if payload.Item.MemberID != client.memberID {
		eh.sendError(client, "MEMBER_MISMATCH", "Member ID does not match client member", "")
		return
	}

	// Sanitize and validate the item request
	payload.Item.Sanitize()
	if validation := payload.Item.Validate(); !validation.IsValid {
		eh.sendError(client, "VALIDATION_ERROR", "Invalid item data", validation.Errors[0].Message)
		return
	}

	// Verify the member exists and belongs to the group
	member, err := eh.repos.Members().GetByID(ctx, payload.Item.MemberID)
	if err != nil {
		log.Printf("Error fetching member %s: %v", payload.Item.MemberID, err)
		eh.sendError(client, "MEMBER_NOT_FOUND", "Member not found", "")
		return
	}

	if member.GroupID != payload.GroupID {
		log.Printf("Member %s does not belong to group %s", payload.Item.MemberID, payload.GroupID)
		eh.sendError(client, "MEMBER_GROUP_MISMATCH", "Member does not belong to this group", "")
		return
	}

	// Create the bucket list item
	item := &models.BucketListItem{
		GroupID:     payload.GroupID,
		Title:       payload.Item.Title,
		Description: payload.Item.Description,
		Completed:   false,
		CreatedBy:   payload.Item.MemberID,
		CreatedAt:   time.Now(),
	}

	if err := eh.repos.BucketItems().Create(ctx, item); err != nil {
		log.Printf("Error creating bucket item: %v", err)
		eh.sendError(client, "CREATE_FAILED", "Failed to create item", err.Error())
		return
	}

	// Broadcast item-added event to all clients in the room
	eh.hub.BroadcastToRoom(payload.GroupID, EventItemAdded, item)
	log.Printf("Item '%s' added to group %s by member %s", item.Title, payload.GroupID, member.Name)
}

// handleToggleCompletion handles toggle-completion events
func (eh *EventHandler) handleToggleCompletion(ctx context.Context, client *Client, data interface{}) {
	var payload ToggleCompletionPayload
	if err := eh.parsePayload(data, &payload); err != nil {
		eh.sendError(client, "INVALID_PAYLOAD", "Invalid toggle-completion payload", err.Error())
		return
	}

	// Validate group ID matches client room
	if payload.GroupID != client.roomID {
		eh.sendError(client, "GROUP_MISMATCH", "Group ID does not match client room", "")
		return
	}

	// Validate member ID matches client member
	if payload.MemberID != client.memberID {
		eh.sendError(client, "MEMBER_MISMATCH", "Member ID does not match client member", "")
		return
	}

	// Verify the member exists and belongs to the group
	member, err := eh.repos.Members().GetByID(ctx, payload.MemberID)
	if err != nil {
		log.Printf("Error fetching member %s: %v", payload.MemberID, err)
		eh.sendError(client, "MEMBER_NOT_FOUND", "Member not found", "")
		return
	}

	if member.GroupID != payload.GroupID {
		log.Printf("Member %s does not belong to group %s", payload.MemberID, payload.GroupID)
		eh.sendError(client, "MEMBER_GROUP_MISMATCH", "Member does not belong to this group", "")
		return
	}

	// Verify the item exists and belongs to the group
	item, err := eh.repos.BucketItems().GetByID(ctx, payload.ItemID)
	if err != nil {
		log.Printf("Error fetching item %s: %v", payload.ItemID, err)
		eh.sendError(client, "ITEM_NOT_FOUND", "Item not found", "")
		return
	}

	if item.GroupID != payload.GroupID {
		log.Printf("Item %s does not belong to group %s", payload.ItemID, payload.GroupID)
		eh.sendError(client, "ITEM_GROUP_MISMATCH", "Item does not belong to this group", "")
		return
	}

	// Toggle the completion status
	if err := eh.repos.BucketItems().ToggleCompletion(ctx, payload.ItemID, payload.MemberID, payload.Completed); err != nil {
		log.Printf("Error toggling item completion: %v", err)
		eh.sendError(client, "UPDATE_FAILED", "Failed to update item", err.Error())
		return
	}

	// Fetch the updated item
	updatedItem, err := eh.repos.BucketItems().GetByID(ctx, payload.ItemID)
	if err != nil {
		log.Printf("Error fetching updated item %s: %v", payload.ItemID, err)
		eh.sendError(client, "FETCH_FAILED", "Failed to fetch updated item", err.Error())
		return
	}

	// Broadcast item-updated event to all clients in the room
	eh.hub.BroadcastToRoom(payload.GroupID, EventItemUpdated, updatedItem)
	
	completionStatus := "incomplete"
	if payload.Completed {
		completionStatus = "complete"
	}
	log.Printf("Item '%s' marked as %s in group %s by member %s", updatedItem.Title, completionStatus, payload.GroupID, member.Name)
}

// parsePayload parses WebSocket event payload data
func (eh *EventHandler) parsePayload(data interface{}, target interface{}) error {
	// Convert data to JSON bytes and then unmarshal to target struct
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(dataBytes, target)
}

// sendError sends an error message to a specific client
func (eh *EventHandler) sendError(client *Client, code, message, details string) {
	errorPayload := ErrorPayload{
		Code:    code,
		Message: message,
		Details: details,
	}

	errorMessage := Message{
		Type:     EventError,
		RoomID:   client.roomID,
		MemberID: client.memberID,
		Data:     errorPayload,
	}

	messageBytes, err := json.Marshal(errorMessage)
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	select {
	case client.send <- messageBytes:
		log.Printf("Sent error to client in room %s: %s - %s", client.roomID, code, message)
	default:
		log.Printf("Failed to send error to client in room %s (channel full): %s - %s", client.roomID, code, message)
	}
}