package config

import (
	"strings"
)

type Reporter interface {
	Get(key string) (string, error)
	GetWithDefault(key string, defaultValue string) string
	WithScopes(scopes ...string) Reporter
}

// RWReporter extends Reporter with read-write methods.
type RWReporter interface {
	Reporter
	Delete(key string)
	Set(key string, value string)
}

func SplitTrimCompact(str string) []string {
	strs := []string{}
	for _, str = range strings.Split(str, ",") {
		if str = strings.TrimSpace(str); str != "" {
			strs = append(strs, str)
		}
	}
	return strs
}
