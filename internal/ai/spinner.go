package ai

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Spinner shows a terminal spinner animation while work runs.
// Returns a done function that must be called to stop the spinner.
func StartSpinner(msg string) (done func()) {
	stop := make(chan struct{})
	done = func() {
		close(stop)
	}

	if isCI() || isPiped() {
		fmt.Fprintf(os.Stderr, "  %s... ", msg)
		return done
	}

	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		i := 0
		for {
			select {
			case <-stop:
				clearLine()
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %s... ", chars[i%len(chars)], msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	// Give a small delay so the spinner is seen before work completes
	return done
}

// PrintDone prints a completion message after spinner stops.
func PrintDone(msg string) {
	fmt.Fprintf(os.Stderr, "%s done.\n", msg)
}

func clearLine() {
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 60))
}

func isCI() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
}

func isPiped() bool {
	stat, _ := os.Stderr.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
