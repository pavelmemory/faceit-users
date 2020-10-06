package config

import (
	"github.com/kelseyhightower/envconfig"
)

// NewEnvSettings returns EnvSettings initialized from environment variables.
// `prefix` allows to add an extra prefix that needs to be used with all env var names.
func NewEnvSettings(prefix string) (EnvSettings, error) {
	var s EnvSettings
	return s, envconfig.Process(prefix, &s)
}

// EnvSettings reads settings from environment variables.
type EnvSettings struct {
	EnvHTTPListenPort int    `envconfig:"HTTP_PORT" default:"8080"`
	EnvLogLevel       string `envconfig:"LOG_LEVEL" default:"info"`
	EnvStorageAddr    string `envconfig:"STORAGE_ADDR" default:"0.0.0.0:5432"`
	EnvStoragePwd     string `envconfig:"STORAGE_PWD"`
}

// HTTPPort returns a port number to listening for incoming HTTP connections.
func (es EnvSettings) HTTPPort() int {
	return es.EnvHTTPListenPort
}

// LogLevel returns a logging level.
func (es EnvSettings) LogLevel() string {
	return es.EnvLogLevel
}

// StorageAddr returns address of the main persistence storage.
func (es EnvSettings) StorageAddr() string {
	return es.EnvStorageAddr
}

func (es EnvSettings) StoragePwd() string {
	return es.EnvStoragePwd
}
