package models

import (
	"testing"
	"time"
)

func TestValidateGroupName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid name", "My Group", true},
		{"empty name", "", false},
		{"too short", "A", false},
		{"too long", string(make([]byte, 101)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateGroupName(tt.input)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateGroupName(%q) = %v, want %v", tt.input, result.IsValid, tt.expected)
			}
		})
	}
}

func TestValidateMemberName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid name", "John Doe", true},
		{"empty name", "", false},
		{"too long", string(make([]byte, 51)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMemberName(tt.input)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateMemberName(%q) = %v, want %v", tt.input, result.IsValid, tt.expected)
			}
		})
	}
}

func TestValidateItemTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid title", "Visit Paris", true},
		{"empty title", "", false},
		{"too long", string(make([]byte, 201)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateItemTitle(tt.input)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateItemTitle(%q) = %v, want %v", tt.input, result.IsValid, tt.expected)
			}
		})
	}
}

func TestValidateItemDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected bool
	}{
		{"valid description", stringPtr("A nice description"), true},
		{"nil description", nil, true},
		{"too long", stringPtr(string(make([]byte, 1001))), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateItemDescription(tt.input)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateItemDescription(%v) = %v, want %v", tt.input, result.IsValid, tt.expected)
			}
		})
	}
}

func TestValidateDeadline(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name     string
		input    *time.Time
		expected bool
	}{
		{"future deadline", &future, true},
		{"nil deadline", nil, true},
		{"past deadline", &past, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDeadline(tt.input)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateDeadline(%v) = %v, want %v", tt.input, result.IsValid, tt.expected)
			}
		})
	}
}

func TestCreateGroupRequestValidate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name     string
		request  CreateGroupRequest
		expected bool
	}{
		{
			"valid request",
			CreateGroupRequest{Name: "My Group", Deadline: &future},
			true,
		},
		{
			"invalid name and deadline",
			CreateGroupRequest{Name: "", Deadline: &past},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.Validate()
			if result.IsValid != tt.expected {
				t.Errorf("CreateGroupRequest.Validate() = %v, want %v", result.IsValid, tt.expected)
			}
		})
	}
}

func TestJoinGroupRequestValidate(t *testing.T) {
	tests := []struct {
		name     string
		request  JoinGroupRequest
		expected bool
	}{
		{"valid request", JoinGroupRequest{MemberName: "John Doe"}, true},
		{"invalid request", JoinGroupRequest{MemberName: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.Validate()
			if result.IsValid != tt.expected {
				t.Errorf("JoinGroupRequest.Validate() = %v, want %v", result.IsValid, tt.expected)
			}
		})
	}
}

func TestCreateItemRequestValidate(t *testing.T) {
	tests := []struct {
		name     string
		request  CreateItemRequest
		expected bool
	}{
		{
			"valid request",
			CreateItemRequest{Title: "Visit Paris", Description: stringPtr("A wonderful trip"), MemberID: "member-123"},
			true,
		},
		{
			"invalid request",
			CreateItemRequest{Title: "", Description: stringPtr(string(make([]byte, 1001))), MemberID: ""},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.Validate()
			if result.IsValid != tt.expected {
				t.Errorf("CreateItemRequest.Validate() = %v, want %v", result.IsValid, tt.expected)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trim whitespace", "  hello world  ", "hello world"},
		{"normalize spaces", "hello    world", "hello world"},
		{"mixed whitespace", "  hello   world  ", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGroupSanitize(t *testing.T) {
	group := &Group{Name: "  My   Group  "}
	group.Sanitize()
	if group.Name != "My Group" {
		t.Errorf("Group.Sanitize() failed, got %q, want %q", group.Name, "My Group")
	}
}

func TestMemberSanitize(t *testing.T) {
	member := &Member{Name: "  John   Doe  "}
	member.Sanitize()
	if member.Name != "John Doe" {
		t.Errorf("Member.Sanitize() failed, got %q, want %q", member.Name, "John Doe")
	}
}

func TestBucketListItemSanitize(t *testing.T) {
	item := &BucketListItem{
		Title:       "  Visit   Paris  ",
		Description: stringPtr("  A   wonderful   trip  "),
	}
	item.Sanitize()
	if item.Title != "Visit Paris" {
		t.Errorf("BucketListItem.Sanitize() failed for title, got %q, want %q", item.Title, "Visit Paris")
	}
	if item.Description == nil || *item.Description != "A wonderful trip" {
		t.Errorf("BucketListItem.Sanitize() failed for description, got %v, want %q", item.Description, "A wonderful trip")
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}