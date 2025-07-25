package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Group represents a bucket list group
type Group struct {
	ID        string     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Deadline  *time.Time `json:"deadline,omitempty" db:"deadline"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	CreatedBy string     `json:"createdBy" db:"created_by"`
}

// Member represents a group member
type Member struct {
	ID        string    `json:"id" db:"id"`
	GroupID   string    `json:"groupId" db:"group_id"`
	UserID    *string   `json:"userId,omitempty" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	JoinedAt  time.Time `json:"joinedAt" db:"joined_at"`
	IsCreator bool      `json:"isCreator" db:"is_creator"`
}

// BucketListItem represents an item in a bucket list
type BucketListItem struct {
	ID          string     `json:"id" db:"id"`
	GroupID     string     `json:"groupId" db:"group_id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	Completed   bool       `json:"completed" db:"completed"`
	CompletedBy *string    `json:"completedBy,omitempty" db:"completed_by"`
	CompletedAt *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	CreatedBy   string     `json:"createdBy" db:"created_by"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
}

// GroupWithDetails includes group with members and items
type GroupWithDetails struct {
	Group   `json:",inline"`
	Members []Member         `json:"members"`
	Items   []BucketListItem `json:"items"`
}

// GroupSummary provides summary information for dashboard
type GroupSummary struct {
	Group           `json:",inline"`
	MemberCount     int     `json:"memberCount"`
	ItemCount       int     `json:"itemCount"`
	CompletedCount  int     `json:"completedCount"`
	ProgressPercent float64 `json:"progressPercent"`
}

// SupabaseUser represents a user from Supabase
type SupabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Request/Response types for API
type CreateGroupRequest struct {
	Name     string     `json:"name" binding:"required"`
	Deadline *time.Time `json:"deadline,omitempty"`
}

type JoinGroupRequest struct {
	MemberName string  `json:"memberName" binding:"required"`
	UserID     *string `json:"userId,omitempty"`
}

type CreateItemRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description,omitempty"`
	MemberID    string  `json:"memberId" binding:"required"`
}

type ToggleCompletionRequest struct {
	Completed bool   `json:"completed"`
	MemberID  string `json:"memberId" binding:"required"`
}

// WebSocket event types
type WebSocketEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type JoinGroupEvent struct {
	GroupID  string `json:"groupId"`
	MemberID string `json:"memberId"`
}

type AddItemEvent struct {
	GroupID string            `json:"groupId"`
	Item    CreateItemRequest `json:"item"`
}

type ToggleCompletionEvent struct {
	GroupID   string `json:"groupId"`
	ItemID    string `json:"itemId"`
	Completed bool   `json:"completed"`
}

// Validation error types
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationResult struct {
	IsValid bool              `json:"isValid"`
	Errors  []ValidationError `json:"errors"`
}

// Constants for validation
const (
	MaxGroupNameLength       = 100
	MinGroupNameLength       = 2
	MaxMemberNameLength      = 50
	MinMemberNameLength      = 1
	MaxItemTitleLength       = 200
	MinItemTitleLength       = 1
	MaxItemDescriptionLength = 1000
)

// Validation functions
func ValidateGroupName(name string) ValidationResult {
	var errors []ValidationError
	
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Group name is required",
		})
	} else if len(name) < MinGroupNameLength {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("Group name must be at least %d characters long", MinGroupNameLength),
		})
	} else if len(name) > MaxGroupNameLength {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("Group name must be less than %d characters", MaxGroupNameLength),
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

func ValidateMemberName(name string) ValidationResult {
	var errors []ValidationError
	
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Member name is required",
		})
	} else if len(name) < MinMemberNameLength {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Member name cannot be empty",
		})
	} else if len(name) > MaxMemberNameLength {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("Member name must be less than %d characters", MaxMemberNameLength),
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

func ValidateItemTitle(title string) ValidationResult {
	var errors []ValidationError
	
	title = strings.TrimSpace(title)
	if len(title) == 0 {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Item title is required",
		})
	} else if len(title) < MinItemTitleLength {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Item title cannot be empty",
		})
	} else if len(title) > MaxItemTitleLength {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: fmt.Sprintf("Item title must be less than %d characters", MaxItemTitleLength),
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

func ValidateItemDescription(description *string) ValidationResult {
	var errors []ValidationError
	
	if description != nil && len(*description) > MaxItemDescriptionLength {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: fmt.Sprintf("Item description must be less than %d characters", MaxItemDescriptionLength),
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

func ValidateDeadline(deadline *time.Time) ValidationResult {
	var errors []ValidationError
	
	if deadline != nil && deadline.Before(time.Now()) {
		errors = append(errors, ValidationError{
			Field:   "deadline",
			Message: "Deadline must be in the future",
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

func ValidateUUID(id string) ValidationResult {
	var errors []ValidationError
	
	// UUID v4 regex pattern
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(uuidPattern, id)
	
	if !matched {
		errors = append(errors, ValidationError{
			Field:   "id",
			Message: "Invalid UUID format",
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// Validation methods for structs
func (req *CreateGroupRequest) Validate() ValidationResult {
	nameValidation := ValidateGroupName(req.Name)
	deadlineValidation := ValidateDeadline(req.Deadline)
	
	var allErrors []ValidationError
	allErrors = append(allErrors, nameValidation.Errors...)
	allErrors = append(allErrors, deadlineValidation.Errors...)
	
	return ValidationResult{
		IsValid: len(allErrors) == 0,
		Errors:  allErrors,
	}
}

func (req *JoinGroupRequest) Validate() ValidationResult {
	return ValidateMemberName(req.MemberName)
}

func (req *CreateItemRequest) Validate() ValidationResult {
	titleValidation := ValidateItemTitle(req.Title)
	descriptionValidation := ValidateItemDescription(req.Description)
	
	var errors []ValidationError
	errors = append(errors, titleValidation.Errors...)
	errors = append(errors, descriptionValidation.Errors...)
	
	if strings.TrimSpace(req.MemberID) == "" {
		errors = append(errors, ValidationError{
			Field:   "memberId",
			Message: "Member ID is required",
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

func (req *ToggleCompletionRequest) Validate() ValidationResult {
	var errors []ValidationError
	
	if strings.TrimSpace(req.MemberID) == "" {
		errors = append(errors, ValidationError{
			Field:   "memberId",
			Message: "Member ID is required",
		})
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// Sanitization functions
func SanitizeString(input string) string {
	// Trim whitespace and normalize multiple spaces to single space
	trimmed := strings.TrimSpace(input)
	spaceRegex := regexp.MustCompile(`\s+`)
	return spaceRegex.ReplaceAllString(trimmed, " ")
}

func (g *Group) Sanitize() {
	g.Name = SanitizeString(g.Name)
}

func (m *Member) Sanitize() {
	m.Name = SanitizeString(m.Name)
}

func (b *BucketListItem) Sanitize() {
	b.Title = SanitizeString(b.Title)
	if b.Description != nil {
		sanitized := SanitizeString(*b.Description)
		b.Description = &sanitized
	}
}

func (req *CreateGroupRequest) Sanitize() {
	req.Name = SanitizeString(req.Name)
}

func (req *JoinGroupRequest) Sanitize() {
	req.MemberName = SanitizeString(req.MemberName)
}

func (req *CreateItemRequest) Sanitize() {
	req.Title = SanitizeString(req.Title)
	if req.Description != nil {
		sanitized := SanitizeString(*req.Description)
		req.Description = &sanitized
	}
}

// Helper functions for data integrity
func (g *Group) IsValid() error {
	validation := ValidateGroupName(g.Name)
	if !validation.IsValid {
		return errors.New(validation.Errors[0].Message)
	}
	
	deadlineValidation := ValidateDeadline(g.Deadline)
	if !deadlineValidation.IsValid {
		return errors.New(deadlineValidation.Errors[0].Message)
	}
	
	if strings.TrimSpace(g.CreatedBy) == "" {
		return errors.New("created by user ID is required")
	}
	
	return nil
}

func (m *Member) IsValid() error {
	validation := ValidateMemberName(m.Name)
	if !validation.IsValid {
		return errors.New(validation.Errors[0].Message)
	}
	
	if strings.TrimSpace(m.GroupID) == "" {
		return errors.New("group ID is required")
	}
	
	return nil
}

func (b *BucketListItem) IsValid() error {
	titleValidation := ValidateItemTitle(b.Title)
	if !titleValidation.IsValid {
		return errors.New(titleValidation.Errors[0].Message)
	}
	
	descriptionValidation := ValidateItemDescription(b.Description)
	if !descriptionValidation.IsValid {
		return errors.New(descriptionValidation.Errors[0].Message)
	}
	
	if strings.TrimSpace(b.GroupID) == "" {
		return errors.New("group ID is required")
	}
	
	if strings.TrimSpace(b.CreatedBy) == "" {
		return errors.New("created by member ID is required")
	}
	
	return nil
}