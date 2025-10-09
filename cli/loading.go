package cli

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unicode/utf8"
)

// AnimationStyle represents different types of loading animations
type AnimationStyle int

const (
	DotsStyle AnimationStyle = iota
	SpinnerStyle
	BarStyle
)

// supportsUnicode checks if the terminal supports Unicode characters
func supportsUnicode() bool {
	// Check if we're in a known Unicode-supporting terminal
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	// Most modern terminals support Unicode
	unicodeTerms := []string{"xterm", "screen", "tmux", "rxvt", "konsole", "gnome", "iterm"}
	for _, uterm := range unicodeTerms {
		if len(term) >= len(uterm) && term[:len(uterm)] == uterm {
			return true
		}
	}

	// Check for UTF-8 support
	return utf8.ValidString("⠋")
}

// LoadingAnimation represents a loading animation that can be started and stopped
type LoadingAnimation struct {
	message       string
	style         AnimationStyle
	completionMsg string
	stopCh        chan bool
	wg            sync.WaitGroup
	isRunning     bool
	mu            sync.Mutex
}

// NewLoadingAnimation creates a new loading animation with the given message
func NewLoadingAnimation(message string) *LoadingAnimation {
	return NewLoadingAnimationWithStyle(message, DotsStyle)
}

// NewLoadingAnimationWithStyle creates a new loading animation with the given message and style
func NewLoadingAnimationWithStyle(message string, style AnimationStyle) *LoadingAnimation {
	return &LoadingAnimation{
		message: message,
		style:   style,
		stopCh:  make(chan bool),
	}
}

// Start begins the loading animation
func (l *LoadingAnimation) Start() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isRunning {
		return // Already running
	}

	l.isRunning = true
	l.wg.Add(1)

	go l.animate()
}

// Stop stops the loading animation and optionally shows a completion message
func (l *LoadingAnimation) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isRunning {
		return // Not running
	}

	l.stopCh <- true
	l.wg.Wait()
	l.isRunning = false

	// Clear the current line
	fmt.Print("\r\033[K")

	// Show completion message if set
	if l.completionMsg != "" {
		fmt.Printf("%s\n", l.completionMsg)
	}
}

// StopWithMessage stops the animation and shows a custom completion message
func (l *LoadingAnimation) StopWithMessage(message string) {
	l.mu.Lock()
	l.completionMsg = message
	l.mu.Unlock()
	l.Stop()
}

// animate runs the actual animation loop
func (l *LoadingAnimation) animate() {
	defer l.wg.Done()

	var frames []string
	var tickerDuration time.Duration
	unicodeSupported := supportsUnicode()

	switch l.style {
	case SpinnerStyle:
		if unicodeSupported {
			frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		} else {
			frames = []string{"|", "/", "-", "\\"}
		}
		tickerDuration = 100 * time.Millisecond
	case BarStyle:
		if unicodeSupported {
			frames = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"}
		} else {
			frames = []string{"[    ]", "[=   ]", "[==  ]", "[=== ]", "[====]", "[=== ]", "[==  ]", "[=   ]"}
		}
		tickerDuration = 200 * time.Millisecond
	default: // DotsStyle
		frames = []string{"", ".", "..", "..."}
		tickerDuration = 300 * time.Millisecond
	}

	i := 0
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			// Clear current line and print message with animation
			if l.style == SpinnerStyle || l.style == BarStyle {
				fmt.Printf("\r\033[K%s %s", frames[i%len(frames)], l.message)
			} else {
				fmt.Printf("\r\033[K%s%s", l.message, frames[i%len(frames)])
			}
			i++
		}
	}
}

// StartLoadingAnimation is a convenience function that creates and starts a loading animation
func StartLoadingAnimation(message string) *LoadingAnimation {
	animation := NewLoadingAnimation(message)
	animation.Start()
	return animation
}

// StartLoadingAnimationWithStyle is a convenience function that creates and starts a loading animation with a specific style
func StartLoadingAnimationWithStyle(message string, style AnimationStyle) *LoadingAnimation {
	animation := NewLoadingAnimationWithStyle(message, style)
	animation.Start()
	return animation
}

// ShowLoadingWhile shows a loading animation while executing the provided function
func ShowLoadingWhile(message string, fn func()) {
	animation := StartLoadingAnimation(message)
	defer animation.Stop()
	fn()
}

// ShowLoadingWhileWithResult shows a loading animation while executing a function that returns a result and error
func ShowLoadingWhileWithResult[T any](message string, fn func() (T, error)) (T, error) {
	animation := StartLoadingAnimation(message)
	defer animation.Stop()
	return fn()
}

// ShowLoadingWhileWithCompletion shows a loading animation and displays a completion message
func ShowLoadingWhileWithCompletion(message string, completionMsg string, fn func()) {
	animation := StartLoadingAnimation(message)
	defer animation.StopWithMessage(completionMsg)
	fn()
}
