package config

import "fmt"

// The shared config sections.
//
// A service puts these in its own Config struct. The `mapstructure` tags tell
// Viper which YAML key fills which field when the service loads its file:
//
//	server:            →  Server
//	  host: 0.0.0.0    →  Server.Host
//	  port: 8000       →  Server.Port
//
// Each section has a Validate function.

// Server is where the HTTP server listens.
type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// ValidateServer checks that the host is set and the port is valid.
func ValidateServer(server Server, key string, problems *Problems) {
	problems.Required(key+".host", server.Host)
	problems.Port(key+".port", server.Port)
}

// Addr is the listen address, e.g. "0.0.0.0:8000".
func (server Server) Addr() string {
	return fmt.Sprintf("%s:%d", server.Host, server.Port)
}

// --------------------------------------------------------------- logging

// LogLevels are the values allowed in Logging.Level.
var LogLevels = []string{"debug", "info", "warn", "error"}

// Logging configures the service's logger.
type Logging struct {
	Level string `mapstructure:"level"`
}

// ValidateLogging checks that the level is one of LogLevels.
func ValidateLogging(logging Logging, key string, problems *Problems) {
	problems.OneOf(key+".level", logging.Level, LogLevels...)
}

// -------------------------------------------------------------- database

// SSLModes are the values allowed in Database.SSLMode (as Postgres defines them).
var SSLModes = []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}

// Database is a Postgres connection. A service connects either with URL or
// with Host, Port, Database and User.
type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	URL      string `mapstructure:"url"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

// ValidateDatabase checks:
//   - there is a way to connect: a URL, or else host, port, database and user
//   - sslmode, if set, is one of SSLModes
//   - the pool sizes are not negative, and min_conns is not above max_conns
func ValidateDatabase(db Database, key string, problems *Problems) {
	// A way to connect.
	if db.URL == "" {
		problems.Required(key+".host", db.Host)
		problems.Port(key+".port", db.Port)
		problems.Required(key+".database", db.Database)
		problems.Required(key+".user", db.User)
	} else if db.Port != 0 {
		problems.Port(key+".port", db.Port)
	}

	// SSL mode, only if one is given.
	if db.SSLMode != "" {
		problems.OneOf(key+".sslmode", db.SSLMode, SSLModes...)
	}

	// Pool sizes. 0 means "use the driver's default".
	if db.MaxConns < 0 {
		problems.Addf(key+".max_conns", "must not be negative")
	}
	if db.MinConns < 0 {
		problems.Addf(key+".min_conns", "must not be negative")
	} else if db.MaxConns > 0 && db.MinConns > db.MaxConns {
		problems.Addf(key+".min_conns", "%d is more than max_conns (%d)", db.MinConns, db.MaxConns)
	}
}
