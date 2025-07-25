package models

import "time"

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