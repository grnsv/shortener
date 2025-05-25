// Package config provides configuration management for the application.
// It supports loading configuration from environment variables and command-line flags.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config holds the application configuration loaded from environment variables and flags.
type Config struct {
	AppEnv          string     `env:"APP_ENV" json:"app_env"`                     // Application environment (e.g., local, production)
	JWTSecret       string     `env:"JWT_SECRET" json:"jwt_secret"`               // Secret key for JWT authentication
	ServerAddress   NetAddress `env:"SERVER_ADDRESS" json:"server_address"`       // Address for the HTTP server
	BaseURL         BaseURL    `env:"BASE_URL" json:"base_url"`                   // Base URL for shortened links
	FileStoragePath string     `env:"FILE_STORAGE_PATH" json:"file_storage_path"` // Path to file storage
	DatabaseDSN     string     `env:"DATABASE_DSN" json:"database_dsn"`           // Database connection string
	EnableHTTPS     bool       `env:"ENABLE_HTTPS" json:"enable_https"`           // Enable HTTPS flag
	CertFile        string     `env:"CERT_FILE" json:"cert_file"`                 // Cert file
	KeyFile         string     `env:"KEY_FILE" json:"key_file"`                   // Key file
	Config          string     `env:"CONFIG"`                                     // Config file
	TrustedSubnet   string     `env:"TRUSTED_SUBNET" json:"trusted_subnet"`       // Trusted subnet
}

// NetAddress represents a network address with a host and port.
type NetAddress struct {
	Host string // Hostname or IP address
	Port int    // Port number
}

// String returns the NetAddress as "host:port".
// It implements the flag.Value interface, allowing it to be used as a command-line flag.
func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

// Set parses and sets the NetAddress from a "host:port" string.
// It implements the flag.Value interface, allowing it to be used as a command-line flag.
func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("address must be in a form host:port")
	}

	_, err := net.LookupHost(hp[0])
	if err != nil {
		return fmt.Errorf("host is invalid or unreachable: %w", err)
	}

	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, %d given", port)
	}

	a.Host = hp[0]
	a.Port = port
	return nil
}

// UnmarshalText parses the NetAddress from text.
// It implements the encoding.TextUnmarshaler interface, allowing it to be used as an environment variable.
func (a *NetAddress) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

// BaseURL represents a base URL (scheme + host:port).
type BaseURL struct {
	Scheme  string     // URL scheme (e.g., "http://", "https://").
	Address NetAddress // NetAddress struct.
}

// String returns the BaseURL as "scheme+host:port".
// It implements the flag.Value interface, allowing it to be used as a command-line flag.
func (b BaseURL) String() string {
	return b.Scheme + b.Address.String()
}

// Set parses and sets the BaseURL from a "scheme+host:port" string.
// It implements the flag.Value interface, allowing it to be used as a command-line flag.
func (b *BaseURL) Set(s string) error {
	if strings.HasPrefix(s, "http://") {
		b.Scheme = "http://"
	} else if strings.HasPrefix(s, "https://") {
		b.Scheme = "https://"
	} else {
		return errors.New(`URL scheme must be "http://" or "https://"`)
	}
	err := b.Address.Set(strings.TrimPrefix(s, b.Scheme))
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalText parses the BaseURL from text.
// It implements the encoding.TextUnmarshaler interface, allowing it to be used as an environment variable.
func (b *BaseURL) UnmarshalText(text []byte) error {
	return b.Set(string(text))
}

// Option is a function that applies a configuration option to Config.
type Option func(*Config)

// WithAppEnv sets the application environment in the Config.
func WithAppEnv(appEnv string) Option {
	return func(c *Config) {
		c.AppEnv = appEnv
	}
}

// WithJWTSecret sets the JWT secret in the Config.
func WithJWTSecret(key string) Option {
	return func(c *Config) {
		c.JWTSecret = key
	}
}

// WithServerAddress sets the server address in the Config.
func WithServerAddress(addr NetAddress) Option {
	return func(c *Config) {
		c.ServerAddress = addr
	}
}

// WithBaseURL sets the base address in the Config.
func WithBaseURL(url BaseURL) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithDatabaseDSN sets the database DSN in the Config.
func WithDatabaseDSN(dsn string) Option {
	return func(c *Config) {
		c.DatabaseDSN = dsn
	}
}

// New creates a new Config with the provided options.
func New(opts ...Option) *Config {
	for _, opt := range opts {
		opt(config)
	}
	return config
}

var config = &Config{
	AppEnv:          "local",
	JWTSecret:       "secret",
	ServerAddress:   NetAddress{"localhost", 8080},
	BaseURL:         BaseURL{"http://", NetAddress{"localhost", 8080}},
	FileStoragePath: "",
	DatabaseDSN:     "",
	CertFile:        "../../certs/cert.pem",
	KeyFile:         "../../certs/key.pem",
}

func parseFlags() error {
	set := flag.NewFlagSet("set", flag.ExitOnError)
	set.Var(&config.ServerAddress, "a", "Address for server")
	set.Var(&config.BaseURL, "b", "Base URL for shorten url")
	set.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "File storage path (/data/storage)")
	set.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "Database DSN (postgresql://user:password@host:port/dbname?sslmode=disable)")
	set.BoolVar(&config.EnableHTTPS, "s", config.EnableHTTPS, "Enable HTTPS")
	set.StringVar(&config.Config, "c", config.Config, "Config file")
	set.StringVar(&config.Config, "config", config.Config, "Config file")
	set.StringVar(&config.TrustedSubnet, "t", config.TrustedSubnet, "Trusted subnet")
	return set.Parse(os.Args[1:])
}

// Parse loads configuration from flags and environment variables and returns a Config pointer.
func Parse() (*Config, error) {
	err := loadConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}
	err = parseFlags()
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}
	err = env.Parse(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}
	return config, nil
}

func loadConfigFile() (err error) {
	err = findConfigFile()
	if err != nil || config.Config == "" {
		return
	}
	f, err := os.Open(config.Config)
	if err != nil {
		return
	}
	return json.NewDecoder(f).Decode(config)
}

func findConfigFile() error {
	config.Config = os.Getenv("CONFIG")
	if config.Config != "" {
		return nil
	}
	return parseFlags()
}
