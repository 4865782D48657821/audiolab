package main

import "os"

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
