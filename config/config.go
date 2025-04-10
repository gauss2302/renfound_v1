package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the application configurations
type Config struct {
	DB       PostgresConfig `mapstructure:"postgres"`
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Telegram TelegramConfig `mapstructure:"telegram"` // Add this line
}

// TelegramConfig holds Telegram configurations
type TelegramConfig struct {
	BotToken string `mapstructure:"bottoken"`
}

// PostgresConfig holds PostgreSQL configurations
type PostgresConfig struct {
	URL string `mapstructure:"url"`
}

// ServerConfig holds server configurations
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// JWTConfig holds JWT configurations
type JWTConfig struct {
	AccessSecret  string        `mapstructure:"accesssecret"`
	RefreshSecret string        `mapstructure:"refreshsecret"`
	AccessTTL     time.Duration `mapstructure:"accessttl"`
	RefreshTTL    time.Duration `mapstructure:"refreshttl"`
}

// RedisConfig holds Redis configurations
type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LoggerConfig holds Zap logger configurations
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"outputpath"`
}

// AppConfig holds the application configuration and logger instance
type AppConfig struct {
	Config *Config
	Logger *zap.Logger
}

// LoadConfig initializes and returns the application configuration with logger
func LoadConfig() (*AppConfig, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Continue without .env file, just log a warning later with the proper logger
	}

	// Viper configuration
	viper.AutomaticEnv() // Automatically read environment variables

	// Replace dots with underscores in environment variables
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Explicitly bind environment variables to configuration keys
	bindEnvs()

	// Set default values
	setDefaults()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(config.Logger)
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %v", err)
	}

	// Log that .env was not found if that was the case
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found", zap.Error(err))
	}

	// Log successful configuration loading
	logger.Info("Configuration loaded successfully",
		zap.String("db_schema", "postgres://****:****@"+strings.Split(config.DB.URL, "@")[1]),
		zap.String("server_host", config.Server.Host),
		zap.String("server_port", config.Server.Port),
	)

	return &AppConfig{
		Config: &config,
		Logger: logger,
	}, nil
}

// initLogger creates and configures a new Zap logger
func initLogger(cfg LoggerConfig) (*zap.Logger, error) {
	// Convert log level string to zapcore.Level
	level := getLogLevel(cfg.Level)

	// Default to JSON in production, console in development
	encoding := cfg.Encoding
	if encoding == "" {
		encoding = "json"
	}

	// Configure logger
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Encoding:         encoding,
		OutputPaths:      []string{getOutputPath(cfg.OutputPath)},
		ErrorOutputPaths: []string{getOutputPath(cfg.OutputPath)},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	return config.Build()
}

// getLogLevel converts string log level to zapcore.Level
func getLogLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // Default to info level
	}
}

// getOutputPath returns the appropriate output path
func getOutputPath(path string) string {
	if path == "" {
		return "stdout"
	}
	return path
}

// bindEnvs binds each configuration key to its corresponding environment variable
func bindEnvs() {
	envBindings := map[string]string{
		"postgres.url":      "DATABASE_URL",
		"server.port":       "APP_SERVER_PORT",
		"server.host":       "APP_SERVER_HOST",
		"jwt.accessSecret":  "APP_JWT_ACCESSSECRET",
		"jwt.refreshSecret": "APP_JWT_REFRESHSECRET",
		"jwt.accessTTL":     "APP_JWT_ACCESSTTL",
		"jwt.refreshTTL":    "APP_JWT_REFRESHTTL",
		"redis.url":         "REDIS_URL",
		"redis.password":    "REDIS_PASSWORD",
		"redis.db":          "REDIS_DB",
		"logger.level":      "APP_LOGGER_LEVEL",
		"logger.encoding":   "APP_LOGGER_ENCODING",
		"logger.outputpath": "APP_LOGGER_OUTPUTPATH",
		"telegram.bottoken": "TELEGRAM_BOT_TOKEN",
	}

	for configKey, envVar := range envBindings {
		if err := viper.BindEnv(configKey, envVar); err != nil {
			// We can't use zap logger yet as it's not initialized
			panic(fmt.Sprintf("Error binding %s: %v", configKey, err))
		}
	}
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8090")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "json")
	viper.SetDefault("logger.outputpath", "stdout")
}
