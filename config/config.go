package config

import (
	"log/slog"
	"strings"

	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type KVstring map[string]string

func (c *KVstring) Decode(value string) error {
	if value == "" {
		*c = make(map[string]string)
		return nil
	}
	contracts := make(map[string]string)
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kvpair := strings.Split(pair, "=")
		if len(kvpair) != 2 {
			return errors.Errorf("invalid map item: %q", pair)
		}
		contracts[kvpair[0]] = kvpair[1]

	}
	*c = KVstring(contracts)
	return nil
}

type Config struct {
	Log        Log        `envconfig:"LOG"`
	HTTPServer HTTPServer `envconfig:"HTTP_SERVER"`

	ExternalHost string `envconfig:"EXTERNAL_HOST" required:"true"`

	SupportedStateContracts KVstring `envconfig:"SUPPORTED_STATE_CONTRACTS" required:"true"`
	SupportedRPC            KVstring `envconfig:"SUPPORTED_RPC" required:"true"`

	MongoDBConnectionString string `envconfig:"MONGODB_CONNECTION_STRING" default:"mongodb://localhost:27017/credentials"`

	Issuers []string `envconfig:"ISSUERS" required:"true"`

	KeysDirPath string `envconfig:"KEYS_DIR_PATH" default:"./keys"`
}

type Log struct {
	Level       string `envconfig:"LEVEL" default:"INFO"`
	Environment string `envconfig:"ENVIRONMENT" default:"production"`
}

func (l *Log) LogLevel() (loglevel slog.Level) {
	switch l.Level {
	case "DEBUG":
		loglevel = slog.LevelDebug
	case "INFO":
		loglevel = slog.LevelInfo
	case "WARN":
		loglevel = slog.LevelWarn
	case "ERROR":
		loglevel = slog.LevelError
	case "FATAL":
		loglevel = logger.LevelFatal
	case "NOTICE":
		loglevel = logger.LevelNotice
	}
	return
}

type HTTPServer struct {
	Host    string   `envconfig:"HOST"`
	Port    string   `envconfig:"PORT" default:"8080"`
	Origins []string `envconfig:"ORIGINS" default:"*"`
}

func Parse() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
