package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// spinnerCharSet renders the phases of the moon as a lightweight progress indicator.
var spinnerCharSet = []string{"◐", "◓", "◑", "◒"}

// progress reports what a long-running refresh loop is currently checking, so
// network-bound commands like `list --installed` and `upgrade` don't look hung.
type progress interface {
	Update(message string)
	Stop()
}

// newProgress is indirected for testing.
var newProgress = newSpinnerProgress

type spinnerProgress struct {
	sp *spinner.Spinner
}

// newSpinnerProgress starts a spinner writing to stderr, so it never mixes with
// a command's stdout output. The underlying library disables itself automatically
// when stderr isn't a terminal (e.g. output is piped, redirected, or under test).
func newSpinnerProgress() progress {
	sp := spinner.New(spinnerCharSet, 120*time.Millisecond, spinner.WithWriterFile(os.Stderr))
	sp.Start()
	return &spinnerProgress{sp: sp}
}

func (p *spinnerProgress) Update(message string) {
	p.sp.Lock()
	p.sp.Suffix = " " + message
	p.sp.Unlock()
}

// Stop halts the spinner and erases its line so it doesn't leave a stray line
// behind once the refresh loop finishes.
func (p *spinnerProgress) Stop() {
	p.sp.Stop()
}

func checkingMessage(name string) string {
	return fmt.Sprintf("checking %s", name)
}
