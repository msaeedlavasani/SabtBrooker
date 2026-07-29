package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Redis    RedisConfig
	NATS     NATSConfig
	JWT      JWTConfig
	OTP      OTPConfig
	Security SecurityConfig
	MinIO    MinIOConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MinConns int
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type NATSConfig struct {
	URL       string
	Stream    string
	MaxReconn int
}

type JWTConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	Issuer         string
}

type OTPConfig struct {
	Length     int
	TTL        time.Duration
	MaxAttempts int
	RateLimit  int
	RateWindow time.Duration
}

type SecurityConfig struct {
	MaxFailedLogin      int
	LockDuration        time.Duration
	BCryptCost          int
	EncryptionKey       string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "sabtbrooker"),
			Password: getEnv("DB_PASSWORD", "sabtbrooker"),
			DBName:   getEnv("DB_NAME", "sabtbrooker"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxConns: getInt("DB_MAX_CONNS", 25),
			MinConns: getInt("DB_MIN_CONNS", 5),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
		NATS: NATSConfig{
			URL:       getEnv("NATS_URL", "nats://localhost:4222"),
			Stream:    getEnv("NATS_STREAM", "sabtbrooker"),
			MaxReconn: getInt("NATS_MAX_RECONN", 10),
		},
		JWT: JWTConfig{
			PrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
			PublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),
			AccessTTL:      getDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:     getDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
			Issuer:         getEnv("JWT_ISSUER", "sabtbrooker"),
		},
		OTP: OTPConfig{
			Length:      getInt("OTP_LENGTH", 5),
			TTL:         getDuration("OTP_TTL", 2*time.Minute),
			MaxAttempts: getInt("OTP_MAX_ATTEMPTS", 3),
			RateLimit:   getInt("OTP_RATE_LIMIT", 5),
			RateWindow:  getDuration("OTP_RATE_WINDOW", 10*time.Minute),
		},
		Security: SecurityConfig{
			MaxFailedLogin: getInt("SECURITY_MAX_FAILED_LOGIN", 5),
			LockDuration:   getDuration("SECURITY_LOCK_DURATION", 30*time.Minute),
			BCryptCost:     getInt("SECURITY_BCRYPT_COST", 12),
			EncryptionKey:  getEnv("SECURITY_ENCRYPTION_KEY", ""),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "sabtbrooker"),
			UseSSL:    getBool("MINIO_USE_SSL", false),
		},
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return defaultVal
}

func getBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
	}
	return defaultVal
}
