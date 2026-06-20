package config

import "testing"

func TestLoadConfigFromEnvironment(t *testing.T) {
	t.Setenv("ENV_APP__NAME", "sms-gateway-test")
	t.Setenv("ENV_POSTGRES_DSN", "postgres://user:pass@postgres:5432/sms_gateway?sslmode=disable")
	t.Setenv("ENV_RABBITMQ_DSN", "amqp://user:pass@rabbitmq:5672/")
	t.Setenv("ENV_FIREBASE_CREDENTIALS__FILE", "/etc/sms-gateway/firebase.json")
	t.Setenv("ENV_DRY__RUN", "true")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.AppName != "sms-gateway-test" {
		t.Fatalf("AppName = %q", cfg.AppName)
	}
	if cfg.Postgres.DSN != "postgres://user:pass@postgres:5432/sms_gateway?sslmode=disable" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
	if cfg.RabbitMQ.DSN != "amqp://user:pass@rabbitmq:5672/" {
		t.Fatalf("RabbitMQ.DSN = %q", cfg.RabbitMQ.DSN)
	}
	if cfg.Firebase.CredentialsFile != "/etc/sms-gateway/firebase.json" {
		t.Fatalf("Firebase.CredentialsFile = %q", cfg.Firebase.CredentialsFile)
	}
	if !cfg.DryRun {
		t.Fatal("DryRun = false")
	}
}

func TestTransformFunctionPreservesDoubleUnderscoreAsSingleUnderscore(t *testing.T) {
	key, value := trasnformFunction("FIREBASE_CREDENTIALS__FILE", "/tmp/firebase.json")

	if key != "firebase.credentials_file" {
		t.Fatalf("key = %q", key)
	}
	if value != "/tmp/firebase.json" {
		t.Fatalf("value = %v", value)
	}
}
