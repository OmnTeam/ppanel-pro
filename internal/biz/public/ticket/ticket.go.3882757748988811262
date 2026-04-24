package ticket

import (
	"context"
)

// TicketStatus constants
const (
	StatusPending   = 1 // Pending - waiting for admin response
	StatusWaiting   = 2 // Waiting - waiting for user response
	StatusProcessed = 3 // Processed - ticket has been handled
	StatusClosed    = 4 // Closed - ticket is closed
)

// FollowType constants
const (
	FollowTypeText  = 1 // Text follow-up
	FollowTypeImage = 2 // Image follow-up
)

// CreateTicketParams contains parameters for creating a ticket
type CreateTicketParams struct {
	UserID      int64
	Title       string
	Description string
}

// GetTicketListParams contains parameters for getting ticket list
type GetTicketListParams struct {
	UserID int64
	Page   int64
	Size   int64
	Status *int32
	Search *string
}

// GetTicketListResult contains the result of getting ticket list
type GetTicketListResult struct {
	Total int32
	List  []*TicketInfo
}

// GetTicketDetailsParams contains parameters for getting ticket details
type GetTicketDetailsParams struct {
	UserID int64
	ID     int64
}

// UpdateTicketStatusParams contains parameters for updating ticket status
type UpdateTicketStatusParams struct {
	UserID int64
	ID     int64
	Status int32
}

// CreateTicketFollowParams contains parameters for creating ticket follow-up
type CreateTicketFollowParams struct {
	UserID   int64
	TicketID int64
	From     string
	Type     int32
	Content  string
}

// TicketInfo represents ticket information with follow-ups
type TicketInfo struct {
	ID          int64
	Title       string
	Description string
	UserID      int64
	Status      int32
	CreatedAt   int // Unix timestamp in milliseconds
	UpdatedAt   int // Unix timestamp in milliseconds
	Follows     []*TicketFollow
}

// TicketFollow represents a follow-up record
type TicketFollow struct {
	ID        int64
	TicketID  int64
	From      string
	Type      int32
	Content   string
	CreatedAt int // Unix timestamp in milliseconds
}

// TicketUseCase defines the interface for ticket business logic
type TicketUseCase interface {
	// CreateTicket creates a new ticket
	CreateTicket(ctx context.Context, params *CreateTicketParams) error

	// GetTicketList gets user's ticket list with pagination
	GetTicketList(ctx context.Context, params *GetTicketListParams) (*GetTicketListResult, error)

	// GetTicketDetails gets ticket details with follow-ups
	GetTicketDetails(ctx context.Context, params *GetTicketDetailsParams) (*TicketInfo, error)

	// UpdateTicketStatus updates ticket status
	UpdateTicketStatus(ctx context.Context, params *UpdateTicketStatusParams) error

	// CreateTicketFollow creates a follow-up and updates ticket status to Pending
	CreateTicketFollow(ctx context.Context, params *CreateTicketFollowParams) error
}

// TicketRepo defines the interface for ticket data access
type TicketRepo interface {
	// CreateTicket creates a new ticket
	CreateTicket(ctx context.Context, userID int, title, description string) error

	// GetTicketList gets user's ticket list
	GetTicketList(ctx context.Context, userID int, page, size int, status *int32, search *string) (int32, []*TicketInfo, error)

	// GetTicketByID gets ticket by ID
	GetTicketByID(ctx context.Context, ticketID int) (*TicketInfo, error)

	// UpdateTicketStatus updates ticket status
	UpdateTicketStatus(ctx context.Context, userID, ticketID int64, status int32) error

	// CreateTicketFollow creates a follow-up record
	CreateTicketFollow(ctx context.Context, ticketID int64, from string, followType int32, content string) error

	// GetTicketFollows gets all follow-ups for a ticket
	GetTicketFollows(ctx context.Context, ticketID int) ([]*TicketFollow, error)
}
