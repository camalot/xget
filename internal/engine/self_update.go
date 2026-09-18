package engine

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func isRunningExecutableDestination(destination, executable string, windows bool) bool {
	if !windows || executable == "" {
		return false
	}
	destinationPath, destinationErr := filepath.Abs(destination)
	executablePath, executableErr := filepath.Abs(executable)
	return destinationErr == nil && executableErr == nil && strings.EqualFold(destinationPath, executablePath)
}

func replaceStagedExecutable(destination string) error {
	safeDestination, err := cleanLocalPath(destination)
	if err != nil {
		return err
	}
	staged := safeDestination + ".new"
	previous := safeDestination + ".old"
	// #nosec G703 -- safeDestination is normalized and validated via cleanLocalPath.
	if err := os.Remove(previous); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// #nosec G703 -- safeDestination is normalized and validated via cleanLocalPath.
	if err := os.Rename(safeDestination, previous); err != nil {
		return err
	}
	// #nosec G703 -- safeDestination is normalized and validated via cleanLocalPath.
	if err := os.Rename(staged, safeDestination); err == nil {
		return nil
		// #nosec G703 -- safeDestination is normalized and validated via cleanLocalPath.
	} else if restoreErr := os.Rename(previous, safeDestination); restoreErr != nil {
		return errors.Join(err, restoreErr)
	} else {
		return err
	}
}

// RemovePreviousExecutable removes the executable retained by a successful
// Windows self-update. It is intentionally a no-op on other platforms.
func RemovePreviousExecutable() error {
	if runtime.GOOS != "windows" {
		return nil
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.Remove(executable + ".old"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
