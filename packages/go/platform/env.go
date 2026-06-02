package platform

import (
	"fmt"
	"strconv"
	"strings"
)

type LookupFunc func(string) (string, bool)

type Env struct {
	lookup LookupFunc
}

func NewEnv(lookup LookupFunc) Env {
	if lookup == nil {
		lookup = func(string) (string, bool) {
			return "", false
		}
	}

	return Env{lookup: lookup}
}

func (env Env) String(key string, fallback string) string {
	value, ok := env.lookup(key)
	if !ok || value == "" {
		return fallback
	}

	return value
}

func (env Env) Int(key string, fallback int) (int, error) {
	value, ok := env.lookup(key)
	if !ok || value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return parsed, nil
}

func (env Env) CSV(key string) []string {
	value, ok := env.lookup(key)
	if !ok || value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}

	return out
}
