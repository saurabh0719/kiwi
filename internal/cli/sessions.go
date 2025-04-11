package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/saurabh0719/kiwi/internal/session"
	"github.com/saurabh0719/kiwi/internal/util"
	"github.com/spf13/cobra"
)

var (
	// Session command flags
	listFlag     bool // -l flag for listing sessions
	newFlag      bool // -n flag for creating a new session
	continueFlag bool // -o flag for continuing a session
	deleteFlag   bool // -d flag for deleting a session
	// clearFlag is now defined in root.go
	sessionName string // session name for new sessions
	sessionID   string // session ID for actions like continue, delete
)

var sessionsCmd *cobra.Command

// Helper function to get full session ID from numeric input
func getFullSessionID(numericID string) string {
	// If it already has the session_ prefix, return as is
	if strings.HasPrefix(numericID, "session_") {
		return numericID
	}

	// Try to parse as a number to validate
	_, err := strconv.ParseInt(numericID, 10, 64)
	if err != nil {
		return numericID // If not a number, return as is (will likely fail later)
	}

	return "session_" + numericID
}

func initSessionsCmd() {
	sessionsCmd = &cobra.Command{
		Use:   "sessions",
		Short: "Manage assistant sessions",
		Long: `Manage assistant sessions using flags.

Available flags:
  -l, --list        List all sessions (default behavior if no flag specified)
  -n, --new         Create a new named session (provide name as argument)
  -o, --continue    Continue a session (provide ID as argument)
  -d, --delete      Delete a session (provide ID as argument)
  -C, --clear       Clear all sessions

Examples:
  kiwi -s            # List all sessions
  kiwi -s -l         # Also lists all sessions
  kiwi -s -n my_project    # Create a new named session
  kiwi -s -o 1234567       # Continue a session by ID
  kiwi -s -d 1234567       # Delete a session by ID
  kiwi -s -C                    # Clear all sessions`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check flags in priority order
			switch {
			case listFlag:
				return handleSessionsList(cmd, args)
			case newFlag:
				if len(args) > 0 {
					return handleSessionsNew(cmd, []string{args[0]})
				} else if sessionName != "" {
					return handleSessionsNew(cmd, []string{sessionName})
				}
				return fmt.Errorf("session name required with -n flag (either as argument or with --name)")
			case continueFlag:
				if len(args) > 0 {
					return handleSessionsContinue(cmd, []string{args[0]})
				} else if sessionID != "" {
					return handleSessionsContinue(cmd, []string{sessionID})
				}
				return fmt.Errorf("session ID required with -o flag (either as argument or with --id)")
			case deleteFlag:
				if len(args) > 0 {
					return handleSessionsDelete(cmd, []string{args[0]})
				} else if sessionID != "" {
					return handleSessionsDelete(cmd, []string{sessionID})
				}
				return fmt.Errorf("session ID required with -d flag (either as argument or with --id)")
			case clearFlag:
				return handleSessionsClear(cmd, args)
			default:
				// Default behavior is to list sessions
				return handleSessionsList(cmd, args)
			}
		},
	}

	// Add flags to sessionsCmd
	sessionsCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List available sessions")
	sessionsCmd.Flags().BoolVarP(&newFlag, "new", "n", false, "Create a new session")
	sessionsCmd.Flags().BoolVarP(&continueFlag, "continue", "o", false, "Continue a session")
	sessionsCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete a session")
	sessionsCmd.Flags().BoolVarP(&clearFlag, "clear", "C", false, "Clear all sessions")

	// Parameters for flags
	sessionsCmd.Flags().StringVar(&sessionName, "name", "", "Name for the new session (used with -n)")
	sessionsCmd.Flags().StringVar(&sessionID, "id", "", "Session ID for continue or delete operations (used with -o or -d)")
}

// Updates the session summary using the LLM
func updateSessionSummary(sessionMgr *session.Manager, sessionID string, cfg *config.Config) error {
	err := session.UpdateSessionSummaryLLM(sessionMgr, sessionID, cfg)
	if err != nil {
		util.WarningColor.Printf("Failed to update session summary: %v\n", err)
	}
	return err
}

func handleSessionsList(cmd *cobra.Command, args []string) error {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	sessions, err := sessionMgr.GetAllSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		util.InfoColor.Println("No sessions found.")
		return nil
	}

	// Load config for summary generation if needed
	cfg, err := config.Load(cmd)
	if err != nil {
		util.WarningColor.Printf("Failed to load config, summaries may not be updated: %v\n", err)
	}

	// Print a header with color
	util.HeaderColor.Println("\n📋 Available Sessions")
	fmt.Println()

	for i, s := range sessions {
		// Extract the numeric part from session_123456
		numericID := strings.TrimPrefix(s.ID, "session_")
		if strings.HasPrefix(s.ID, "custom_") {
			parts := strings.Split(s.ID, "_")
			if len(parts) > 2 {
				numericID = parts[1] + " (" + parts[2] + ")"
			}
		}

		// Display session with colors and formatting
		util.SessionIDColor.Printf("  %d. ", i+1)
		util.SessionIDColor.Printf("ID: %s", numericID)
		fmt.Printf(" • %s\n", formatTime(s.CreatedAt))

		// If summary is empty and we have messages, try to generate one
		if s.Summary == "" && len(s.Messages) > 0 && cfg != nil {
			// Try to generate a summary for display
			if err := session.UpdateSessionSummaryLLM(sessionMgr, s.ID, cfg); err == nil {
				// Reload the session to get the updated summary
				if updated, err := sessionMgr.GetSession(s.ID); err == nil {
					s.Summary = updated.Summary
				}
			}
		}

		if s.Summary == "" {
			s.Summary = "No summary available"
		}

		fmt.Printf("     ")
		util.SummaryColor.Printf("Summary: ")
		fmt.Printf("%s\n", s.Summary)

		fmt.Printf("     ")
		util.InfoColor.Printf("Messages: ")
		fmt.Printf("%d\n", len(s.Messages))

		// Add a blank line between sessions instead of dotted lines
		fmt.Println()
	}

	// Print command instructions with improved formatting
	util.HeaderColor.Println("Commands:")
	fmt.Println()
	util.CommandColor.Printf("  • ")
	fmt.Printf("Continue: ")
	util.HighlightColor.Printf("kiwi -s -o <id>\n")

	util.CommandColor.Printf("  • ")
	fmt.Printf("Delete:   ")
	util.HighlightColor.Printf("kiwi -s -d <id>\n")

	util.CommandColor.Printf("  • ")
	fmt.Printf("New:      ")
	util.HighlightColor.Printf("kiwi -s -n <name>\n")

	fmt.Println()

	return nil
}

func handleSessionsContinue(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("session ID required")
	}

	numericID := args[0]
	sessionID := getFullSessionID(numericID)

	// Continue the session by starting the assistant with this session
	return continueSession(cmd, sessionID)
}

func handleSessionsDelete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("session ID required")
	}

	numericID := args[0]
	sessionID := getFullSessionID(numericID)

	sessionMgr, err := session.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	if err := sessionMgr.DeleteSession(sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	fmt.Printf("Session %s deleted successfully.\n", numericID)
	return nil
}

func handleSessionsClear(cmd *cobra.Command, args []string) error {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	if err := sessionMgr.ClearSessions(); err != nil {
		return fmt.Errorf("failed to clear sessions: %w", err)
	}

	fmt.Println("All sessions cleared successfully.")
	return nil
}

// Continue an existing session
func continueSession(cmd *cobra.Command, sessionID string) error {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	// Verify the session exists
	sess, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session %s: %w", sessionID, err)
	}

	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Show the last 2 messages for context before continuing
	if len(sess.Messages) > 0 {
		fmt.Println("\nLast messages from this conversation:")
		fmt.Println("----------------------------------------")

		// Determine how many messages to show (up to last 2 exchanges, which is 4 messages)
		startIdx := 0
		if len(sess.Messages) > 4 {
			startIdx = len(sess.Messages) - 4
		}

		// Display the last messages
		for i := startIdx; i < len(sess.Messages); i++ {
			msg := sess.Messages[i]

			if msg.Role == "user" {
				util.UserColor.Print("You: ")
			} else if msg.Role == "assistant" {
				util.AssistantColor.Print("Kiwi: ")
			} else {
				continue // Skip system messages
			}

			// Print first 100 chars of message to avoid flooding terminal
			content := msg.Content
			if len(content) > 100 {
				content = content[:97] + "..."
			}
			fmt.Println(content)
		}
		fmt.Println("----------------------------------------")
	}

	// Use the unified session runner with isNewSession=false
	return sessionMgr.RunInteractiveSession(sessionID, cfg, false)
}

// Helper function to format time
func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006 15:04")
}

// Handle creating a new session with a custom name
func handleSessionsNew(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("session name required")
	}

	sessionName := args[0]

	// Create session ID with format custom_<name>_<timestamp>
	// This helps differentiate from auto-generated sessions and ensures uniqueness
	sessionID := fmt.Sprintf("custom_%s_%d", sessionName, time.Now().Unix())

	return startCustomSession(cmd, sessionID)
}

// Create and start an assistant session with the given session ID
// This refactored function is used by both startAssistant and handleSessionsNew
func startCustomSession(cmd *cobra.Command, sessionID string) error {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	// Create the session
	_, err = sessionMgr.CreateSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use the unified session runner with isNewSession=true
	return sessionMgr.RunInteractiveSession(sessionID, cfg, true)
}
