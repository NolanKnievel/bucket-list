package database

import (
	"os"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Set test environment variables
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSL_MODE", "require")

	config := LoadConfigFromEnv()

	if config.Host != "testhost" {
		t.Errorf("Expected Host to be 'testhost', got '%s'", config.Host)
	}
	if config.Port != "5433" {
		t.Errorf("Expected Port to be '5433', got '%s'", config.Port)
	}
	if config.User != "testuser" {
		t.Errorf("Expected User to be 'testuser', got '%s'", config.User)
	}
	if config.Password != "testpass" {
		t.Errorf("Expected Password to be 'testpass', got '%s'", config.Password)
	}
	if config.DBName != "testdb" {
		t.Errorf("Expected DBName to be 'testdb', got '%s'", config.DBName)
	}
	if config.SSLMode != "require" {
		t.Errorf("Expected SSLMode to be 'require', got '%s'", config.SSLMode)
	}

	// Clean up
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSL_MODE")
}

func TestLoadConfigFromEnvDefaults(t *testing.T) {
	// Ensure no environment variables are set
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSL_MODE")

	config := LoadConfigFromEnv()

	if config.Host != "localhost" {
		t.Errorf("Expected default Host to be 'localhost', got '%s'", config.Host)
	}
	if config.Port != "5432" {
		t.Errorf("Expected default Port to be '5432', got '%s'", config.Port)
	}
	if config.User != "postgres" {
		t.Errorf("Expected default User to be 'postgres', got '%s'", config.User)
	}
	if config.Password != "" {
		t.Errorf("Expected default Password to be empty, got '%s'", config.Password)
	}
	if config.DBName != "collaborative_bucket_list" {
		t.Errorf("Expected default DBName to be 'collaborative_bucket_list', got '%s'", config.DBName)
	}
	if config.SSLMode != "disable" {
		t.Errorf("Expected default SSLMode to be 'disable', got '%s'", config.SSLMode)
	}
}