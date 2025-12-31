// Package config provides configuration loading and management for the application.
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		envVars     map[string]string
		expectError bool
		validate    func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid config",
			configYAML: `
server:
  addr: ":8080"
  mode: "debug"
  read_timeout: 30
  write_timeout: 30
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "password"
  dbname: "test_db"
redis:
  addr: "localhost:6379"
jwt:
  secret: "this-is-a-very-long-secret-key-for-testing"
  access_token_ttl: 60
  refresh_token_ttl: 168
  issuer: "test"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, ":8080", cfg.Server.Addr)
				assert.Equal(t, "localhost", cfg.Database.Host)
				assert.Equal(t, 3306, cfg.Database.Port)
			},
		},
		{
			name: "config with env override",
			configYAML: `
server:
  addr: ":8080"
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "password"
  dbname: "test_db"
jwt:
  secret: "this-is-a-very-long-secret-key-for-testing"
`,
			envVars: map[string]string{
				"VC_SERVER_ADDR": ":9090",
				"VC_DB_HOST":     "db.example.com",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, ":9090", cfg.Server.Addr)
				assert.Equal(t, "db.example.com", cfg.Database.Host)
			},
		},
		{
			name: "missing required field",
			configYAML: `
server:
  addr: ""
database:
  host: "localhost"
  dbname: "test_db"
jwt:
  secret: "this-is-a-very-long-secret-key-for-testing"
`,
			expectError: true,
		},
		{
			name: "jwt secret too short",
			configYAML: `
server:
  addr: ":8080"
database:
  host: "localhost"
  dbname: "test_db"
jwt:
  secret: "short"
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tmpFile, err := os.CreateTemp("", "config-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.configYAML)
			require.NoError(t, err)
			require.NoError(t, tmpFile.Close())

			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Load config
			cfg, err := Load(tmpFile.Name())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
}

func TestLoadConfigEmptyPath(t *testing.T) {
	_, err := Load("")
	assert.Error(t, err)
}

func TestDatabaseDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		DBName:   "testdb",
	}

	expected := "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	assert.Equal(t, expected, cfg.DSN())
}
