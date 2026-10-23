package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/quockhanhcao/simple_store/internal/platform"
)

type Config struct {
	Env         string
	HTTPAddr    string
	DatabaseURL string
	LogFormat   string
	LogLevel    slog.Level
}

func Load(getenv func(string) string, defaultAddr string) (*Config, error) {
	cfg := Config{}

	errs := []error{}
	dbURL := getenv("DATABASE_URL")
	if dbURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	httpAddr := getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = defaultAddr
	}

	var logFormat string
	parsedFormat := strings.ToLower(getenv("LOG_FORMAT"))
	switch parsedFormat {
	case "":
		logFormat = "text"
	case "json":
		logFormat = "json"
	case "text":
		logFormat = "text"
	default:
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be \"text\" or \"json\", got %q", getenv("LOG_FORMAT")))
	}

	logLevel := getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	parsedLogLvl, err := platform.ParseLogLevel(logLevel)
	if err != nil {
		errs = append(errs, err)
	}
	var env string
	parsedEnv := strings.ToLower(getenv("APP_ENV"))
	switch parsedEnv {
	case "":
		env = "Production"
	case "production":
		env = "Production"
	case "development":
		env = "Development"
	case "staging":
		env = "Staging"
	default:
		// Deliberately stricter than ASP.NET (which accepts any name) to catch typos.
		errs = append(errs, fmt.Errorf("APP_ENV must be \"Production\", \"Development\" or \"Staging\", got %q", getenv("APP_ENV")))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	cfg.HTTPAddr = httpAddr
	cfg.DatabaseURL = dbURL
	cfg.LogFormat = logFormat
	cfg.LogLevel = parsedLogLvl
	cfg.Env = env
	return &cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.Env, "Development")
}
