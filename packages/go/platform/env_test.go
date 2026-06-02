package platform

import (
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/testkit"
)

func TestEnvReadsTypedValues(t *testing.T) {
	env := NewEnv(func(key string) (string, bool) {
		values := map[string]string{
			"SERVICE_PORT":    "8080",
			"ALLOWED_ORIGINS": "https://app.example.com, https://admin.example.com",
		}

		value, ok := values[key]
		return value, ok
	})

	port, err := env.Int("SERVICE_PORT", 80)
	testkit.NoError(t, err)
	testkit.Equal(t, port, 8080)
	testkit.Equal(t, env.String("SERVICE_NAME", "api"), "api")
	testkit.Equal(t, len(env.CSV("ALLOWED_ORIGINS")), 2)
}
