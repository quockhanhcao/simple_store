package config

import (
	"log/slog"
	"strings"
	"testing"
)

func fakeEnv(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}

func TestLoad(t *testing.T) {
	t.Parallel()

	const defaultAddr = ":5200"
	const dbURL = "postgres://u:p@localhost:5432/storedb"

	tests := []struct {
		name     string
		env      map[string]string
		want     Config
		wantErrs []string
	}{
		{
			name: "only required var set, defaults applied",
			env:  map[string]string{"DATABASE_URL": dbURL},
			want: Config{Env: "Production", HTTPAddr: defaultAddr, DatabaseURL: dbURL, LogFormat: "text", LogLevel: slog.LevelInfo},
		},
		{
			name: "all vars set",
			env: map[string]string{
				"DATABASE_URL": dbURL,
				"HTTP_ADDR":    ":9999",
				"APP_ENV":      "Development",
				"LOG_FORMAT":   "json",
				"LOG_LEVEL":    "debug",
			},
			want: Config{Env: "Development", HTTPAddr: ":9999", DatabaseURL: dbURL, LogFormat: "json", LogLevel: slog.LevelDebug},
		},
		{
			name: "values are case-insensitive and normalized",
			env: map[string]string{
				"DATABASE_URL": dbURL,
				"APP_ENV":      "dEvElOpMeNt",
				"LOG_FORMAT":   "JSON",
				"LOG_LEVEL":    "WARN",
			},
			want: Config{Env: "Development", HTTPAddr: defaultAddr, DatabaseURL: dbURL, LogFormat: "json", LogLevel: slog.LevelWarn},
		},
		{
			name: "staging is accepted",
			env:  map[string]string{"DATABASE_URL": dbURL, "APP_ENV": "staging"},
			want: Config{Env: "Staging", HTTPAddr: defaultAddr, DatabaseURL: dbURL, LogFormat: "text", LogLevel: slog.LevelInfo},
		},
		{
			name:     "missing DATABASE_URL",
			env:      map[string]string{},
			wantErrs: []string{"DATABASE_URL is required"},
		},
		{
			name:     "invalid LOG_FORMAT reports the raw value",
			env:      map[string]string{"DATABASE_URL": dbURL, "LOG_FORMAT": "XML"},
			wantErrs: []string{`LOG_FORMAT must be "text" or "json", got "XML"`},
		},
		{
			name:     "invalid LOG_LEVEL",
			env:      map[string]string{"DATABASE_URL": dbURL, "LOG_LEVEL": "verbose"},
			wantErrs: []string{`invalid LOG_LEVEL "verbose"`},
		},
		{
			name:     "invalid APP_ENV reports the raw value",
			env:      map[string]string{"DATABASE_URL": dbURL, "APP_ENV": "Prod"},
			wantErrs: []string{`APP_ENV must be`, `got "Prod"`},
		},
		{
			name: "all problems are reported together",
			env:  map[string]string{"LOG_FORMAT": "xml", "LOG_LEVEL": "loud", "APP_ENV": "qa"},
			wantErrs: []string{
				"DATABASE_URL is required",
				`got "xml"`,
				`invalid LOG_LEVEL "loud"`,
				`got "qa"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Load(fakeEnv(tt.env), defaultAddr)

			if len(tt.wantErrs) > 0 {
				if err == nil {
					t.Fatalf("Load() error = nil, want error containing %q", tt.wantErrs)
				}
				for _, want := range tt.wantErrs {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("Load() error = %q\nwant it to contain %q", err, want)
					}
				}
				if got != nil {
					t.Errorf("Load() config = %+v, want nil on error", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if *got != tt.want {
				t.Errorf("Load() =\n %+v\nwant\n %+v", *got, tt.want)
			}
		})
	}
}

func TestIsDevelopment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env  string
		want bool
	}{
		{"Development", true},
		{"development", true},
		{"Production", false},
		{"Staging", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Parallel()
			c := Config{Env: tt.env}
			if got := c.IsDevelopment(); got != tt.want {
				t.Errorf("Config{Env: %q}.IsDevelopment() = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}
