package logger

import (
	"testing"
)

const invalidLogLevel = "GopherLevel"

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "debug -> ok",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "panic -> ok",
			level:   "panic",
			wantErr: false,
		},
		{
			name:    "invalid level -> error",
			level:   invalidLogLevel,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLogger(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
