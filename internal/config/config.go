package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv          string     `env:"APP_ENV"`
	ServerAddress   NetAddress `env:"SERVER_ADDRESS"`
	BaseAddress     BaseURI    `env:"BASE_URL"`
	FileStoragePath string     `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string     `env:"DATABASE_DSN"`
}

type NetAddress struct {
	Host string
	Port int
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

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

func (a *NetAddress) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

type BaseURI struct {
	Scheme  string
	Address NetAddress
}

func (b BaseURI) String() string {
	return b.Scheme + b.Address.String()
}

func (b *BaseURI) Set(s string) error {
	if strings.HasPrefix(s, "http://") {
		b.Scheme = "http://"
	} else if strings.HasPrefix(s, "https://") {
		b.Scheme = "https://"
	} else {
		return errors.New(`URI scheme must be "http://" or "https://"`)
	}
	err := b.Address.Set(strings.TrimPrefix(s, b.Scheme))
	if err != nil {
		return err
	}
	return nil
}

func (b *BaseURI) UnmarshalText(text []byte) error {
	return b.Set(string(text))
}

var config *Config

func Get() Config {
	if config != nil {
		return *config
	}

	config = &Config{
		AppEnv:        "local",
		ServerAddress: NetAddress{"localhost", 8080},
		BaseAddress:   BaseURI{"http://", NetAddress{"localhost", 8080}},
	}
	flag.Var(&config.ServerAddress, "a", "Address for server")
	flag.Var(&config.BaseAddress, "b", "Base address for shorten url")
	flag.StringVar(&config.FileStoragePath, "f", "", "File storage path (/data/storage)")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Database DSN (postgresql://user:password@host:port/dbname?sslmode=disable")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		log.Fatalf("Failed to parse env: %v", err)
	}

	return *config
}
