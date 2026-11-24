# CMP kubectl shell switch

The dedicated CMP kubectl shell endpoint is controlled by
`CMP_KUBECTL_SHELL_ENABLED`.

It defaults to `false`. In this state, requests to either endpoint below return
`404` before cluster permission checks, service account token lookup, or pod
exec:

```text
/api/websocket/k8s/clusters/{clusterName}/kubectl-shell
/api/k8s/clusters/{clusterName}/kubectl-shell
```

Set the variable to `true` and restart the CMP service to enable the endpoint:

```yaml
env:
  - name: CMP_KUBECTL_SHELL_ENABLED
    value: "true"
```

This switch only controls the dedicated kubectl shell. It does not disable pod
logs or other Kubernetes websocket and exec APIs.
