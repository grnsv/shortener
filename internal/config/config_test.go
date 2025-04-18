package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := New()
	assert.Equal(t, "local", cfg.AppEnv)
	assert.Equal(t, "secret", cfg.JWTSecret)
	assert.Equal(t, NetAddress{"localhost", 8080}, cfg.ServerAddress)
	assert.Equal(t, BaseURI{"http://", NetAddress{"localhost", 8080}}, cfg.BaseAddress)
	assert.Equal(t, "", cfg.FileStoragePath)
	assert.Equal(t, "", cfg.DatabaseDSN)
}

func TestWithOptions(t *testing.T) {
	cfg := New(
		WithAppEnv("testenv"),
		WithJWTSecret("jwtkey"),
		WithServerAddress(NetAddress{"127.0.0.1", 9090}),
		WithBaseAddress(BaseURI{"https://", NetAddress{"test.com", 443}}),
		WithDatabaseDSN("postgresql://user:password@localhost:5432/dbname?sslmode=disable"),
	)
	assert.Equal(t, "testenv", cfg.AppEnv)
	assert.Equal(t, "jwtkey", cfg.JWTSecret)
	assert.Equal(t, NetAddress{"127.0.0.1", 9090}, cfg.ServerAddress)
	assert.Equal(t, BaseURI{"https://", NetAddress{"test.com", 443}}, cfg.BaseAddress)
	assert.Equal(t, "postgresql://user:password@localhost:5432/dbname?sslmode=disable", cfg.DatabaseDSN)
}

func TestNetAddressSetAndString(t *testing.T) {
	var addr NetAddress
	err := addr.Set("localhost:8081")
	assert.NoError(t, err)
	assert.Equal(t, NetAddress{"localhost", 8081}, addr)
	assert.Equal(t, "localhost:8081", addr.String())

	err = addr.Set("badaddress")
	assert.Error(t, err)

	err = addr.Set("localhost:port")
	assert.Error(t, err)

	err = addr.Set("localhost:70000")
	assert.Error(t, err)
}

func TestBaseURISetAndString(t *testing.T) {
	var uri BaseURI
	err := uri.Set("http://example.com:1234")
	assert.NoError(t, err)
	assert.Equal(t, BaseURI{"http://", NetAddress{"example.com", 1234}}, uri)
	assert.Equal(t, "http://example.com:1234", uri.String())

	err = uri.Set("ftp://example.com:21")
	assert.Error(t, err)
}

func TestParseEnvVariables(t *testing.T) {
	os.Setenv("APP_ENV", "envtest")
	os.Setenv("JWT_SECRET", "envsecret")
	os.Setenv("SERVER_ADDRESS", "127.0.0.1:9000")
	os.Setenv("BASE_URL", "https://env.com:443")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/testdata")
	os.Setenv("DATABASE_DSN", "postgresql://user:password@localhost:5432/dbname?sslmode=disable")
	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
	}()

	cfg := Parse()
	assert.Equal(t, "envtest", cfg.AppEnv)
	assert.Equal(t, "envsecret", cfg.JWTSecret)
	assert.Equal(t, NetAddress{"127.0.0.1", 9000}, cfg.ServerAddress)
	assert.Equal(t, BaseURI{"https://", NetAddress{"env.com", 443}}, cfg.BaseAddress)
	assert.Equal(t, "/tmp/testdata", cfg.FileStoragePath)
	assert.Equal(t, "postgresql://user:password@localhost:5432/dbname?sslmode=disable", cfg.DatabaseDSN)
}
