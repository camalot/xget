package config

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/joho/godotenv"
)

// LoadDotenvFiles loads environment files in descending priority without
// replacing values already provided by the parent process.
func LoadDotenvFiles() error {
	directory, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, path := range dotenvPaths(directory) {
		if err := godotenv.Load(path); err != nil {
			return err
		}
	}
	return nil
}

func dotenvPaths(directory string) []string {
	paths := make([]string, 0, 4)
	for _, pattern := range []string{".secrets", "*.secrets", ".env", "*.env"} {
		if pattern[0] == '.' {
			path := filepath.Join(directory, pattern)
			if _, err := os.Stat(path); err == nil {
				paths = append(paths, path)
			}
			continue
		}
		matches, err := filepath.Glob(filepath.Join(directory, pattern))
		if err != nil {
			continue
		}
		sort.Strings(matches)
		for _, path := range matches {
			base := filepath.Base(path)
			if base != ".secrets" && base != ".env" {
				paths = append(paths, path)
			}
		}
	}
	return paths
}
