package config

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		env            map[string]string
		expectedConfig Config
	}{
		{
			name: "Default values",
			args: []string{"shortener.exe"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "./storage.json",
			},
		},
		{
			name: "Flags only",
			args: []string{"shortener.exe", "-a=localhost:9090", "-b=https://shortener"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:9090",
				BaseURL:         "https://shortener",
				LogLevel:        "info",
				FileStoragePath: "./storage.json",
			},
		},
		{
			name: "Environment variables only",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"SERVER_ADDRESS": "shortener:7070",
				"BASE_URL":       "https://shortener",
			},
			expectedConfig: Config{
				ServerAddr:      "shortener:7070",
				BaseURL:         "https://shortener",
				LogLevel:        "info",
				FileStoragePath: "./storage.json",
			},
		},
		{
			name: "Env for address, flag for base url",
			args: []string{"shortener.exe", "-b=https://shortener"},
			env: map[string]string{
				"SERVER_ADDRESS": "localhost:6060",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:6060",
				BaseURL:         "https://shortener",
				LogLevel:        "info",
				FileStoragePath: "./storage.json",
			},
		},
		{
			name: "Both environment and flags (use env)",
			args: []string{"shortener.exe", "-a=shortener:8080", "-b=http://shortener"},
			env: map[string]string{
				"SERVER_ADDRESS": "localhost:8080",
				"BASE_URL":       "http://localhost",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost",
				LogLevel:        "info",
				FileStoragePath: "./storage.json",
			},
		},
		{
			name: "Env for log level",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"LOG_LEVEL": "fatal",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "fatal",
				FileStoragePath: "./storage.json",
			},
		},
		{
			name: "Flag for log level",
			args: []string{"shortener.exe", "-l=fatal"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "fatal",
				FileStoragePath: "./storage.json",
			},
		},
		{
			name: "Env for file storage",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"FILE_STORAGE_PATH": "./storage_new.json",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "./storage_new.json",
			},
		},
		{
			name: "Flag for file storage",
			args: []string{"shortener.exe", "-f=./storage_new.json"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "./storage_new.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			for key, val := range tt.env {
				err := os.Setenv(key, val)
				if err != nil {
					t.Errorf("Failed to set env, error: %v", err)
					return
				}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			AppConfig = Config{}

			ParseConfig()

			if AppConfig != tt.expectedConfig {
				t.Errorf("Expected %+v, got %+v", tt.expectedConfig, AppConfig)
			}

			for key := range tt.env {
				err := os.Unsetenv(key)
				if err != nil {
					t.Errorf("Failed to unset env, error: %v", err)
					return
				}
			}
		})
	}
}
