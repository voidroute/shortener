package config

import (
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

type ServerConfig struct {
	Addr            string        `mapstructure:"addr"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type AppConfig struct {
	BaseURL   string `mapstructure:"base_url"`
	Env       string `mapstructure:"env"`
	GeoIPPath string `mapstructure:"geoip_path"`
}

type KafkaConfig struct {
	Addr  string `mapstructure:"addr"`
	Topic string `mapstructure:"topic"`
}

type DatabaseConfig struct {
	UsePostgres     bool          `mapstructure:"use_postgres"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

type CacheConfig struct {
	UseRedis        bool          `mapstructure:"use_redis"`
	RedisAddr       string        `mapstructure:"redis_addr"`
	TTL             time.Duration `mapstructure:"ttl"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

var (
	instance *Config
	once     sync.Once
)

func Init() (*Config, error) {
	var initErr error

	once.Do(func() {

		setDefaults()

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		bindEnvVars()

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok {
				slog.Warn("Config file not found, using defaults")
			} else {
				initErr = err
				return
			}
		}

		cfg := &Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			initErr = err
			return
		}

		instance = cfg

		slog.Info("Config loaded",
			"env", cfg.App.Env,
			"use_postgres", cfg.Database.UsePostgres,
			"use_redis", cfg.Cache.UseRedis,
		)
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

func setDefaults() {
	viper.SetDefault("server.addr", "0.0.0.0:8080")

	viper.SetDefault("app.base_url", "http://localhost:8080")
	viper.SetDefault("app.env", "dev")
	viper.SetDefault("app.geoip_path", "GeoLite2-Country.mmdb")

	viper.SetDefault("kafka.addr", "kafka:29092")
	viper.SetDefault("kafka.topic", "click_events")

	viper.SetDefault("database.use_postgres", false)
	viper.SetDefault("database.host", "postgres")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "url_shortener")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")
	viper.SetDefault("database.conn_max_idle_time", "1m")

	viper.SetDefault("cache.use_redis", false)
	viper.SetDefault("cache.redis_addr", "redis:6379")
	viper.SetDefault("cache.ttl", "24h")
	viper.SetDefault("cache.cleanup_interval", "1m")
}

func bindEnvVars() {
	_ = viper.BindEnv("database.host", "POSTGRES_HOST")
	_ = viper.BindEnv("database.port", "POSTGRES_PORT")
	_ = viper.BindEnv("database.user", "POSTGRES_USER")
	_ = viper.BindEnv("database.password", "POSTGRES_PASSWORD")
	_ = viper.BindEnv("database.dbname", "POSTGRES_DB")
	_ = viper.BindEnv("database.ssl_mode", "POSTGRES_SSL_MODE")
	_ = viper.BindEnv("app.env", "APP_ENV")

	_ = viper.BindEnv("cache.use_redis", "CACHE_USE_REDIS")
	_ = viper.BindEnv("cache.redis_addr", "REDIS_ADDR")

	_ = viper.BindEnv("kafka.addr", "KAFKA_ADDR")
	_ = viper.BindEnv("kafka.topic", "KAFKA_TOPIC")
}
