# CMP shell console switch

The CMP kubectl shell and container terminal endpoints are controlled by
`CMP_SHELL_CONSOLE_ENABLED`.

It defaults to `true`. Set the variable to `false` and restart the CMP service
to disable both terminal types. In this state, requests matching either the
dedicated kubectl shell or the Kubernetes pod exec subresource return `404`
before cluster permission checks, service account token lookup, or pod exec:

```text
/api/k8s/clusters/{clusterName}/kubectl-shell
/api/k8s/clusters/{clusterName}/{version}/namespaces/{namespace}/pods/{pod}/exec
```

To disable both CMP terminal types:

```yaml
env:
  - name: CMP_SHELL_CONSOLE_ENABLED
    value: "false"
```

This switch controls only CMP. It does not affect pod logs, other Kubernetes
APIs, or the DevOps container console at `/api/terminal`.
