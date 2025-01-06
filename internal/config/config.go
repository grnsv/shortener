package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddress NetAddress
	BaseAddress   BaseURI
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
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
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
	b.Address.Set(strings.TrimPrefix(s, b.Scheme))
	return nil
}

var config *Config

func Get() Config {
	if config == nil {
		config = &Config{
			ServerAddress: NetAddress{"localhost", 8080},
			BaseAddress:   BaseURI{"http://", NetAddress{"localhost", 8080}},
		}
		flag.Var(&config.ServerAddress, "a", "Address for server")
		flag.Var(&config.BaseAddress, "b", "Base address for shorten url")
		flag.Parse()
	}
	return *config
}
