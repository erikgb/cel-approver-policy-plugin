package internal

import (
	"testing"
)

func Test_Validator_Compile(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{name: "no-var-use", expr: "'www.example.com'.endsWith('.com')"},
		{name: "simple-checks", expr: "size(cr.namespace) < 24"},
		{name: "standard-macros", expr: "[1,2,3].all(i, i % 2 > 0)"},
		{name: "extended-string-function-library", expr: "self.startsWith('spiffe://trust-domain.com/')"},
		{name: "doc-example", expr: "['.svc', '.svc.cluster.local'].exists(d, self.endsWith(cr.namespace + d))"},
		{name: "err-no-expression", wantErr: true},
		{name: "err-undeclared-vars", expr: "foo = bar", wantErr: true},
		{name: "err-must-return-bool", expr: "size('foo')", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{Expression: tt.expr}
			err := v.Compile()
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_Validator_Validate(t *testing.T) {
	v := &Validator{Expression: "self.startsWith('spiffe://acme.com/ns/%s/sa/'.format([cr.namespace]))"}
	err := v.Compile()
	if err != nil {
		t.Errorf("Compile() error = %v", err)
	}

	type args struct {
		val string
		cr  CertificateRequest
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "correct-namespace", args: args{val: "spiffe://acme.com/ns/foo-ns/sa/bar", cr: certificateRequest{namespace: "foo-ns"}}, want: true},
		{name: "wrong-namespace", args: args{val: "spiffe://acme.com/ns/foo-ns/sa/bar", cr: certificateRequest{namespace: "bar-ns"}}, want: false},
		{name: "invalid-spiffeid", args: args{val: "spiffe://example.com", cr: certificateRequest{namespace: "foo-ns"}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v.Validate(tt.args.val, tt.args.cr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Validate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
