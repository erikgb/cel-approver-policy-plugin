# cel-approver-policy-plugin

This repo contains an experimental CEL
[cert-manager/approver-policy plugin](https://cert-manager.io/docs/projects/approver-policy/#plugins)
that allows to specify CEL expressions used to decide if `CertificateRequest`s can be approved

## Installation

[Plugins](https://cert-manager.io/docs/projects/approver-policy/#plugins) are external approvers that are built
into [approver-policy](https://cert-manager.io/docs/projects/approver-policy/) at compile time.

To install approver-policy with cel-approver-policy-plugin follow the
[approver-policy installation instructions](https://cert-manager.io/docs/projects/approver-policy/#installation),
but replace the default approver-policy image with an image from this project.

All commits on the default branch will push to `ghcr.io/erikgb/cel-approver-policy-plugin:main`.

Supported "extra" flags:

- `policy-with-no-plugin-allowed`: Whether a CertificateRequestPolicy without cel-approver-policy plugin should be allowed in the cluster

## Usage

Example `CertificateRequestPolicy` that allows issuance if all `dnsName`s ends with `<namespace>.svc`
or `<namespace>.svc.cluster.local`:

```yaml
apiVersion: policy.cert-manager.io/v1alpha1
kind: CertificateRequestPolicy
metadata:
  name: cluster-local-service
spec:
  allowed:
    ... # Be aware that using a plugin does not disable the core approver - a CertificateRequest still has to match the allowed block here even if a plugin is specified
  selector:
   ...
  plugins:
    cel-approver-policy-plugin:
      values:
        dnsNames: >-
          ['.svc', '.svc.cluster.local'].exists(d, self.endsWith(cr.namespace + d))
```

### Writing CEL expressions for this plugin

The plugin has access to the same CEL community libraries as
[Kubernetes](https://kubernetes.io/docs/reference/using-api/cel/#cel-community-libraries):

- CEL standard functions, defined in the [list of standard definitions](https://github.com/google/cel-spec/blob/master/doc/langdef.md#list-of-standard-definitions)
- CEL standard [macros](https://github.com/google/cel-spec/blob/v0.7.0/doc/langdef.md#macros)
- CEL [extended string function library](https://pkg.go.dev/github.com/google/cel-go/ext#Strings)

The following CEL variables are available to use in expressions:

- `self`: the `string` typed value to validate obtained from the decoded CSR
- `cr`: a `map` with selected fields from `CertificateRequest`; currently `namespace` and `name` keys
