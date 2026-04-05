package logging

import (
	"testing"
)

func TestIsValidLogLevel(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"TRACE":     {input: "TRACE", want: true},
		"DEBUG":     {input: "DEBUG", want: true},
		"INFO":      {input: "INFO", want: true},
		"WARN":      {input: "WARN", want: true},
		"ERROR":     {input: "ERROR", want: true},
		"lowercase": {input: "debug", want: true},
		"mixed":     {input: "Info", want: true},
		"invalid":   {input: "VERBOSE", want: false},
		"empty":     {input: "", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isValidLogLevel(tt.input)
			if got != tt.want {
				t.Errorf("isValidLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLogLevel(t *testing.T) {
	tests := map[string]struct {
		envVal string
		want   string
	}{
		"not set": {
			envVal: "",
			want:   "",
		},
		"valid level": {
			envVal: "DEBUG",
			want:   "DEBUG",
		},
		"lowercase valid": {
			envVal: "info",
			want:   "INFO",
		},
		"invalid falls to TRACE": {
			envVal: "GARBAGE",
			want:   "TRACE",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(EnvLog, tt.envVal)
			got := LogLevel()
			if got != tt.want {
				t.Errorf("LogLevel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsDebugOrHigher(t *testing.T) {
	tests := map[string]struct {
		envVal string
		want   bool
	}{
		"DEBUG": {envVal: "DEBUG", want: true},
		"TRACE": {envVal: "TRACE", want: true},
		"INFO":  {envVal: "INFO", want: false},
		"empty": {envVal: "", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(EnvLog, tt.envVal)
			got := IsDebugOrHigher()
			if got != tt.want {
				t.Errorf("IsDebugOrHigher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTrace(t *testing.T) {
	tests := map[string]struct {
		envVal string
		want   bool
	}{
		"TRACE": {envVal: "TRACE", want: true},
		"DEBUG": {envVal: "DEBUG", want: false},
		"empty": {envVal: "", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(EnvLog, tt.envVal)
			got := IsTrace()
			if got != tt.want {
				t.Errorf("IsTrace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSet(t *testing.T) {
	tests := map[string]struct {
		envVal string
		want   bool
	}{
		"set":   {envVal: "INFO", want: true},
		"empty": {envVal: "", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(EnvLog, tt.envVal)
			got := IsSet()
			if got != tt.want {
				t.Errorf("IsSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
