# cel-approver-policy-plugin

This repo contains an experimental CEL
[cert-manager/approver-policy plugin](https://cert-manager.io/docs/projects/approver-policy/#plugins)
that allows to specify CEL expressions used to decide if `CertificateRequest`s can be approved.

Validating CSR attributes with CEL could be considered a core feature in
[cert-manager/approver-policy](https://cert-manager.io/docs/projects/approver-policy/), and there are ongoing
discussions with the cert-manager maintainers to somehow merge this plugin into the core of approver-policy.

I have no plans to extend the features of this plugin at present, and will not accept pull requests attempting to do so.
But please feel free to open an issue or PR (better) for bugs.
For any questions or comments feel free to ping me on Kubernetes Slack #cert-manager-dev (@erikgb).

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
    # Be aware that using a plugin does not disable the core approver
    # A CertificateRequest still has to match the allowed block here even if a plugin is specified
    dnsNames:
      values:
        - "*"
  plugins:
    cel-approver-policy-plugin:
      values:
        dnsNames: >-
          ['.svc', '.svc.cluster.local'].exists(d, self.endsWith(cr.namespace + d))
  selector:
    issuerRef: {}
```

At this experimental stage, the plugin only supports CEL validation for a couple of commonly used CSR-fields:

- `dnsNames`
- `uris`

### Writing CEL expressions for this plugin

The plugin has access to the same CEL community libraries as
[Kubernetes](https://kubernetes.io/docs/reference/using-api/cel/#cel-community-libraries):

- CEL standard functions, defined in the [list of standard definitions](https://github.com/google/cel-spec/blob/master/doc/langdef.md#list-of-standard-definitions)
- CEL standard [macros](https://github.com/google/cel-spec/blob/v0.7.0/doc/langdef.md#macros)
- CEL [extended string function library](https://pkg.go.dev/github.com/google/cel-go/ext#Strings)

The following CEL variables are available to use in expressions:

- `self`: the `string` typed value to validate obtained from the decoded CSR
- `cr`: a `map` with selected fields from `CertificateRequest`; currently `namespace` and `name` keys
