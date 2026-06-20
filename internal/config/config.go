package config

import (
	"strings"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

type AppConfig struct {
	AppName  string         `koanf:"app_name"`
	Postgres PostgresConfig `koanf:"postgres"`
	RabbitMQ RabbitMQConfig `koanf:"rabbitmq"`
	Firebase FirebaseConfig `koanf:"firebase"`
	DryRun   bool           `koanf:"dry_run"`
}

type PostgresConfig struct {
	DSN string `koanf:"dsn"`
}

type RabbitMQConfig struct {
	DSN string `koanf:"dsn"`
}

type FirebaseConfig struct {
	CredentialsFile string `koanf:"credentials_file"`
}

const ENV_PREFIX = "ENV_"

func LoadConfig() (AppConfig, error) {
	k := koanf.New(".")

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: ENV_PREFIX,
		TransformFunc: trasnformFunction,
	}), nil); err != nil {
		return AppConfig{}, err
	}

	var cfg AppConfig
	if err := k.Unmarshal("", &cfg); err != nil {
		return AppConfig{}, err
	}

	return cfg, nil
}

func trasnformFunction(k, v string) (string, any) {
	// convert to lowercase and replace underscores with dots, except for double underscores which are replaced with a single underscore
	k = StripUnderscore(strings.ToLower(strings.TrimPrefix(k, ENV_PREFIX)), ".")
	// split for array, like "ENV_MY_ARRAY=val1 val2 val3" into []string{"val1", "val2", "val3"}
	if strings.Contains(v, " ") {
		return k, strings.Split(v, " ")
	}
	return k, v
}