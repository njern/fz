package api

import (
	"strings"
	"time"
)

// Identity represents the response from /my/identity.
type Identity struct {
	Accounts []Account `json:"accounts"`
}

// Account represents a Fizzy account/organization.
type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
}

// SlugTrimmed returns the account slug without the leading slash.
func (a Account) SlugTrimmed() string {
	return strings.TrimPrefix(a.Slug, "/")
}

// User represents a user within an account.
type User struct {
	CreatedAt    time.Time `json:"created_at"`
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	EmailAddress string    `json:"email_address"`
	URL          string    `json:"url"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Active       bool      `json:"active"`
}

// Board represents a Fizzy board.
type Board struct {
	Creator   User      `json:"creator"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	PublicURL string    `json:"public_url,omitempty"`
	AllAccess bool      `json:"all_access"`
}

// Column represents a workflow column on a board.
type Column struct {
	CreatedAt time.Time   `json:"created_at"`
	Color     ColumnColor `json:"color"`
	ID        string      `json:"id"`
	Name      string      `json:"name"`
}

// ColumnColor represents a column's color.
type ColumnColor struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Card represents a work item on a board.
type Card struct {
	LastActiveAt    time.Time `json:"last_active_at"`
	CreatedAt       time.Time `json:"created_at"`
	ImageURL        *string   `json:"image_url"`
	Column          *Column   `json:"column,omitempty"`
	Board           Board     `json:"board"`
	Creator         User      `json:"creator"`
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	Description     string    `json:"description"`
	DescriptionHTML string    `json:"description_html"`
	URL             string    `json:"url"`
	CommentsURL     string    `json:"comments_url"`
	ReactionsURL    string    `json:"reactions_url,omitempty"`
	Tags            []string  `json:"tags"`
	Steps           []Step    `json:"steps,omitempty"`
	Number          int       `json:"number"`
	HasAttachments  bool      `json:"has_attachments"`
	Closed          bool      `json:"closed"`
	Postponed       bool      `json:"postponed"`
	Golden          bool      `json:"golden"`
}

// Comment represents a comment on a card.
type Comment struct {
	Creator   User        `json:"creator"`
	Body      CommentBody `json:"body"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	ID        string      `json:"id"`
	URL       string      `json:"url"`
}

// CommentBody holds the plain text and HTML versions of a comment.
type CommentBody struct {
	PlainText string `json:"plain_text"`
	HTML      string `json:"html"`
}

// Step represents a to-do item on a card.
type Step struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Completed bool   `json:"completed"`
}

// Tag represents a label for organizing cards.
type Tag struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
}

// Notification represents a user notification.
type Notification struct {
	Creator   User       `json:"creator"`
	Card      CardRef    `json:"card"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at"`
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	URL       string     `json:"url"`
	Read      bool       `json:"read"`
}

// CardRef is a minimal card reference in notifications.
type CardRef struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

// Reaction represents a boost/reaction on a card or comment.
type Reaction struct {
	Reacter User   `json:"reacter"`
	ID      string `json:"id"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// Webhook represents a webhook on a board.
type Webhook struct {
	Board             Board     `json:"board"`
	CreatedAt         time.Time `json:"created_at"`
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	PayloadURL        string    `json:"payload_url"`
	SigningSecret     string    `json:"signing_secret"`
	URL               string    `json:"url"`
	SubscribedActions []string  `json:"subscribed_actions"`
	Active            bool      `json:"active"`
}
