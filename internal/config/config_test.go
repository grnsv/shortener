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
	assert.Equal(t, BaseURL{"http://", NetAddress{"localhost", 8080}}, cfg.BaseURL)
	assert.Equal(t, "", cfg.FileStoragePath)
	assert.Equal(t, "", cfg.DatabaseDSN)
}

func TestWithOptions(t *testing.T) {
	cfg := New(
		WithAppEnv("testenv"),
		WithJWTSecret("jwtkey"),
		WithServerAddress(NetAddress{"127.0.0.1", 9090}),
		WithBaseURL(BaseURL{"https://", NetAddress{"test.com", 443}}),
		WithDatabaseDSN("postgresql://user:password@localhost:5432/dbname?sslmode=disable"),
	)
	assert.Equal(t, "testenv", cfg.AppEnv)
	assert.Equal(t, "jwtkey", cfg.JWTSecret)
	assert.Equal(t, NetAddress{"127.0.0.1", 9090}, cfg.ServerAddress)
	assert.Equal(t, BaseURL{"https://", NetAddress{"test.com", 443}}, cfg.BaseURL)
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

func TestBaseURLSetAndString(t *testing.T) {
	var uri BaseURL
	err := uri.Set("http://example.com:1234")
	assert.NoError(t, err)
	assert.Equal(t, BaseURL{"http://", NetAddress{"example.com", 1234}}, uri)
	assert.Equal(t, "http://example.com:1234", uri.String())

	err = uri.Set("ftp://example.com:21")
	assert.Error(t, err)
}

func TestParseEnvVariables(t *testing.T) {
	assert.NoError(t, os.Setenv("APP_ENV", "envtest"))
	assert.NoError(t, os.Setenv("JWT_SECRET", "envsecret"))
	assert.NoError(t, os.Setenv("SERVER_ADDRESS", "127.0.0.1:9000"))
	assert.NoError(t, os.Setenv("BASE_URL", "https://env.com:443"))
	assert.NoError(t, os.Setenv("FILE_STORAGE_PATH", "/tmp/testdata"))
	assert.NoError(t, os.Setenv("DATABASE_DSN", "postgresql://user:password@localhost:5432/dbname?sslmode=disable"))
	defer func() {
		assert.NoError(t, os.Unsetenv("APP_ENV"))
		assert.NoError(t, os.Unsetenv("JWT_SECRET"))
		assert.NoError(t, os.Unsetenv("SERVER_ADDRESS"))
		assert.NoError(t, os.Unsetenv("BASE_URL"))
		assert.NoError(t, os.Unsetenv("FILE_STORAGE_PATH"))
		assert.NoError(t, os.Unsetenv("DATABASE_DSN"))
	}()

	osArgs := os.Args
	defer func() { os.Args = osArgs }()
	os.Args = os.Args[:1]

	cfg, err := Parse()
	assert.NoError(t, err)
	assert.Equal(t, "envtest", cfg.AppEnv)
	assert.Equal(t, "envsecret", cfg.JWTSecret)
	assert.Equal(t, NetAddress{"127.0.0.1", 9000}, cfg.ServerAddress)
	assert.Equal(t, BaseURL{"https://", NetAddress{"env.com", 443}}, cfg.BaseURL)
	assert.Equal(t, "/tmp/testdata", cfg.FileStoragePath)
	assert.Equal(t, "postgresql://user:password@localhost:5432/dbname?sslmode=disable", cfg.DatabaseDSN)
}
