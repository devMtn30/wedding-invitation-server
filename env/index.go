package env

import "os"

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

var (
	AdminPassword = getEnv("ADMIN_PASSWORD", "asdf")
	AllowOrigin   = getEnv("ALLOW_ORIGIN", "https://devmtn30.github.io")
	GCPProjectID  = os.Getenv("GCP_PROJECT_ID")
)
