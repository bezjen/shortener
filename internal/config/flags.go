// Package config provides configuration management for the URL shortening service.
package config

import (
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all application configuration settings.
type Config struct {
	ServerAddr      string `mapstructure:"server_address" json:"server_address"`
	BaseURL         string `mapstructure:"base_url" json:"base_url"`
	LogLevel        string `mapstructure:"log_level" json:"log_level"`
	FileStoragePath string `mapstructure:"file_storage_path" json:"file_storage_path"`
	DatabaseDSN     string `mapstructure:"database_dsn" json:"database_dsn"`
	SecretKey       string `mapstructure:"secret_key" json:"secret_key"`
	AuditFile       string `mapstructure:"audit_file" json:"audit_file"`
	AuditURL        string `mapstructure:"audit_url" json:"audit_url"`
	EnableHTTPS     bool   `mapstructure:"enable_https" json:"enable_https"`
	TrustedSubnet   string `mapstructure:"trusted_subnet" json:"trusted_subnet"`
}

// AppConfig is the global application configuration instance.
var AppConfig Config

// ParseConfig parses command-line flags, environment variables, and config files to populate AppConfig.
// Priority (Standard Viper):
// 1. Command-line flags
// 2. Environment variables
// 3. Configuration file
// 4. Default values
func ParseConfig() {
	// 1. Set Defaults
	viper.SetDefault("server_address", "localhost:8080")
	viper.SetDefault("base_url", "http://localhost:8080")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("enable_https", false)

	// 2. Define Flags
	// Check if flags are already defined to avoid "flag redefined" panic in tests if ParseConfig is called multiple times without reset
	if pflag.Lookup("a") == nil {
		pflag.StringP("a", "a", "", "port to run server")
		pflag.StringP("b", "b", "", "address and port of tiny url")
		pflag.StringP("l", "l", "", "log level")
		pflag.StringP("f", "f", "", "path to file with data")
		pflag.StringP("d", "d", "", "postgres data source name")
		pflag.StringP("k", "k", "", "authorization secret key")
		pflag.String("audit-file", "", "path to audit file")
		pflag.String("audit-url", "", "audit url")
		pflag.BoolP("s", "s", false, "enable https")
		pflag.StringP("t", "t", "", "trusted subnet (CIDR)")
		pflag.StringP("config", "c", "", "path to config file")
	}

	// Parse the flags
	// Note: In tests we set ContinueOnError, so this won't exit the app on bad flags
	pflag.Parse()

	// 3. Bind Flags to Viper Keys
	bindFlag("server_address", "a")
	bindFlag("base_url", "b")
	bindFlag("log_level", "l")
	bindFlag("file_storage_path", "f")
	bindFlag("database_dsn", "d")
	bindFlag("secret_key", "k")
	bindFlag("audit_file", "audit-file")
	bindFlag("audit_url", "audit-url")
	bindFlag("enable_https", "s")
	bindFlag("trusted_subnet", "t")

	// 4. Bind Environment Variables
	viper.AutomaticEnv()
	bindEnv("server_address", "SERVER_ADDRESS")
	bindEnv("base_url", "BASE_URL")
	bindEnv("log_level", "LOG_LEVEL")
	bindEnv("file_storage_path", "FILE_STORAGE_PATH")
	bindEnv("database_dsn", "DATABASE_DSN")
	bindEnv("secret_key", "SECRET_KEY")
	bindEnv("audit_file", "AUDIT_FILE")
	bindEnv("audit_url", "AUDIT_URL")
	bindEnv("enable_https", "ENABLE_HTTPS")
	bindEnv("config", "CONFIG")
	bindEnv("trusted_subnet", "TRUSTED_SUBNET")

	// 5. Load Config File
	cfgPath, _ := pflag.CommandLine.GetString("config")
	if cfgPath == "" {
		cfgPath = viper.GetString("config")
	}

	if cfgPath != "" {
		viper.SetConfigFile(cfgPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("Error reading config file: %s", err)
		}
	}

	// 6. Unmarshal
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
}

func bindFlag(key, flagName string) {
	if err := viper.BindPFlag(key, pflag.Lookup(flagName)); err != nil {
		log.Printf("Error binding flag %s: %v", flagName, err)
	}
}

func bindEnv(key, envName string) {
	if err := viper.BindEnv(key, envName); err != nil {
		log.Printf("Error binding env %s: %v", envName, err)
	}
}
