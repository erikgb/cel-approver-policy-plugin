# cel-approver-policy-plugin

This repo contains an experimental CEL
[cert-manager/approver-policy plugin](https://cert-manager.io/docs/projects/approver-policy/#plugins).


> :warning:  This is plugin is not meant to actually be used. This repo does not contain best-practices, production
> ready code.

## Implementing a custom approver plugin

[cert-manager/approver-policy](https://cert-manager.io/docs/projects/approver-policy/) can be extended via a plugin
mechanism where a custom plugin can be written with specific logic for evaluating `CertificateRequest`s and
`CertificateRequestPolicy`s. This can then be registered with the core cert-manager/approver-policy (in Go code)
and a single image can be built that will have both the core approver and the custom plugin.

The approximate flow when writing an approver-policy plugin (that this sample implementation follows):

- implement the [`cert-manager/approver-policy.Interface`](https://github.com/cert-manager/approver-policy/blob/v0.6.3/pkg/approver/approver.go#L27-L53).
  This should contain all the logic of the new plugin for evaluating `CertificateRequest`s and `CertificateRequestPolicy`s.

- ensure that the implementation of `approver-policy.Interface` is registered with
  [the global approver registry shared with core approver](https://github.com/cert-manager/approver-policy/blob/v0.6.3/pkg/registry/registry.go#L28)

- build a single Go binary that contains the custom plugin(s) that you wish to use as well as the upstream approver-policy.
  The entrypoint should be [root command of approver-policy](https://github.com/cert-manager/approver-policy/blob/v0.6.3/cmd/main.go#L24)

- package the whole project using your favourite packaging mechanism.

## CEL plugin

This repo contains an experimental plugin `cel-approver-policy-plugin` that allows to specify CEL expressions
used to decide if `CertificateRequest`s can be approved.

See an example `CertificateRequestPolicy` that allows issuance if all `dnsName`s ends with `<namespace>.svc`
or `<namespace>.svc.cluster.local`:

```yaml
apiVersion: policy.cert-manager.io/v1alpha1
kind: CertificateRequestPolicy
metadata:
  name: tuesday
spec:
  allowed:
    ... # Be aware that using a plugin does not disable the core approver - a CertificateRequest still has to match the allowed block here even if a plugin is specified
  selector:
   ...
  plugins:
    cel-approver-policy-plugin:
      values:
        dnsNames: self.endsWith('%s.svc'.format([namespace])) || self.endsWith('%s.svc.cluster.local'.format([namespace]))
```