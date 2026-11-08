package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func listAlbums(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var albums []string
	for _, entry := range entries {
		if entry.IsDir() {
			albums = append(albums, entry.Name())
		}
	}

	return albums, nil
}

func validateDestination(path string) (string, error) {
	destination, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Zielpfad konnte nicht aufgelöst werden: %w", err)
	}

	info, err := os.Stat(destination)
	if err != nil {
		return "", fmt.Errorf("Zielordner nicht erreichbar: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("der Zielpfad ist kein Ordner")
	}

	return destination, nil
}
