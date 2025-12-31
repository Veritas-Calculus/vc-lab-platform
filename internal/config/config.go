// Package config provides configuration loading and management for the application.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	SSO      SSOConfig      `yaml:"sso"`
}

// ServerConfig represents HTTP server configuration.
type ServerConfig struct {
	Addr         string `yaml:"addr"`
	Mode         string `yaml:"mode"` // debug, release, test
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"dbname"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // in minutes
}

// RedisConfig represents Redis configuration.
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// JWTConfig represents JWT configuration.
type JWTConfig struct {
	Secret          string `yaml:"secret"`
	AccessTokenTTL  int    `yaml:"access_token_ttl"`  // in minutes
	RefreshTokenTTL int    `yaml:"refresh_token_ttl"` // in hours
	Issuer          string `yaml:"issuer"`
}

// SSOConfig represents SSO configuration.
type SSOConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ProviderURL  string `yaml:"provider_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
}

// Load loads configuration from the specified file path.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	cfg.applyEnvOverrides()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration.
func (c *Config) applyEnvOverrides() {
	if addr := os.Getenv("VC_SERVER_ADDR"); addr != "" {
		c.Server.Addr = addr
	}
	if dbHost := os.Getenv("VC_DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if dbPass := os.Getenv("VC_DB_PASSWORD"); dbPass != "" {
		c.Database.Password = dbPass
	}
	if redisAddr := os.Getenv("VC_REDIS_ADDR"); redisAddr != "" {
		c.Redis.Addr = redisAddr
	}
	if jwtSecret := os.Getenv("VC_JWT_SECRET"); jwtSecret != "" {
		c.JWT.Secret = jwtSecret
	}
}

// validate validates the configuration.
func (c *Config) validate() error {
	var errs []string

	if c.Server.Addr == "" {
		errs = append(errs, "server.addr is required")
	}
	if c.Database.Host == "" {
		errs = append(errs, "database.host is required")
	}
	if c.Database.DBName == "" {
		errs = append(errs, "database.dbname is required")
	}
	if c.JWT.Secret == "" {
		errs = append(errs, "jwt.secret is required")
	}
	if len(c.JWT.Secret) < 32 {
		errs = append(errs, "jwt.secret must be at least 32 characters")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// DSN returns the database connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}
