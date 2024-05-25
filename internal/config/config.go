package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

type AppConfig struct {
	ServiceName    string `yaml:"service_name"`
	FirebaseConfig struct {
		CredentialsFile string `yaml:"credentials_file"`
		Sms             string `yaml:"collection_sms"`
		UserAccount     string `yaml:"collection_user_account"`
		Phone           string `yaml:"collection_phone"`
	} `yaml:"firebase"`
	DryRun                string `yaml:"dry_run"`
	MongoConnectionString string `yaml:"mongo_connection_string"`
	MongoDatabaseName     string `yaml:"mongo_database_name"`
}

func LoadConfig(configPath string) AppConfig {
	f, err := os.Open(configPath)
	if err != nil {
		fmt.Println("No config file found")
	}
	defer f.Close()

	var cfg AppConfig
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println("No config file found", err)
	}
	err = envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
