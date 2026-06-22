package app

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const DefaultConfigPath = "configs/config.yaml"

// Config holds application configuration loaded from YAML with ENV overrides.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

type ServerConfig struct {
	Host            string        `yaml:"host" env:"SERVER_HOST" env-default:"0.0.0.0"`
	Port            int           `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT" env-default:"15s"`
}

type DatabaseConfig struct {
	DSN             string        `yaml:"dsn" env:"DATABASE_DSN"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DATABASE_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DATABASE_CONN_MAX_LIFETIME" env-default:"5m"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

type JWTConfig struct {
	Secret    string        `yaml:"secret" env:"JWT_SECRET"`
	AccessTTL time.Duration `yaml:"access_ttl" env:"JWT_ACCESS_TTL" env-default:"24h"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute" env:"RATE_LIMIT_REQUESTS_PER_MINUTE" env-default:"100"`
}

// Load reads configuration from CONFIG_PATH or DefaultConfigPath.
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = DefaultConfigPath
	}
	return loadFrom(path)
}

func loadFrom(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	return &cfg, nil
}

// Addr returns the HTTP listen address for the server.
func (c *Config) Addr() string {
	return net.JoinHostPort(c.Server.Host, strconv.Itoa(c.Server.Port))
}
