package document

import (
	"time"
)

type Document struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	OwnerID   string    `json:"owner_id"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DocumentVersion struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Content    string    `json:"content"`
	EditorID   string    `json:"editor_id"`
	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
}

type Permission struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	DocumentID string    `json:"document_id"`
	Level      string    `json:"level"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateDocumentParams struct {
	Title   string
	Content string
	OwnerID string
}

type UpdateDocumentParams struct {
	DocumentID string
	Title      string
	Content    string
	Version    string
	EditorID   string
}

type ShareDocumentParams struct {
	DocumentID string
	UserID     string
	Level      string
}

const (
	PermissionLevelViewer = "VIEWER"
	PermissionLevelEditor = "EDITOR"
	PermissionLevelOwner  = "OWNER"
)
