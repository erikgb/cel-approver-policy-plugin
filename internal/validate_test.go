package internal

import (
	"reflect"
	"testing"
)

func TestNewValidator(t *testing.T) {
	type args struct {
		expression string
	}
	tests := []struct {
		name    string
		args    args
		want    *validator
		wantErr bool
	}{
		{name: "no-expression", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewValidator(tt.args.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewValidator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValidator() got = %v, want %v", got, tt.want)
			}
		})
	}
}
