package config

import (
	"flag"
	"os"
	"testing"
)

// helper to create a temp config file
func createTempConfig(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpFile.Name()
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		env            map[string]string
		jsonContent    string
		useConfigFlag  string
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
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
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
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Env for data source name",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"DATABASE_DSN": "ds",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "ds",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Flag for data source name",
			args: []string{"shortener.exe", "-d=ds"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "ds",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Env for secret key",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"SECRET_KEY": "secret_key",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "secret_key",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Flag for secret key",
			args: []string{"shortener.exe", "-k=secret_key1"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "secret_key1",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Env for audit file",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"AUDIT_FILE": "file_path",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "file_path",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Flag for audit file",
			args: []string{"shortener.exe", "--audit-file=file_path"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "file_path",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Env for audit url",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"AUDIT_URL": "audit_url",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "audit_url",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Flag for audit url",
			args: []string{"shortener.exe", "--audit-url=audit_url"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "audit_url",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Env for enable https true",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"ENABLE_HTTPS": "true",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     true,
			},
		},
		{
			name: "Env for enable https not true",
			args: []string{"shortener.exe"},
			env: map[string]string{
				"ENABLE_HTTPS": "false",
			},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Flag for enable https true",
			args: []string{"shortener.exe", "--s=true"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     true,
			},
		},
		{
			name: "Flag for enable https not true",
			args: []string{"shortener.exe", "--s=bad"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Config file only (via -c)",
			args: []string{"shortener.exe"}, // flag added dynamically
			env:  map[string]string{},
			jsonContent: `{
				"server_address": "localhost:5555",
				"base_url": "http://config-file",
				"enable_https": true
			}`,
			useConfigFlag: "-c",
			expectedConfig: Config{
				ServerAddr:      "localhost:5555",
				BaseURL:         "http://config-file",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     true,
			},
		},
		{
			name: "Config file only (via -config)",
			args: []string{"shortener.exe"},
			env:  map[string]string{},
			jsonContent: `{
				"server_address": "localhost:4444"
			}`,
			useConfigFlag: "-config",
			expectedConfig: Config{
				ServerAddr:      "localhost:4444",
				BaseURL:         "http://localhost:8080", // default kept
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Config file only (via ENV)",
			args: []string{"shortener.exe"},
			env:  map[string]string{}, // CONFIG env added dynamically
			jsonContent: `{
				"log_level": "debug"
			}`,
			useConfigFlag: "env",
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "debug",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Priority: Flag > Config File",
			args: []string{"shortener.exe", "-a=flag:1111"},
			env:  map[string]string{},
			jsonContent: `{
				"server_address": "json:2222",
				"base_url": "http://json-url"
			}`,
			useConfigFlag: "-c",
			expectedConfig: Config{
				ServerAddr:      "flag:1111",       // Flag wins
				BaseURL:         "http://json-url", // JSON used because flag not set
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Priority: Env > Flag > Config File",
			args: []string{"shortener.exe", "-a=flag:1111", "-b=http://flag-url"},
			env: map[string]string{
				"SERVER_ADDRESS": "env:0000",
			},
			jsonContent: `{
				"server_address": "json:2222",
				"base_url": "http://json-url",
				"log_level": "warn"
			}`,
			useConfigFlag: "-c",
			expectedConfig: Config{
				ServerAddr:      "env:0000",        // Env wins over Flag and JSON
				BaseURL:         "http://flag-url", // Flag wins over JSON
				LogLevel:        "warn",            // JSON used
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false,
			},
		},
		{
			name: "Boolean logic: JSON true, Flag default (false)",
			args: []string{"shortener.exe"},
			env:  map[string]string{},
			jsonContent: `{
				"enable_https": true
			}`,
			useConfigFlag: "-c",
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     true, // JSON should win over default flag
			},
		},
		{
			name: "Boolean logic: JSON true, Flag explicitly false",
			args: []string{"shortener.exe", "-s=false"},
			env:  map[string]string{},
			jsonContent: `{
				"enable_https": true
			}`,
			useConfigFlag: "-c",
			expectedConfig: Config{
				ServerAddr:      "localhost:8080",
				BaseURL:         "http://localhost:8080",
				LogLevel:        "info",
				FileStoragePath: "",
				DatabaseDSN:     "",
				SecretKey:       "",
				AuditFile:       "",
				AuditURL:        "",
				EnableHTTPS:     false, // Explicit flag should win over JSON
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Config File if needed
			var tmpFile string
			if tt.jsonContent != "" {
				tmpFile = createTempConfig(t, tt.jsonContent)
				defer os.Remove(tmpFile)

				switch tt.useConfigFlag {
				case "-c":
					tt.args = append(tt.args, "-c="+tmpFile)
				case "-config":
					tt.args = append(tt.args, "-config="+tmpFile)
				case "env":
					tt.env["CONFIG"] = tmpFile
				}
			}

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
