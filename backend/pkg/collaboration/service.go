package collaboration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/HardMax71/syncwrite/backend/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrUserNotFound    = errors.New("user not found")
)

type Service struct {
	db            *pgxpool.Pool
	logger        *zap.Logger
	mqtt          *MQTTClient
	sessionMgr    *SessionManager
	changeStreams map[string]map[string]chan *DocumentChange
	streamsMutex  sync.RWMutex
}

func NewService(db *pgxpool.Pool, mqtt *MQTTClient) *Service {
	return &Service{
		db:            db,
		logger:        utils.Logger(),
		mqtt:          mqtt,
		sessionMgr:    NewSessionManager(),
		changeStreams: make(map[string]map[string]chan *DocumentChange),
	}
}

func (s *Service) JoinSession(documentID string, user *ActiveUser) ([]*ActiveUser, error) {
	session := s.sessionMgr.GetOrCreateSession(documentID)
	session.AddUser(user)

	// Subscribe to document changes
	err := s.mqtt.Subscribe(GetDocumentTopic(documentID), func(payload []byte) {
		var change DocumentChange
		if err := json.Unmarshal(payload, &change); err != nil {
			s.logger.Error("Error unmarshaling document change", zap.Error(err))
			return
		}
		s.broadcastChange(documentID, &change)
	})

	if err != nil {
		return nil, fmt.Errorf("error subscribing to document changes: %w", err)
	}

	// Notify other users about the new user
	presenceUpdate := map[string]interface{}{
		"type": "join",
		"user": user,
	}
	if err := s.mqtt.Publish(GetPresenceTopic(documentID), presenceUpdate); err != nil {
		s.logger.Error("Error publishing presence update", zap.Error(err))
	}

	return session.GetActiveUsers(), nil
}

func (s *Service) LeaveSession(documentID, userID string) error {
	session := s.sessionMgr.GetOrCreateSession(documentID)
	session.RemoveUser(userID)

	// If no users left, remove the session and unsubscribe from MQTT topics
	if len(session.GetActiveUsers()) == 0 {
		s.sessionMgr.RemoveSession(documentID)
		if err := s.mqtt.Unsubscribe(GetDocumentTopic(documentID)); err != nil {
			s.logger.Error("Error unsubscribing from document topic", zap.Error(err))
		}
	}

	// Notify other users about the user leaving
	presenceUpdate := map[string]interface{}{
		"type":    "leave",
		"user_id": userID,
	}
	if err := s.mqtt.Publish(GetPresenceTopic(documentID), presenceUpdate); err != nil {
		s.logger.Error("Error publishing presence update", zap.Error(err))
	}

	return nil
}

func (s *Service) GetActiveUsers(documentID string) ([]*ActiveUser, error) {
	session := s.sessionMgr.GetOrCreateSession(documentID)
	return session.GetActiveUsers(), nil
}

func (s *Service) UpdateUserActivity(documentID, userID, cursorPosition string) error {
	session := s.sessionMgr.GetOrCreateSession(documentID)
	session.UpdateUserActivity(userID, cursorPosition)

	// Publish cursor position update
	update := map[string]interface{}{
		"type":            "cursor",
		"user_id":         userID,
		"cursor_position": cursorPosition,
	}
	return s.mqtt.Publish(GetPresenceTopic(documentID), update)
}

func (s *Service) StreamChanges(documentID, userID string) (<-chan *DocumentChange, func(), error) {
	s.streamsMutex.Lock()
	defer s.streamsMutex.Unlock()

	if _, exists := s.changeStreams[documentID]; !exists {
		s.changeStreams[documentID] = make(map[string]chan *DocumentChange)
	}

	changeChan := make(chan *DocumentChange, 100)
	s.changeStreams[documentID][userID] = changeChan

	cleanup := func() {
		s.streamsMutex.Lock()
		defer s.streamsMutex.Unlock()

		if streams, exists := s.changeStreams[documentID]; exists {
			if ch, ok := streams[userID]; ok {
				close(ch)
				delete(streams, userID)
			}
			if len(streams) == 0 {
				delete(s.changeStreams, documentID)
			}
		}
	}

	return changeChan, cleanup, nil
}

func (s *Service) SyncDocument(ctx context.Context, documentID string, operations []Operation, baseVersion string) (string, []*DocumentChange, error) {
	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check current version
	var currentVersion string
	err = tx.QueryRow(ctx, `
        SELECT version FROM documents WHERE id = $1
    `, documentID).Scan(&currentVersion)

	if err != nil {
		return "", nil, fmt.Errorf("error getting current version: %w", err)
	}

	if currentVersion != baseVersion {
		// Get concurrent changes
		rows, err := tx.Query(ctx, `
            SELECT content FROM document_versions
            WHERE document_id = $1 AND version > $2
            ORDER BY created_at ASC
        `, documentID, baseVersion)
		if err != nil {
			return "", nil, fmt.Errorf("error getting concurrent changes: %w", err)
		}
		defer rows.Close()

		var concurrentChanges []*DocumentChange
		for rows.Next() {
			var content string
			if err := rows.Scan(&content); err != nil {
				return "", nil, fmt.Errorf("error scanning concurrent change: %w", err)
			}

			var change DocumentChange
			if err := json.Unmarshal([]byte(content), &change); err != nil {
				return "", nil, fmt.Errorf("error unmarshaling concurrent change: %w", err)
			}
			concurrentChanges = append(concurrentChanges, &change)
		}

		return currentVersion, concurrentChanges, nil
	}

	// Apply changes and create new version
	newVersion := fmt.Sprintf("%d", time.Now().UnixNano())
	change := &DocumentChange{
		DocumentID: documentID,
		Version:    newVersion,
		Operations: operations,
		Timestamp:  time.Now(),
	}

	changeJSON, err := json.Marshal(change)
	if err != nil {
		return "", nil, fmt.Errorf("error marshaling change: %w", err)
	}

	// Update document version
	_, err = tx.Exec(ctx, `
        UPDATE documents SET version = $1 WHERE id = $2
    `, newVersion, documentID)
	if err != nil {
		return "", nil, fmt.Errorf("error updating document version: %w", err)
	}

	// Store change in version history
	_, err = tx.Exec(ctx, `
        INSERT INTO document_versions (document_id, content, version)
        VALUES ($1, $2, $3)
    `, documentID, string(changeJSON), newVersion)
	if err != nil {
		return "", nil, fmt.Errorf("error storing version history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", nil, fmt.Errorf("error committing transaction: %w", err)
	}

	// Broadcast change to all connected clients
	if err := s.mqtt.Publish(GetDocumentTopic(documentID), change); err != nil {
		s.logger.Error("Error broadcasting document change", zap.Error(err))
	}

	return newVersion, nil, nil
}

func (s *Service) broadcastChange(documentID string, change *DocumentChange) {
	s.streamsMutex.RLock()
	defer s.streamsMutex.RUnlock()

	if streams, exists := s.changeStreams[documentID]; exists {
		for _, ch := range streams {
			select {
			case ch <- change:
			default:
				s.logger.Warn("Change channel full, dropping message",
					zap.String("document_id", documentID))
			}
		}
	}
}
