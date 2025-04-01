package util

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

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
				fmt.Printf("\r%s\r", clearLine())
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
	close(s.stopChan)
	s.waitGroup.Wait()
	s.running = false
}

// SetMessage updates the spinner's message
func (s *Spinner) SetMessage(message string) {
	s.message = message
}

// clearLine returns a string that, when printed, clears the current line
func clearLine() string {
	return "\033[2K"
}
