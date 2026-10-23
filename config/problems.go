// Package config holds the configuration sections every DAN service shares,
// the functions that validate them, and Problems, the list those functions
// record into.
//
// How a service uses it:
//
//	var problems config.Problems
//	config.ValidateServer(cfg.Server, "server", &problems)   // shared check
//	if ... { problems.Addf("datasets", "...") }               // service's own check
//	return problems.Err()                                     // nil if nothing was wrong
//
// Nothing here reads a config file. The service reads its own file, then
// hands the values it read to these functions.
package config

import (
	"fmt"
	"sort"
	"strings"
)

// Problems is the list of everything wrong with a config.
// Each line starts with the YAML key it is about, e.g.
// "server.port: 70000 is out of range".
type Problems struct {
	list []string
}

// Addf adds one problem about the YAML key named key.
func (problems *Problems) Addf(key, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	problems.list = append(problems.list, key+": "+message)
}

// Required adds a problem if value is empty or only spaces.
func (problems *Problems) Required(key, value string) {
	if strings.TrimSpace(value) == "" {
		problems.Addf(key, "required")
	}
}

// Port adds a problem if port is missing or not a valid TCP port (1-65535).
func (problems *Problems) Port(key string, port int) {
	if port == 0 {
		problems.Addf(key, "required")
		return
	}
	if port < 1 || port > 65535 {
		problems.Addf(key, "%d is out of range", port)
	}
}

// OneOf adds a problem if value is empty or not one of allowed.
func (problems *Problems) OneOf(key, value string, allowed ...string) {
	if value == "" {
		problems.Addf(key, "required")
		return
	}
	for _, option := range allowed {
		if value == option {
			return
		}
	}
	problems.Addf(key, "%q is not one of %s", value, strings.Join(allowed, ", "))
}

// Err returns nil if no problems were added. Otherwise it returns one error
// listing every problem, sorted by key:
//
//	config: 2 problem(s)
//	  server.host: required
//	  server.port: 70000 is out of range
func (problems *Problems) Err() error {
	if len(problems.list) == 0 {
		return nil
	}

	sorted := append([]string(nil), problems.list...)
	sort.Strings(sorted)

	return fmt.Errorf("config: %d problem(s)\n  %s", len(sorted), strings.Join(sorted, "\n  "))
}
