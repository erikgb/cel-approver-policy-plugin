package internal

import (
	"context"
	"fmt"
	"github.com/cert-manager/approver-policy/pkg/apis/policy/v1alpha1"
	"github.com/cert-manager/approver-policy/pkg/approver"
	"github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const name = "cel-approver-policy-plugin"

var (
	valueKeys = map[string]CSRAttribute{"dnsNames": CSRAttrDNSNames, "uris": CSRAttrURIs}
	basePath  = field.NewPath("spec", "plugins", name)
)

// CELPlugin is an implementation of approver-policy.Interface
// https://github.com/cert-manager/approver-policy/blob/v0.6.3/pkg/approver/approver.go#L27-L53
type CELPlugin struct {
	// whether a CertificateRequestPolicy without this plugin defined should
	// be allowed
	policyWithNoPluginAllowed bool
	enqueueChan               <-chan string
	log                       logr.Logger
}

var _ approver.Interface = &CELPlugin{}

func (e *CELPlugin) Name() string {
	return name
}

func (e *CELPlugin) RegisterFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&e.policyWithNoPluginAllowed, "policy-with-no-plugin-allowed", true, "Whether a CertificateRequestPolicy without cel-approver-policy plugin should be allowed in the cluster")
}

// Prepare is called once when the approver plugin is being initialized and before the controllers have started.
// https://github.com/cert-manager/approver-policy/blob/v0.6.3/pkg/internal/cmd/cmd.go#L86
func (e *CELPlugin) Prepare(_ context.Context, log logr.Logger, _ manager.Manager) error {
	e.log = log.WithName(name)
	// The example plugin does not utilize this channel
	e.enqueueChan = make(<-chan string)
	return nil
}

// Evaluate will be called when a CertificateRequest is synced with each
// combination of the CertificateRequest and an applicable
// CertificateRequestPolicy that has this plugin enabled.
// For any combination:
// - If Evaluate returns an error, the CertificateRequest will not be denied or
// approved and will be resynced.
// - If Evalute returns Denied, the CertificateRequest will be Denied.
// - If Evaluate returns Approved and all other relevant plugins (including core
// approver in cert-manager/approver-policy) also return Approved, the
// CertificateRequst will be approved.
// https://github.com/cert-manager/approver-policy/blob/v0.6.3/pkg/internal/approver/manager/review.go#L128
func (e *CELPlugin) Evaluate(_ context.Context, crp *v1alpha1.CertificateRequestPolicy, cr *v1.CertificateRequest) (approver.EvaluationResponse, error) {
	e.log.V(5).Info("evaluating CertificateRequest", "certificaterequest", cr.Name, "certificaterequestpolicy", crp.Name)
	plugin, ok := crp.Spec.Plugins[name]
	if !ok {
		if e.policyWithNoPluginAllowed {
			// nothing to do here
			return approver.EvaluationResponse{Result: approver.ResultNotDenied}, nil
		}
		msg := fmt.Sprintf("required plugin %s is not defined", name)
		return approver.EvaluationResponse{Result: approver.ResultDenied, Message: msg}, nil
	}

	req, err := NewCertificateRequest(cr)
	if err != nil {
		return approver.EvaluationResponse{}, err
	}
	allErrors, err := validateCertificateRequest(req, plugin.Values)
	if err != nil {
		return approver.EvaluationResponse{}, err
	}
	if len(allErrors) > 0 {
		return approver.EvaluationResponse{Result: approver.ResultDenied, Message: allErrors.ToAggregate().Error()}, nil
	}

	return approver.EvaluationResponse{Result: approver.ResultNotDenied}, nil
}

// Validate will be run by the approver-policy's admission webhook.
// https://github.com/cert-manager/approver-policy/blob/v0.6.3/deploy/charts/approver-policy/templates/webhook.yaml#L22-L52
// An error returned here will result in failed creation of update of the
// CertificateRequestPolicy being validated.
func (e *CELPlugin) Validate(_ context.Context, crp *v1alpha1.CertificateRequestPolicy) (approver.WebhookValidationResponse, error) {
	e.log.V(5).Info("validating CertificateRequestPolicy", "certificaterequestpolicy", crp.Name)
	plugin, ok := crp.Spec.Plugins[name]
	if !ok {
		if e.policyWithNoPluginAllowed {
			// nothing to do here
			return approver.WebhookValidationResponse{Allowed: true, Errors: nil}, nil
		}
		e := fmt.Errorf("required plugin %s is not defined", name)
		return approver.WebhookValidationResponse{Allowed: false, Errors: []*field.Error{field.Required(basePath, e.Error())}}, nil
	}

	allErrors := validatePluginValues(plugin.Values)
	if len(allErrors) > 0 {
		return approver.WebhookValidationResponse{Allowed: false, Errors: allErrors}, nil
	}

	return approver.WebhookValidationResponse{Allowed: true}, nil
}

// Ready will be called every time a CertificateRequestPolicy is reconciled in
// response to events against CertificateRequestPolicy as well as events sent by
// the plugin via EnqueueChan. CertificateRequestPolicy's Ready status is set
// depending on the response returned by Ready methods of applicable plugins
// (including core approver) - if any returns false, Ready status will be false.
// https://github.com/cert-manager/approver-policy/blob/v0.6.3/pkg/internal/controllers/certificaterequestpolicies.go#L184
func (e *CELPlugin) Ready(_ context.Context, crp *v1alpha1.CertificateRequestPolicy) (approver.ReconcilerReadyResponse, error) {
	e.log.V(5).Info("validating that CertificateRequestPolicy is ready", "certificaterequestpolicy", crp.Name)
	plugin, ok := crp.Spec.Plugins[name]
	if !ok {
		if e.policyWithNoPluginAllowed {
			// nothing to do here
			return approver.ReconcilerReadyResponse{Ready: true, Errors: nil}, nil
		}
		e := fmt.Errorf("required plugin %s is not defined", name)
		return approver.ReconcilerReadyResponse{Ready: false, Errors: []*field.Error{field.Required(basePath, e.Error())}}, nil
	}

	allErrors := validatePluginValues(plugin.Values)
	if len(allErrors) > 0 {
		return approver.ReconcilerReadyResponse{Ready: false, Errors: allErrors}, nil
	}

	return approver.ReconcilerReadyResponse{Ready: true}, nil
}

func validateCertificateRequest(cr CertificateRequest, cpValues map[string]string) (field.ErrorList, error) {
	var allErrors field.ErrorList

	for k, attr := range valueKeys {
		expr, ok := cpValues[k]
		if !ok {
			continue
		}
		validator := &Validator{Expression: expr}
		// TODO: Consider caching compiled validators
		err := validator.Compile()
		if err != nil {
			return nil, err
		}
		for i, val := range attr.GetValues(cr) {
			valid, err := validator.Validate(val, cr)
			if err != nil {
				return nil, err
			}
			if !valid {
				e := fmt.Errorf("does not satisfy policy expression %s", expr)
				allErrors = append(allErrors, field.Invalid(field.NewPath("CSR", string(attr)).Index(i), val, e.Error()))
			}
		}

	}

	return allErrors, nil
}

func validatePluginValues(values map[string]string) field.ErrorList {
	var allErrors field.ErrorList
	for k := range valueKeys {
		expr, ok := values[k]
		if !ok {
			continue
		}
		validator := &Validator{Expression: expr}
		// TODO: Consider caching compiled validators
		err := validator.Compile()
		if err != nil {
			allErrors = append(allErrors, field.Invalid(basePath.Child(k), expr, err.Error()))
		}
		// delete processed key to validate for unsupported keys afterwards
		delete(values, k)
	}
	for key := range values {
		allErrors = append(allErrors, field.NotSupported(basePath, key, maps.Keys(valueKeys)))
	}
	return allErrors
}

// EnqueueChan returns a channel to which the plugin can send applicable
// CertificateRequestPolicy names to cause them to be resynced. This is useful
// if readiness of CertificateRequestPolicies with the plugin enabled needs to
// be re-evaluated in response to changes in some external system used by the
// plugin.
func (e *CELPlugin) EnqueueChan() <-chan string {
	return e.enqueueChan
}
