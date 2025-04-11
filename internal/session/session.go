package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/saurabh0719/kiwi/internal/input"
	"github.com/saurabh0719/kiwi/internal/llm"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
)

// Maximum number of sessions to keep
const MaxSessions = 10

type Message struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Summary   string    `json:"summary"` // Simple 1-line summary
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
	// Check if we need to prune older sessions
	sessions, err := m.GetAllSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to get existing sessions: %w", err)
	}

	// If we already have MaxSessions, delete the oldest one
	if len(sessions) >= MaxSessions {
		// Sessions are sorted newest first, so the oldest is last
		oldestSession := sessions[len(sessions)-1]
		if err := m.DeleteSession(oldestSession.ID); err != nil {
			return nil, fmt.Errorf("failed to delete oldest session: %w", err)
		}
	}

	session := &Session{
		ID:        id,
		CreatedAt: time.Now(),
		Summary:   "New session", // Will be updated after first message
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

	wasPreviouslyEmpty := len(session.Messages) == 0

	session.Messages = append(session.Messages, Message{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})

	if err := m.saveSession(session); err != nil {
		return err
	}

	// If this was the first user message, update the summary
	if wasPreviouslyEmpty && role == "user" {
		return m.UpdateSessionSummary(sessionID)
	}

	return nil
}

// UpdateSessionSummary generates a simple one-line summary from the first user message
// or sets the provided summaryText if not empty
func (m *Manager) UpdateSessionSummary(sessionID string, summaryText ...string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	// If a summary text is provided, use that
	if len(summaryText) > 0 && summaryText[0] != "" {
		session.Summary = summaryText[0]
		return m.saveSession(session)
	}

	// Otherwise, generate from the first user message
	// Look for the first substantive user message
	var firstUserMsg string
	for _, msg := range session.Messages {
		if msg.Role == "user" && len(msg.Content) > 5 {
			firstUserMsg = msg.Content
			break
		}
	}

	// Truncate and clean up for summary
	if len(firstUserMsg) > 0 {
		// Take first 50 chars or first line, whichever is shorter
		summary := firstUserMsg
		if idx := strings.IndexAny(summary, "\n\r"); idx > 0 {
			summary = summary[:idx]
		}
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}
		session.Summary = summary
		return m.saveSession(session)
	}

	return nil
}

// GetAllSessions returns all sessions sorted by creation time (newest first)
func (m *Manager) GetAllSessions() ([]*Session, error) {
	ids, err := m.ListSessions()
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, 0, len(ids))
	for _, id := range ids {
		s, err := m.GetSession(id)
		if err != nil {
			continue // Skip sessions that can't be loaded
		}
		sessions = append(sessions, s)
	}

	// Sort by creation time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})

	return sessions, nil
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

	// Save the session file
	if err := os.WriteFile(m.sessionPath(session.ID), data, 0644); err != nil {
		return err
	}

	// After saving, check if we need to enforce the MaxSessions limit
	// This ensures we maintain at most MaxSessions sessions at all times
	sessions, err := m.GetAllSessions()
	if err != nil {
		// Just log this error but don't fail the save operation
		fmt.Printf("Warning: could not check session count: %v\n", err)
		return nil
	}

	// If we have more than MaxSessions, delete the oldest sessions
	if len(sessions) > MaxSessions {
		// Sessions are already sorted, newest first
		for i := MaxSessions; i < len(sessions); i++ {
			// Delete oldest sessions (beyond our limit)
			if err := m.DeleteSession(sessions[i].ID); err != nil {
				// Log the error but continue
				fmt.Printf("Warning: could not delete old session %s: %v\n", sessions[i].ID, err)
			}
		}
	}

	return nil
}

func (m *Manager) sessionPath(id string) string {
	return filepath.Join(m.baseDir, id+".json")
}

// RunInteractiveSession handles an interactive chat session with a given session ID
// It takes care of the entire lifecycle of a session including setup, message processing,
// summary updates, and cleanup when the session ends
func (m *Manager) RunInteractiveSession(sessionID string, cfg *config.Config, isNewSession bool) error {
	sess, err := m.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Display a user-friendly ID
	displayID := sessionID
	if strings.HasPrefix(sessionID, "custom_") {
		// Extract the name part for display
		parts := strings.Split(sessionID, "_")
		if len(parts) > 1 {
			displayID = parts[1]
		}
	} else if strings.HasPrefix(sessionID, "session_") {
		// Just show the numeric part for auto-generated sessions
		displayID = strings.TrimPrefix(sessionID, "session_")
	}

	if isNewSession {
		util.InfoColor.Printf("Created new session: %s\n", displayID)
	} else {
		util.InfoColor.Printf("Continuing session: %s\n", displayID)
	}

	toolRegistry := tools.NewRegistry()
	tools.RegisterStandardTools(toolRegistry)

	adapter, err := llm.NewAdapter(cfg.LLM.Provider, cfg.LLM.Model, cfg.LLM.APIKey, toolRegistry)
	if err != nil {
		return fmt.Errorf("failed to create LLM adapter: %w", err)
	}

	// Session info
	if isNewSession {
		fmt.Println("Assistant session started. Type 'exit' to end the session. Use Shift+Enter for new lines, Enter to submit")
	} else {
		fmt.Println("Assistant session continued. Type 'exit' to end the session. Use Shift+Enter for new lines, Enter to submit")
		fmt.Printf("Previous conversation has %d messages.\n", len(sess.Messages))
	}

	util.InfoColor.Printf("Using %s model: %s\n", adapter.GetProvider(), adapter.GetModel())
	util.PrintChatDivider()

	// Check if input is being piped in (non-interactive mode)
	isPiped, singleInput := input.IsInputPiped()

	// If we're in piped mode, process the single input and exit
	if isPiped {
		if singleInput == "" {
			return fmt.Errorf("no input provided in non-interactive mode")
		}

		fmt.Println()
		util.UserColor.Print("You: ")
		fmt.Println(singleInput)

		// Process a single message and then exit
		if err := ProcessChatMessage(m, *sess, *cfg, adapter, singleInput); err != nil {
			return err
		}

		// Update the summary after adding a message
		return UpdateSessionSummaryLLM(m, sessionID, cfg)
	}

	// Interactive chat loop
	for {
		fmt.Println()
		util.UserColor.Print("You: ")

		userInput, err := input.ReadMultiLine("")
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Skip empty inputs and request a non-empty message
		if strings.TrimSpace(userInput) == "" {
			// Silently continue the loop to re-prompt the user instead of showing an error
			continue
		}

		if strings.ToLower(strings.TrimSpace(userInput)) == "exit" {
			fmt.Println("Ending session...")
			// Update summary before exiting
			UpdateSessionSummaryLLM(m, sessionID, cfg)
			return nil
		}

		if err := ProcessChatMessage(m, *sess, *cfg, adapter, userInput); err != nil {
			return err
		}

		// Update the session after processing
		sess, err = m.GetSession(sess.ID)
		if err != nil {
			return fmt.Errorf("failed to get updated session: %w", err)
		}

		// Update the summary after each message exchange
		if err := UpdateSessionSummaryLLM(m, sessionID, cfg); err != nil {
			// Just log the error but continue the session
			util.InfoColor.Printf("Note: Could not update session summary: %v\n", err)
		}
	}
}
