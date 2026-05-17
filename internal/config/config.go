package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type NATSConfig struct {
	URL             string `mapstructure:"url"`
	TaskSubject     string `mapstructure:"task_subject"`
	ResultSubject   string `mapstructure:"result_subject"`
	CancelSubject   string `mapstructure:"cancel_subject"`
	HeartbeatSubject string `mapstructure:"heartbeat_subject"`
}

type WorkerManagerConfig struct {
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	OfflineThreshold  time.Duration `mapstructure:"offline_threshold"`
	MaxTaskDuration   time.Duration `mapstructure:"max_task_duration"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// AppConfig is the top-level config for the API server.
type AppConfig struct {
	Server  ServerConfig       `mapstructure:"server"`
	Database DatabaseConfig     `mapstructure:"database"`
	NATS    NATSConfig         `mapstructure:"nats"`
	Worker  WorkerManagerConfig `mapstructure:"worker"`
	Log     LogConfig          `mapstructure:"log"`
}

// WorkerAppConfig is the top-level config for a worker node.
type WorkerAppConfig struct {
	WorkerID   string        `mapstructure:"worker_id"`
	WorkerName string        `mapstructure:"worker_name"`
	Nuclei     NucleiConfig  `mapstructure:"nuclei"`
	NATS       NATSConfig    `mapstructure:"nats"`
	Heartbeat  time.Duration `mapstructure:"heartbeat_interval"`
	Log        LogConfig     `mapstructure:"log"`
}

type NucleiConfig struct {
	TemplatesDir    string   `mapstructure:"templates_dir"`
	ExcludeTags     []string `mapstructure:"exclude_tags"`
	Concurrency     int      `mapstructure:"concurrency"`
	RateLimit       int      `mapstructure:"rate_limit"`
}

func LoadServerConfig(path string) (*AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvPrefix("NUCLEI_PLATFORM")

	// defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "nuclei-platform.db")
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.task_subject", "nuclei.task.create")
	v.SetDefault("nats.result_subject", "nuclei.task.result")
	v.SetDefault("nats.cancel_subject", "nuclei.task.cancel")
	v.SetDefault("nats.heartbeat_subject", "nuclei.worker.heartbeat")
	v.SetDefault("worker.heartbeat_interval", "15s")
	v.SetDefault("worker.offline_threshold", "45s")
	v.SetDefault("worker.max_task_duration", "1h")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

func LoadWorkerConfig(path string) (*WorkerAppConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvPrefix("NUCLEI_PLATFORM")

	// defaults
	v.SetDefault("worker_id", "")
	v.SetDefault("worker_name", "")
	v.SetDefault("nuclei.templates_dir", "/root/nuclei-templates")
	v.SetDefault("nuclei.exclude_tags", []string{"dos"})
	v.SetDefault("nuclei.concurrency", 25)
	v.SetDefault("nuclei.rate_limit", 0)
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.task_subject", "nuclei.task.create")
	v.SetDefault("nats.result_subject", "nuclei.task.result")
	v.SetDefault("nats.cancel_subject", "nuclei.task.cancel")
	v.SetDefault("nats.heartbeat_subject", "nuclei.worker.heartbeat")
	v.SetDefault("heartbeat_interval", "15s")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg WorkerAppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
