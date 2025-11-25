package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 应用配置
type Config struct {
	Nebula NebulaConfig `mapstructure:"nebula"`
	Milvus MilvusConfig `mapstructure:"milvus"`
	Logger LoggerConfig `mapstructure:"logger"`
}

// NebulaConfig Nebula Graph 配置
type NebulaConfig struct {
	Addresses []string      `mapstructure:"addresses"`
	Username  string        `mapstructure:"username"`
	Password  string        `mapstructure:"password"`
	Space     string        `mapstructure:"space"`
	Timeout   time.Duration `mapstructure:"timeout"`
	MaxConn   int           `mapstructure:"max_conn"`
	MinConn   int           `mapstructure:"min_conn"`
}

// MilvusConfig Milvus 配置
type MilvusConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Username       string        `mapstructure:"username"`
	Password       string        `mapstructure:"password"`
	Database       string        `mapstructure:"database"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	EnableTLS      bool          `mapstructure:"enable_tls"`
	TLSCert        string        `mapstructure:"tls_cert"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	OutputPath string `mapstructure:"output_path"` // 日志输出路径
	Encoding   string `mapstructure:"encoding"`    // json, console
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 支持环境变量覆盖
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// Nebula 默认配置
	viper.SetDefault("nebula.addresses", []string{"127.0.0.1:9669"})
	viper.SetDefault("nebula.username", "root")
	viper.SetDefault("nebula.password", "password")
	viper.SetDefault("nebula.space", "brain")
	viper.SetDefault("nebula.timeout", "10s")
	viper.SetDefault("nebula.max_conn", 10)
	viper.SetDefault("nebula.min_conn", 2)

	// Milvus 默认配置
	viper.SetDefault("milvus.host", "127.0.0.1")
	viper.SetDefault("milvus.port", 19530)
	viper.SetDefault("milvus.username", "")
	viper.SetDefault("milvus.password", "")
	viper.SetDefault("milvus.database", "default")
	viper.SetDefault("milvus.connect_timeout", "10s")
	viper.SetDefault("milvus.enable_tls", false)

	// Logger 默认配置
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.output_path", "stdout")
	viper.SetDefault("logger.encoding", "json")
}

// NewLogger 根据配置创建 Logger
func NewLogger(cfg LoggerConfig) (*zap.Logger, error) {
	var config zap.Config

	if cfg.Encoding == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	switch cfg.Level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// 设置输出路径
	if cfg.OutputPath != "stdout" && cfg.OutputPath != "" {
		config.OutputPaths = []string{cfg.OutputPath}
		config.ErrorOutputPaths = []string{cfg.OutputPath}
	}

	return config.Build()
}

