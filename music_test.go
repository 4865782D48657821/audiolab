package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestListAlbums(t *testing.T) {
	root := t.TempDir()

	for _, name := range []string{"Album B", "Album A"} {
		if err := os.Mkdir(filepath.Join(root, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "Not an album.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := listAlbums(root)
	if err != nil {
		t.Fatalf("listAlbums() error = %v", err)
	}

	want := []string{"Album A", "Album B"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("listAlbums() = %v, want %v", got, want)
	}
}

func TestListAlbumsMissingDirectory(t *testing.T) {
	_, err := listAlbums(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("listAlbums() wollte für einen fehlenden Ordner keinen Fehler liefern")
	}
}

func TestValidateDestination(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "existierender Ordner", path: root},
		{name: "Datei statt Ordner", path: file, wantErr: true},
		{name: "fehlender Ordner", path: filepath.Join(root, "missing"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateDestination(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateDestination() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && !filepath.IsAbs(got) {
				t.Errorf("validateDestination() = %q, kein absoluter Pfad", got)
			}
		})
	}
}

func TestResolveAlbumPath(t *testing.T) {
	root := t.TempDir()
	albumDir := filepath.Join(root, "Album")
	if err := os.Mkdir(albumDir, 0755); err != nil {
		t.Fatal(err)
	}

	got, err := resolveAlbumPath(root, "Album")
	if err != nil {
		t.Fatalf("resolveAlbumPath() error = %v", err)
	}

	want, err := filepath.Abs(albumDir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("resolveAlbumPath() = %q, want %q", got, want)
	}
}

func TestResolveAlbumPathRejectsMissingAlbum(t *testing.T) {
	root := t.TempDir()

	_, err := resolveAlbumPath(root, "Missing")
	if err == nil {
		t.Fatal("resolveAlbumPath() wollte für ein fehlendes Album keinen Fehler liefern")
	}
}
