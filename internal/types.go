package internal

import (
	"crypto/x509"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	utilpki "github.com/cert-manager/cert-manager/pkg/util/pki"
)

type CertificateRequest interface {
	GetRequest() *x509.CertificateRequest
	GetNamespace() string
	GetName() string
}

func NewCertificateRequest(request *v1.CertificateRequest) (CertificateRequest, error) {
	csr, err := utilpki.DecodeX509CertificateRequestBytes(request.Spec.Request)
	if err != nil {
		return nil, err
	}
	return &certificateRequest{request: csr, namespace: request.Namespace, name: request.Name}, nil
}

type certificateRequest struct {
	namespace string
	name      string
	request   *x509.CertificateRequest
}

func (cr certificateRequest) GetRequest() *x509.CertificateRequest {
	return cr.request
}

func (cr certificateRequest) GetNamespace() string {
	return cr.namespace
}

func (cr certificateRequest) GetName() string {
	return cr.name
}
