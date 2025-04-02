package util

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

// Minimum time between spinner transitions to prevent flickering
const minTransitionTime = 300 * time.Millisecond

// Debug mode - set to true to enable debug messages for spinner transitions
var debugSpinners = false

// Global spinner manager instance
var globalSpinnerManager = NewSpinnerManager()

// Spinner represents a simple text-based loading spinner
type Spinner struct {
	frames    []string
	message   string
	stopChan  chan struct{}
	waitGroup sync.WaitGroup
	running   bool
}

// NewSpinner creates a new spinner with the specified message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:   []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		message:  message,
		stopChan: make(chan struct{}),
		running:  false,
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	if s.running {
		return
	}
	s.running = true

	// Set up signal handling to properly clean up the spinner on program termination
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		s.Stop()
		os.Exit(1)
	}()

	s.waitGroup.Add(1)
	go func() {
		defer s.waitGroup.Done()
		frameIndex := 0
		for {
			select {
			case <-s.stopChan:
				fmt.Printf("\r%s\r", ClearLine())
				return
			default:
				frame := s.frames[frameIndex%len(s.frames)]
				fmt.Printf("\r%s %s", frame, s.message)
				frameIndex++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !s.running {
		return
	}

	// Use a new channel to avoid closing an already closed channel
	select {
	case <-s.stopChan:
		// Channel already closed, nothing to do
	default:
		close(s.stopChan)
	}

	s.waitGroup.Wait()
	s.running = false

	// Always make sure to clear the line after stopping
	fmt.Printf("\r%s\r", ClearLine())
}

// SetMessage updates the spinner's message
func (s *Spinner) SetMessage(message string) {
	s.message = message
}

// ClearLine returns a string that, when printed, clears the current line
func ClearLine() string {
	return "\033[2K"
}

// SpinnerManager ensures only one spinner is active at a time
// It acts as a singleton manager for all spinners in the application
type SpinnerManager struct {
	activeSpinner  *Spinner
	mutex          sync.Mutex
	currentState   SpinnerState
	lastTransition time.Time
	locked         bool
}

// NewSpinnerManager creates a new spinner manager
func NewSpinnerManager() *SpinnerManager {
	return &SpinnerManager{
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

// ensureTransitionDelay ensures enough time has passed between transitions
func (sm *SpinnerManager) ensureTransitionDelay() {
	elapsed := time.Since(sm.lastTransition)
	if elapsed < minTransitionTime {
		delay := minTransitionTime - elapsed
		if debugSpinners {
			fmt.Printf("\n[SPINNER] Delaying transition by %s\n", delay)
		}
		time.Sleep(delay)
	}
}

// LockSpinners prevents any spinner state changes
func (sm *SpinnerManager) LockSpinners() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.locked = true

	sm.logTransition(sm.currentState, sm.currentState, "LOCK - no transitions allowed")

	// Stop any active spinner when locking
	if sm.activeSpinner != nil {
		sm.activeSpinner.Stop()
		sm.activeSpinner = nil
	}
}

// UnlockSpinners allows spinner state changes again
func (sm *SpinnerManager) UnlockSpinners() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.locked = false
	sm.logTransition(sm.currentState, sm.currentState, "UNLOCK - transitions allowed")
}

// StartThinkingSpinner starts the main thinking spinner
func (sm *SpinnerManager) StartThinkingSpinner(message string) *Spinner {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED thinking: "+message)
		return nil
	}

	// Only start if we're not already in this state
	if sm.currentState == SpinnerStateThinking && sm.activeSpinner != nil {
		// Just update the message
		sm.activeSpinner.SetMessage(message)
		sm.logTransition(sm.currentState, sm.currentState, "UPDATE thinking: "+message)
		return sm.activeSpinner
	}

	// Ensure minimum time between transitions
	sm.ensureTransitionDelay()

	// Stop any active spinner first
	if sm.activeSpinner != nil {
		sm.activeSpinner.Stop()
		sm.activeSpinner = nil
	}

	oldState := sm.currentState
	sm.currentState = SpinnerStateThinking
	sm.logTransition(oldState, sm.currentState, "START thinking: "+message)

	// Create and start a new spinner
	spinner := NewSpinner(message)
	spinner.Start()
	sm.activeSpinner = spinner
	sm.lastTransition = time.Now()

	return spinner
}

// StartToolSpinner starts a tool execution spinner
func (sm *SpinnerManager) StartToolSpinner(message string) *Spinner {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED tool: "+message)
		return nil
	}

	// Only start if we're not already in this state
	if sm.currentState == SpinnerStateTool && sm.activeSpinner != nil {
		// Just update the message
		sm.activeSpinner.SetMessage(message)
		sm.logTransition(sm.currentState, sm.currentState, "UPDATE tool: "+message)
		return sm.activeSpinner
	}

	// Ensure minimum time between transitions
	sm.ensureTransitionDelay()

	// Stop any active spinner first
	if sm.activeSpinner != nil {
		sm.activeSpinner.Stop()
		sm.activeSpinner = nil
	}

	oldState := sm.currentState
	sm.currentState = SpinnerStateTool
	sm.logTransition(oldState, sm.currentState, "START tool: "+message)

	// Create and start a new spinner
	spinner := NewSpinner(message)
	spinner.Start()
	sm.activeSpinner = spinner
	sm.lastTransition = time.Now()

	return spinner
}

// TransitionToResponse stops the spinner for showing a response
func (sm *SpinnerManager) TransitionToResponse() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED transition to response")
		return
	}

	// No need to transition if we're already in this state
	if sm.currentState == SpinnerStateNone {
		return
	}

	// Ensure minimum time between transitions
	sm.ensureTransitionDelay()

	oldState := sm.currentState

	// Stop any active spinner
	if sm.activeSpinner != nil {
		sm.activeSpinner.Stop()
		sm.activeSpinner = nil
	}

	sm.currentState = SpinnerStateNone
	sm.lastTransition = time.Now()

	sm.logTransition(oldState, sm.currentState, "TRANSITION to response")
}

// StopAllSpinners stops any active spinner and clears the state
func (sm *SpinnerManager) StopAllSpinners() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// No need to do anything if there's no active spinner
	if sm.activeSpinner == nil {
		return
	}

	oldState := sm.currentState

	// Stop the active spinner
	sm.activeSpinner.Stop()
	sm.activeSpinner = nil
	sm.currentState = SpinnerStateNone
	sm.lastTransition = time.Now()

	sm.logTransition(oldState, sm.currentState, "STOP all spinners")
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

// Ensure the spinner remains active during all stages of processing
// Start the spinner at the beginning of any long-running process
func (sm *SpinnerManager) StartProcessingSpinner(message string) *Spinner {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED processing: "+message)
		return nil
	}

	// Only start if we're not already in this state
	if sm.currentState == SpinnerStateTool && sm.activeSpinner != nil {
		// Just update the message
		sm.activeSpinner.SetMessage(message)
		sm.logTransition(sm.currentState, sm.currentState, "UPDATE processing: "+message)
		return sm.activeSpinner
	}

	// Ensure minimum time between transitions
	sm.ensureTransitionDelay()

	// Stop any active spinner first
	if sm.activeSpinner != nil {
		sm.activeSpinner.Stop()
		sm.activeSpinner = nil
	}

	oldState := sm.currentState
	sm.currentState = SpinnerStateTool
	sm.logTransition(oldState, sm.currentState, "START processing: "+message)

	// Create and start a new spinner
	spinner := NewSpinner(message)
	spinner.Start()
	sm.activeSpinner = spinner
	sm.lastTransition = time.Now()

	return spinner
}

// Stop the spinner once the process is complete
func (sm *SpinnerManager) StopProcessingSpinner() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.locked {
		sm.logTransition(sm.currentState, sm.currentState, "BLOCKED stop processing")
		return
	}

	// No need to transition if we're already in this state
	if sm.currentState == SpinnerStateNone {
		return
	}

	// Ensure minimum time between transitions
	sm.ensureTransitionDelay()

	oldState := sm.currentState

	// Stop any active spinner
	if sm.activeSpinner != nil {
		sm.activeSpinner.Stop()
		sm.activeSpinner = nil
	}

	sm.currentState = SpinnerStateNone
	sm.lastTransition = time.Now()

	sm.logTransition(oldState, sm.currentState, "STOP processing")
}
