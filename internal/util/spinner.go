package util

import (
	"fmt"
	"sync"
	"time"

	bs "github.com/briandowns/spinner"
)

// SpinnerState represents the current state of spinners in the application
type SpinnerState int

const (
	// SpinnerStateNone means no spinner is active
	SpinnerStateNone SpinnerState = iota
	// SpinnerStateThinking means the main thinking spinner is active
	SpinnerStateThinking
	// SpinnerStateTool means a tool execution spinner is active
	SpinnerStateTool
)

// Debug mode - set to true to enable debug messages for spinner transitions
var debugSpinners = false

// Global spinner manager instance
var globalSpinnerManager = NewSpinnerManager()

// SpinnerManager ensures only one spinner is active at a time
// It acts as a singleton manager for all spinners in the application
type SpinnerManager struct {
	spinner        *bs.Spinner
	mutex          sync.Mutex
	currentState   SpinnerState
	currentMsg     string
	lastMsg        string
	lastTransition time.Time
	locked         bool
}

// NewSpinnerManager creates a new spinner manager
func NewSpinnerManager() *SpinnerManager {
	return &SpinnerManager{
		spinner:        nil,
		currentState:   SpinnerStateNone,
		lastTransition: time.Now().Add(-1 * time.Second), // Initialize with past time
		locked:         false,
	}
}

// logTransition logs a transition for debugging purposes
func (sm *SpinnerManager) logTransition(from, to SpinnerState, message string) {
	if debugSpinners {
		stateNames := map[SpinnerState]string{
			SpinnerStateNone:     "None",
			SpinnerStateThinking: "Thinking",
			SpinnerStateTool:     "Tool",
		}

		fromStr := stateNames[from]
		toStr := stateNames[to]

		fmt.Printf("\n[SPINNER] %s -> %s: %s\n", fromStr, toStr, message)
	}
}

// LockSpinners prevents any spinner state changes
func (sm *SpinnerManager) LockSpinners() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.locked = true

	sm.logTransition(sm.currentState, sm.currentState, "LOCK - no transitions allowed")

	// Stop any active spinner when locking
	if sm.spinner != nil {
		sm.spinner.Stop()
		sm.spinner = nil
	}
}

// UnlockSpinners allows spinner state changes again
func (sm *SpinnerManager) UnlockSpinners() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.locked = false
	sm.logTransition(sm.currentState, sm.currentState, "UNLOCK - transitions allowed")
}

// clearSpinner stops the current spinner if any
func (sm *SpinnerManager) clearSpinner() {
	if sm.spinner != nil {
		sm.spinner.Stop()
		// Clear the line to make sure no text is left
		fmt.Print("\r\033[K")
		sm.spinner = nil
	}
}

// StartThinkingSpinner starts the main thinking spinner
func (sm *SpinnerManager) StartThinkingSpinner(message string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED thinking: "+message)
		return
	}

	// Clear any existing spinner
	sm.clearSpinner()

	// Create and start a new spinner
	s := bs.New(bs.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("cyan")
	s.Start()

	sm.spinner = s
	sm.currentMsg = message
	sm.currentState = SpinnerStateThinking
	sm.lastTransition = time.Now()

	sm.logTransition(sm.currentState, SpinnerStateThinking, "START thinking: "+message)
}

// StartToolSpinner starts a tool execution spinner
func (sm *SpinnerManager) StartToolSpinner(message string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED tool: "+message)
		return
	}

	// Store the last message before clearing
	sm.lastMsg = sm.currentMsg

	// Clear any existing spinner
	sm.clearSpinner()

	// Create and start a new spinner
	s := bs.New(bs.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("yellow")
	s.Start()

	sm.spinner = s
	sm.currentMsg = message
	sm.currentState = SpinnerStateTool
	sm.lastTransition = time.Now()

	sm.logTransition(sm.currentState, SpinnerStateTool, "START tool: "+message)
}

// TransitionToResponse stops the spinner for showing a response
func (sm *SpinnerManager) TransitionToResponse() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED transition to response")
		return
	}

	oldState := sm.currentState

	// Clear any existing spinner
	sm.clearSpinner()

	sm.currentState = SpinnerStateNone
	sm.lastTransition = time.Now()

	sm.logTransition(oldState, SpinnerStateNone, "TRANSITION to response")
}

// StopAllSpinners stops any active spinner and clears the state
func (sm *SpinnerManager) StopAllSpinners() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	oldState := sm.currentState

	// Clear any existing spinner
	sm.clearSpinner()

	sm.currentState = SpinnerStateNone
	sm.lastTransition = time.Now()

	sm.logTransition(oldState, SpinnerStateNone, "STOP all spinners")
}

// GetCurrentState returns the current spinner state
func (sm *SpinnerManager) GetCurrentState() SpinnerState {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return sm.currentState
}

// GetGlobalSpinnerManager returns the global spinner manager instance
// This makes it easy to access the same spinner manager throughout the application
func GetGlobalSpinnerManager() *SpinnerManager {
	return globalSpinnerManager
}

// StartProcessingSpinner starts a processing spinner
func (sm *SpinnerManager) StartProcessingSpinner(message string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED processing: "+message)
		return
	}

	// Clear any existing spinner
	sm.clearSpinner()

	// Create and start a new spinner
	s := bs.New(bs.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("yellow")
	s.Start()

	sm.spinner = s
	sm.currentMsg = message
	sm.currentState = SpinnerStateTool
	sm.lastTransition = time.Now()

	sm.logTransition(sm.currentState, SpinnerStateTool, "START processing: "+message)
}

// StopProcessingSpinner stops the processing spinner
func (sm *SpinnerManager) StopProcessingSpinner() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED stop processing")
		return
	}

	oldState := sm.currentState

	// Clear any existing spinner
	sm.clearSpinner()

	sm.currentState = SpinnerStateNone
	sm.lastTransition = time.Now()

	sm.logTransition(oldState, SpinnerStateNone, "STOP processing")
}

// ClearSpinner stops all spinners and clears the line
func ClearSpinner(spinnerManager *SpinnerManager) {
	spinnerManager.StopAllSpinners()
}

// PrepareForResponse handles the common pattern of stopping spinners
// before displaying a response
func PrepareForResponse(spinnerManager *SpinnerManager) {
	ClearSpinner(spinnerManager)
}
