package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Relay      RelayConfig
	Centrifugo CentrifugoConfig
	Stripe     StripeConfig
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port string
}

// RelayConfig holds relay/tunnel settings for both embedded client and relay server binary.
type RelayConfig struct {
	DeploymentMode string
	Token          string
	ServerURL      string
	LocalAPIURL    string
	Domain         string
	ListenAddr     string
	ConfigFile     string
	Subdomain      string
}

// CentrifugoConfig holds real-time engine integration settings.
type CentrifugoConfig struct {
	Enabled         bool
	APIURL          string
	APIKey          string
	TokenHMACSecret string
	ChannelPrefix   string
}

// StripeConfig holds Stripe billing settings.
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	PriceIDPro    string
	PriceIDTeam   string
}

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT authentication settings
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB value: %w", err)
	}

	// Parse JWT expiry duration
	jwtExpiryStr := getEnv("JWT_EXPIRY", "24h")
	jwtExpiry, err := time.ParseDuration(jwtExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY value: %w", err)
	}

	// JWT secret is required in production
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	serverPort := getEnv("SERVER_PORT", "8080")
	relayConfigFile := getEnv("RELAY_CONFIG_FILE", "data/relay_config.json")
	deploymentMode := getEnv("DEPLOYMENT_MODE", "self-hosted")
	relayToken := getEnv("RELAY_TOKEN", "")
	relaySubdomain := ""
	if relayToken == "" {
		if persisted, err := loadPersistedRelayConfig(relayConfigFile); err == nil {
			if persisted.RelayToken != "" {
				relayToken = persisted.RelayToken
			}
			if deploymentMode == "self-hosted" && persisted.DeploymentMode != "" {
				deploymentMode = persisted.DeploymentMode
			}
			relaySubdomain = persisted.Subdomain
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: serverPort,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "agentunited"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
			Expiry: jwtExpiry,
		},
		Relay: RelayConfig{
			DeploymentMode: deploymentMode,
			Token:          relayToken,
			ServerURL:      getEnv("RELAY_SERVER", "ws://localhost:8090/tunnel"),
			LocalAPIURL:    getEnv("RELAY_LOCAL_API", fmt.Sprintf("http://127.0.0.1:%s", serverPort)),
			Domain:         getEnv("RELAY_DOMAIN", "tunnel.agentunited.ai"),
			ListenAddr:     getEnv("RELAY_LISTEN_ADDR", ":8090"),
			ConfigFile:     relayConfigFile,
			Subdomain:      relaySubdomain,
		},
		Centrifugo: CentrifugoConfig{
			Enabled:         getEnvBool("CENTRIFUGO_ENABLED", false),
			APIURL:          getEnv("CENTRIFUGO_API_URL", "http://localhost:8000/api"),
			APIKey:          getEnv("CENTRIFUGO_API_KEY", ""),
			TokenHMACSecret: getEnv("CENTRIFUGO_TOKEN_HMAC_SECRET", ""),
			ChannelPrefix:   getEnv("CENTRIFUGO_CHANNEL_PREFIX", "channel:"),
		},
		Stripe: StripeConfig{
			SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
			PriceIDPro:    getEnv("STRIPE_PRICE_ID_PRO", ""),
			PriceIDTeam:   getEnv("STRIPE_PRICE_ID_TEAM", ""),
		},
	}

	return cfg, nil
}

// DSN returns PostgreSQL connection string
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode,
	)
}

// RedisAddr returns Redis connection address
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// getEnv reads environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultValue
	}
	return b
}

type persistedRelayConfig struct {
	DeploymentMode string `json:"deployment_mode"`
	RelayToken     string `json:"relay_token"`
	Subdomain      string `json:"subdomain,omitempty"`
}

func loadPersistedRelayConfig(path string) (*persistedRelayConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out persistedRelayConfig
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
