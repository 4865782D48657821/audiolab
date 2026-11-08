package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultMusicDir              = "/home/fast/Music"
	defaultMountedDeviceMusicDir = "/mnt/ak100/Music"
)

type config struct {
	DefaultMusicDir              string `json:"default_music_dir"`
	DefaultMountedDeviceMusicDir string `json:"default_mounted_device_music_dir"`
}

func loadConfig() (config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return config{}, err
	}

	dir := filepath.Join(configDir, "audiolab")
	file := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(file)

	if err == nil {
		var cfg config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return config{}, fmt.Errorf("Config-Datei lesen: %w", err)
		}

		if cfg.DefaultMusicDir == "" {
			cfg.DefaultMusicDir = defaultMusicDir
		}
		if cfg.DefaultMountedDeviceMusicDir == "" {
			cfg.DefaultMountedDeviceMusicDir = defaultMountedDeviceMusicDir
		}

		return cfg, nil
	}

	if !os.IsNotExist(err) {
		return config{}, fmt.Errorf("Config-Datei öffnen: %w", err)
	}

	cfg := config{
		DefaultMusicDir:              "/home/fast/Music",
		DefaultMountedDeviceMusicDir: "/mnt/ak100/Music",
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return config{}, fmt.Errorf("Config-Ordner erstellen: %w", err)
	}

	data, err = json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return config{}, fmt.Errorf("Config erzeugen: %w", err)
	}

	if err := os.WriteFile(file, append(data, '\n'), 0644); err != nil {
		return config{}, fmt.Errorf("Config-Datei speichern: %w", err)
	}

	return cfg, nil
}
