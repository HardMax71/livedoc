package document

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HardMax71/syncwrite/backend/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrDocumentNotFound  = errors.New("document not found")
	ErrVersionMismatch   = errors.New("version mismatch")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrInvalidPermission = errors.New("invalid permission level")
)

type Service struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:     db,
		logger: utils.Logger(),
	}
}

func (s *Service) CreateDocument(ctx context.Context, params CreateDocumentParams) (*Document, error) {
	var doc Document
	err := s.db.QueryRow(ctx, `
        INSERT INTO documents (title, content, owner_id, version)
        VALUES ($1, $2, $3, $4)
        RETURNING id, title, content, owner_id, version, created_at, updated_at
    `, params.Title, params.Content, params.OwnerID, "1").Scan(
		&doc.ID, &doc.Title, &doc.Content, &doc.OwnerID,
		&doc.Version, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating document: %w", err)
	}

	// Create owner permission
	_, err = s.db.Exec(ctx, `
        INSERT INTO document_permissions (document_id, user_id, permission_level)
        VALUES ($1, $2, $3)
    `, doc.ID, params.OwnerID, PermissionLevelOwner)

	if err != nil {
		return nil, fmt.Errorf("error creating document permission: %w", err)
	}

	return &doc, nil
}

func (s *Service) GetDocument(ctx context.Context, documentID, userID string) (*Document, error) {
	var doc Document
	err := s.db.QueryRow(ctx, `
        SELECT d.id, d.title, d.content, d.owner_id, d.version, d.created_at, d.updated_at
        FROM documents d
        JOIN document_permissions p ON d.id = p.document_id
        WHERE d.id = $1 AND p.user_id = $2
    `, documentID, userID).Scan(
		&doc.ID, &doc.Title, &doc.Content, &doc.OwnerID,
		&doc.Version, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err != nil {
		return nil, ErrDocumentNotFound
	}

	return &doc, nil
}

func (s *Service) UpdateDocument(ctx context.Context, params UpdateDocumentParams) (*Document, error) {
	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check version
	var currentVersion string
	err = tx.QueryRow(ctx, `
        SELECT version FROM documents WHERE id = $1
    `, params.DocumentID).Scan(&currentVersion)

	if err != nil {
		return nil, ErrDocumentNotFound
	}

	if currentVersion != params.Version {
		return nil, ErrVersionMismatch
	}

	// Check permission
	var permissionLevel string
	err = tx.QueryRow(ctx, `
        SELECT permission_level FROM document_permissions
        WHERE document_id = $1 AND user_id = $2
    `, params.DocumentID, params.EditorID).Scan(&permissionLevel)

	if err != nil {
		return nil, ErrPermissionDenied
	}

	if permissionLevel != PermissionLevelEditor && permissionLevel != PermissionLevelOwner {
		return nil, ErrPermissionDenied
	}

	// Update document
	newVersion := fmt.Sprintf("%d", time.Now().UnixNano())
	var doc Document
	err = tx.QueryRow(ctx, `
        UPDATE documents
        SET title = $1, content = $2, version = $3, updated_at = NOW()
        WHERE id = $4
        RETURNING id, title, content, owner_id, version, created_at, updated_at
    `, params.Title, params.Content, newVersion, params.DocumentID).Scan(
		&doc.ID, &doc.Title, &doc.Content, &doc.OwnerID,
		&doc.Version, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error updating document: %w", err)
	}

	// Create version history
	_, err = tx.Exec(ctx, `
        INSERT INTO document_versions (document_id, content, editor_id, version)
        VALUES ($1, $2, $3, $4)
    `, doc.ID, params.Content, params.EditorID, newVersion)

	if err != nil {
		return nil, fmt.Errorf("error creating version history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &doc, nil
}

func (s *Service) DeleteDocument(ctx context.Context, documentID, userID string) error {
	// Check ownership
	var permissionLevel string
	err := s.db.QueryRow(ctx, `
        SELECT permission_level FROM document_permissions
        WHERE document_id = $1 AND user_id = $2
    `, documentID, userID).Scan(&permissionLevel)

	if err != nil {
		return ErrDocumentNotFound
	}

	if permissionLevel != PermissionLevelOwner {
		return ErrPermissionDenied
	}

	// Delete document and related data
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete permissions
	_, err = tx.Exec(ctx, `
        DELETE FROM document_permissions WHERE document_id = $1
    `, documentID)

	if err != nil {
		return fmt.Errorf("error deleting permissions: %w", err)
	}

	// Delete versions
	_, err = tx.Exec(ctx, `
        DELETE FROM document_versions WHERE document_id = $1
    `, documentID)

	if err != nil {
		return fmt.Errorf("error deleting versions: %w", err)
	}

	// Delete document
	_, err = tx.Exec(ctx, `
        DELETE FROM documents WHERE id = $1
    `, documentID)

	if err != nil {
		return fmt.Errorf("error deleting document: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *Service) ListDocuments(ctx context.Context, userID string, page, pageSize int32) ([]*Document, int32, error) {
	// Get total count
	var total int32
	err := s.db.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM documents d
        JOIN document_permissions p ON d.id = p.document_id
        WHERE p.user_id = $1
    `, userID).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("error counting documents: %w", err)
	}

	// Get documents
	rows, err := s.db.Query(ctx, `
        SELECT d.id, d.title, d.content, d.owner_id, d.version, d.created_at, d.updated_at
        FROM documents d
        JOIN document_permissions p ON d.id = p.document_id
        WHERE p.user_id = $1
        ORDER BY d.updated_at DESC
        LIMIT $2 OFFSET $3
    `, userID, pageSize, (page-1)*pageSize)

	if err != nil {
		return nil, 0, fmt.Errorf("error querying documents: %w", err)
	}
	defer rows.Close()

	var documents []*Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Content, &doc.OwnerID,
			&doc.Version, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning document: %w", err)
		}
		documents = append(documents, &doc)
	}

	return documents, total, nil
}

func (s *Service) ShareDocument(ctx context.Context, params ShareDocumentParams) (*Permission, error) {
	// Validate permission level
	switch params.Level {
	case PermissionLevelViewer, PermissionLevelEditor:
		// Valid levels
	default:
		return nil, ErrInvalidPermission
	}

	var permission Permission
	err := s.db.QueryRow(ctx, `
        INSERT INTO document_permissions (document_id, user_id, permission_level)
        VALUES ($1, $2, $3)
        RETURNING id, document_id, user_id, permission_level, created_at, updated_at
    `, params.DocumentID, params.UserID, params.Level).Scan(
		&permission.ID, &permission.DocumentID, &permission.UserID,
		&permission.Level, &permission.CreatedAt, &permission.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error sharing document: %w", err)
	}

	return &permission, nil
}

func (s *Service) GetDocumentHistory(ctx context.Context, documentID string, page, pageSize int32) ([]*DocumentVersion, int32, error) {
	// Get total count
	var total int32
	err := s.db.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM document_versions
        WHERE document_id = $1
    `, documentID).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("error counting versions: %w", err)
	}

	// Get versions
	rows, err := s.db.Query(ctx, `
        SELECT id, document_id, content, editor_id, version, created_at
        FROM document_versions
        WHERE document_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `, documentID, pageSize, (page-1)*pageSize)

	if err != nil {
		return nil, 0, fmt.Errorf("error querying versions: %w", err)
	}
	defer rows.Close()

	var versions []*DocumentVersion
	for rows.Next() {
		var version DocumentVersion
		err := rows.Scan(
			&version.ID, &version.DocumentID, &version.Content,
			&version.EditorID, &version.Version, &version.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning version: %w", err)
		}
		versions = append(versions, &version)
	}

	return versions, total, nil
}

func (s *Service) RestoreVersion(ctx context.Context, documentID, versionID, userID string) (*Document, error) {
	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check permission
	var permissionLevel string
	err = tx.QueryRow(ctx, `
        SELECT permission_level FROM document_permissions
        WHERE document_id = $1 AND user_id = $2
    `, documentID, userID).Scan(&permissionLevel)

	if err != nil {
		return nil, ErrPermissionDenied
	}

	if permissionLevel != PermissionLevelEditor && permissionLevel != PermissionLevelOwner {
		return nil, ErrPermissionDenied
	}

	// Get version content
	var versionContent string
	err = tx.QueryRow(ctx, `
        SELECT content FROM document_versions
        WHERE id = $1 AND document_id = $2
    `, versionID, documentID).Scan(&versionContent)

	if err != nil {
		return nil, ErrDocumentNotFound
	}

	// Update document with version content
	newVersion := fmt.Sprintf("%d", time.Now().UnixNano())
	var doc Document
	err = tx.QueryRow(ctx, `
        UPDATE documents
        SET content = $1, version = $2, updated_at = NOW()
        WHERE id = $3
        RETURNING id, title, content, owner_id, version, created_at, updated_at
    `, versionContent, newVersion, documentID).Scan(
		&doc.ID, &doc.Title, &doc.Content, &doc.OwnerID,
		&doc.Version, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error restoring version: %w", err)
	}

	// Create version history entry for restoration
	_, err = tx.Exec(ctx, `
        INSERT INTO document_versions (document_id, content, editor_id, version)
        VALUES ($1, $2, $3, $4)
    `, doc.ID, versionContent, userID, newVersion)

	if err != nil {
		return nil, fmt.Errorf("error creating version history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &doc, nil
}
