package config

import (
	"strings"
	"testing"
)

// check runs validate and asserts the error names exactly the want paths.
func check(t *testing.T, validate func(*Problems), want ...string) {
	t.Helper()
	var p Problems
	validate(&p)
	err := p.Err()

	if len(want) == 0 {
		if err != nil {
			t.Fatalf("unexpected problems: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("no problems, want %v", want)
	}
	for _, path := range want {
		if !strings.Contains(err.Error(), path+":") {
			t.Errorf("missing %s in:\n%v", path, err)
		}
	}
	if got := strings.Count(err.Error(), "\n  "); got != len(want) {
		t.Errorf("%d problems, want %d:\n%v", got, len(want), err)
	}
}

func TestProblemsErrIsSortedAndCounted(t *testing.T) {
	var p Problems
	p.Addf("b.key", "second")
	p.Addf("a.key", "first")

	want := "config: 2 problem(s)\n  a.key: first\n  b.key: second"
	if got := p.Err().Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProblemsEmptyIsNil(t *testing.T) {
	var p Problems
	p.Required("x", "set")
	p.Port("y", 8080)
	p.OneOf("z", "info", LogLevels...)
	if err := p.Err(); err != nil {
		t.Errorf("Err = %v, want nil", err)
	}
}

func TestServer(t *testing.T) {
	tests := []struct {
		name string
		s    Server
		want []string
	}{
		{"valid", Server{Host: "0.0.0.0", Port: 8080}, nil},
		{"empty", Server{}, []string{"server.host", "server.port"}},
		{"blank host", Server{Host: "  ", Port: 8080}, []string{"server.host"}},
		{"port too high", Server{Host: "h", Port: 70000}, []string{"server.port"}},
		{"negative port", Server{Host: "h", Port: -1}, []string{"server.port"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, func(problems *Problems) { ValidateServer(tt.s, "server", problems) }, tt.want...)
		})
	}
}

func TestServerAddr(t *testing.T) {
	if got := (Server{Host: "0.0.0.0", Port: 8080}).Addr(); got != "0.0.0.0:8080" {
		t.Errorf("Addr = %q", got)
	}
}

func TestLogging(t *testing.T) {
	check(t, func(problems *Problems) { ValidateLogging(Logging{Level: "warn"}, "logging", problems) })
	check(t, func(problems *Problems) { ValidateLogging(Logging{}, "logging", problems) }, "logging.level")
	check(t, func(problems *Problems) { ValidateLogging(Logging{Level: "verbose"}, "logging", problems) }, "logging.level")
}

func TestDatabase(t *testing.T) {
	fields := Database{Host: "h", Port: 5432, Database: "d", User: "u"}

	tests := []struct {
		name string
		d    Database
		want []string
	}{
		{"url only", Database{URL: "postgres://x"}, nil},
		{"fields only", fields, nil},
		{"neither", Database{}, []string{"db.host", "db.port", "db.database", "db.user"}},
		{"url with bad port", Database{URL: "postgres://x", Port: 99999}, []string{"db.port"}},
		{"bad sslmode", Database{URL: "postgres://x", SSLMode: "on"}, []string{"db.sslmode"}},
		{"negative pool", Database{URL: "postgres://x", MaxConns: -1, MinConns: -1}, []string{"db.max_conns", "db.min_conns"}},
		{"min above max", Database{URL: "postgres://x", MaxConns: 2, MinConns: 5}, []string{"db.min_conns"}},
		{"min with unset max", Database{URL: "postgres://x", MinConns: 5}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, func(problems *Problems) { ValidateDatabase(tt.d, "db", problems) }, tt.want...)
		})
	}
}
