package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

type AppConfig struct {
	FirebaseConfig struct {
		CredentialsFile string `yaml:"credentials_file"`
		Sms             string `yaml:"collection_sms"`
		UserAccount     string `yaml:"collection_user_account"`
		Phone           string `yaml:"collection_phone"`
	} `yaml:"firebase"`
}

func LoadConfig(configPath string) AppConfig {
	f, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg AppConfig
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}
	err = envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
