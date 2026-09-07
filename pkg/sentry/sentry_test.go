package exceptions

import (
	"testing"

	"maphraohom.app/maphraohom-backoffice/config"
)

func TestSentryInitialize(t *testing.T) {
	config.Global = config.NewConfig()
	tests := []struct {
		name string
	}{
		{
			name: "Test SentryInitialize",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Initialize()
		})
	}
}
