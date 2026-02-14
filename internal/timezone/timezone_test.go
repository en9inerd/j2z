package timezone

import (
	"testing"
	"time"
)

func TestGetTimeZone(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantTZ string
	}{
		{
			name:   "valid timezone",
			input:  "America/New_York",
			wantTZ: "America/New_York",
		},
		{
			name:   "UTC",
			input:  "UTC",
			wantTZ: "UTC",
		},
		{
			name:   "invalid timezone falls back to local",
			input:  "Invalid/Timezone",
			wantTZ: time.Local.String(),
		},
		{
			name:   "empty string falls back to local",
			input:  "",
			wantTZ: time.Local.String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTimeZone(tt.input)
			if got.String() != tt.wantTZ {
				t.Errorf("GetTimeZone(%q) = %q, want %q", tt.input, got.String(), tt.wantTZ)
			}
		})
	}
}
