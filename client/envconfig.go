package client

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/config"
)

type envconfigReporter struct {
	prefix string
	config *Config
	onceFn sync.Once
	err    error
}

func NewEnvconfigReporter(prefix string) *envconfigReporter {
	return &envconfigReporter{
		prefix: prefix,
		config: NewConfig(),
		err:    nil,
	}
}

func (r *envconfigReporter) Get(key string) (string, error) {
	if err := r.once(); err != nil {
		return "", err
	}

	normalizedKey := normalizeKey(key)
	v := reflect.Indirect(reflect.ValueOf(r.config))
	field := v.FieldByNameFunc(func(name string) bool {
		return normalizeKey(name) == normalizedKey
	})
	if field.IsZero() {
		return "", config.ErrorKeyNotFound(key)
	}

	return field.String(), nil
}

func (r *envconfigReporter) once() error {
	var err error
	r.onceFn.Do(func() {
		err = envconfig.Process(r.prefix, r.config)
	})
	return err
}

func normalizeKey(key string) string {
	lower := strings.ToLower(key)
	return nonWordRe.ReplaceAllString(lower, "")
}

var nonWordRe = regexp.MustCompile(`\W+`)

func (r *envconfigReporter) GetWithDefault(key string, defaultValue string) string {
	v, err := r.Get(key)
	if err != nil {
		return defaultValue
	}
	return v
}

func (r *envconfigReporter) WithScopes(scopes ...string) config.Reporter {
	r.prefix = strings.ToUpper(strings.Join(append([]string{r.prefix}, scopes...), "_"))
	return r
}
