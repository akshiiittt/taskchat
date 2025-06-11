package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return Config{Port: port}
}

func ValidateEnvVars() error {
	vars := []string{"DATABASE_URL", "JWT_SECRET", "CORS_ALLOW_ORIGINS"}
	for _, v := range vars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("%v is is not there", v)
		}
	}
	return nil
}
