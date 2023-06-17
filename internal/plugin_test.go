package internal

import (
	"crypto/x509"
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net/url"
	"reflect"
	"testing"
)

func Test_validatePluginValues(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]string
		wantErr bool
	}{
		{name: "no-values"},
		{name: "valid-value", values: map[string]string{"dnsNames": "has(cr.name)"}},
		{name: "err-invalid-value", values: map[string]string{"dnsNames": "foo"}, wantErr: true},
		{name: "err-invalid-key", values: map[string]string{"foo": "has(cr.name)"}, wantErr: true},
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

func Test_validateCertificateRequest(t *testing.T) {
	type args struct {
		cr       CertificateRequest
		cpValues map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    field.ErrorList
		wantErr bool
	}{
		{name: "valid-dnsNames", args: args{
			cr:       certificateRequest{request: &x509.CertificateRequest{DNSNames: []string{"service.foo-ns.apps"}}, namespace: "foo-ns"},
			cpValues: map[string]string{"dnsNames": "self.endsWith(cr.namespace + '.apps')"},
		}},
		{name: "valid-uris", args: args{
			cr:       certificateRequest{request: &x509.CertificateRequest{URIs: []*url.URL{mustParseURL("spiffe://acme.com/ns/foo-ns/sa/bar")}}, namespace: "foo-ns"},
			cpValues: map[string]string{"uris": "self.startsWith('spiffe://acme.com/ns/%s/sa'.format([cr.namespace]))"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateCertificateRequest(tt.args.cr, tt.args.cpValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCertificateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateCertificateRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Errorf("cannot parse '%v': %v", rawURL, err))
	}
	return u
}
