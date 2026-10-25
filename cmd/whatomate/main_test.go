package main

import (
	"testing"

	"github.com/shridarpatil/whatomate/internal/config"
)

func TestValidateProductionSecrets(t *testing.T) {
	valid := config.Config{
		App:          config.AppConfig{EncryptionKey: "0123456789abcdef0123456789abcdef"},
		JWT:          config.JWTConfig{Secret: "fedcba9876543210fedcba9876543210"},
		DefaultAdmin: config.DefaultAdminConfig{Password: "unique-admin-password-123"},
	}
	if err := validateProductionSecrets(&valid); err != nil {
		t.Fatalf("valid production configuration rejected: %v", err)
	}

	for _, tc := range []struct {
		name   string
		modify func(*config.Config)
	}{
		{"example JWT secret", func(c *config.Config) { c.JWT.Secret = "your-super-secret-jwt-key-change-in-production" }},
		{"short JWT secret", func(c *config.Config) { c.JWT.Secret = "short" }},
		{"missing encryption key", func(c *config.Config) { c.App.EncryptionKey = "" }},
		{"default admin password", func(c *config.Config) { c.DefaultAdmin.Password = "admin" }},
		{"short admin password", func(c *config.Config) { c.DefaultAdmin.Password = "short" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			tc.modify(&cfg)
			if err := validateProductionSecrets(&cfg); err == nil {
				t.Fatal("insecure production configuration accepted")
			}
		})
	}
}
