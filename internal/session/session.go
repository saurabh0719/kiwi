package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Message struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Messages  []Message `json:"messages"`
}

type Manager struct {
	baseDir string
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".kiwi", "sessions")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &Manager{baseDir: baseDir}, nil
}

func (m *Manager) CreateSession(id string) (*Session, error) {
	session := &Session{
		ID:        id,
		CreatedAt: time.Now(),
		Messages:  make([]Message, 0),
	}

	if err := m.saveSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (m *Manager) GetSession(id string) (*Session, error) {
	data, err := os.ReadFile(m.sessionPath(id))
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (m *Manager) AddMessage(sessionID string, role, content string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.Messages = append(session.Messages, Message{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})

	return m.saveSession(session)
}

func (m *Manager) ListSessions() ([]string, error) {
	files, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			sessions = append(sessions, file.Name()[:len(file.Name())-5])
		}
	}

	return sessions, nil
}

func (m *Manager) DeleteSession(id string) error {
	return os.Remove(m.sessionPath(id))
}

func (m *Manager) ClearSessions() error {
	files, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read sessions directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			if err := os.Remove(filepath.Join(m.baseDir, file.Name())); err != nil {
				return fmt.Errorf("failed to delete session %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

func (m *Manager) saveSession(session *Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return os.WriteFile(m.sessionPath(session.ID), data, 0644)
}

func (m *Manager) sessionPath(id string) string {
	return filepath.Join(m.baseDir, id+".json")
}
