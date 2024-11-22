package collaboration

import (
	"sync"
	"time"
)

type ActiveUser struct {
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	CursorPosition string    `json:"cursor_position"`
	LastActive     time.Time `json:"last_active"`
}

type DocumentSession struct {
	DocumentID  string                 `json:"document_id"`
	ActiveUsers map[string]*ActiveUser `json:"active_users"`
	mutex       sync.RWMutex
}

type Operation struct {
	Type     OperationType `json:"type"`
	Position int32         `json:"position"`
	Content  string        `json:"content"`
	Length   int32         `json:"length"`
}

type OperationType int

const (
	OperationTypeInsert OperationType = iota
	OperationTypeDelete
	OperationTypeReplace
)

type DocumentChange struct {
	DocumentID string      `json:"document_id"`
	UserID     string      `json:"user_id"`
	Version    string      `json:"version"`
	Operations []Operation `json:"operations"`
	Timestamp  time.Time   `json:"timestamp"`
}

type SessionManager struct {
	sessions map[string]*DocumentSession
	mutex    sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*DocumentSession),
	}
}

func (sm *SessionManager) GetOrCreateSession(documentID string) *DocumentSession {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[documentID]
	if !exists {
		session = &DocumentSession{
			DocumentID:  documentID,
			ActiveUsers: make(map[string]*ActiveUser),
		}
		sm.sessions[documentID] = session
	}
	return session
}

func (sm *SessionManager) RemoveSession(documentID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.sessions, documentID)
}

func (s *DocumentSession) AddUser(user *ActiveUser) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ActiveUsers[user.UserID] = user
}

func (s *DocumentSession) RemoveUser(userID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.ActiveUsers, userID)
}

func (s *DocumentSession) GetActiveUsers() []*ActiveUser {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	users := make([]*ActiveUser, 0, len(s.ActiveUsers))
	for _, user := range s.ActiveUsers {
		users = append(users, user)
	}
	return users
}

func (s *DocumentSession) UpdateUserActivity(userID string, cursorPosition string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if user, exists := s.ActiveUsers[userID]; exists {
		user.CursorPosition = cursorPosition
		user.LastActive = time.Now()
	}
}
