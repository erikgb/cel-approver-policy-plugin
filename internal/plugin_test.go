package internal

import (
	"testing"
)

func Test_validatePluginValues(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]string
		wantErr bool
	}{
		{name: "no-values"},
		{name: "valid-value", values: map[string]string{"dnsNames": "has(csr.name)"}},
		{name: "err-invalid-value", values: map[string]string{"dnsNames": "foo"}, wantErr: true},
		{name: "err-invalid-key", values: map[string]string{"foo": "has(csr.name)"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePluginValues(tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePluginValues()error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
