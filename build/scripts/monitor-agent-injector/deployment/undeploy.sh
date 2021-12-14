#!/usr/bin/env bash

set -e pipefail

usage() {
    cat <<EOF
usage: ${0} <key directory> [OPTIONS]

OPTIONS:
       --service            Service name of webhook.
       --namespace          Namespace where webhook service and secret reside.
       --secret             Secret name for CA certificate and server certificate/key pair.
       --webhook            MutatingWebhookConfiguration name.
       --deployment         Deployment name.
EOF
    exit 1
}

while [[ $# -gt 1 ]]; do
    case ${2} in
        --service)
            service="$3"
            shift
            ;;
        --namespace)
            namespace="$3"
            shift
            ;;
        --secret)
            secret="$3"
            shift
            ;;
        --webhook)
            webhook="$3"
            shift
            ;;
        --deployment)
            deployment="$3"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

[ -z "${webhook}" ] && webhook=monitor-injector
[ -z "${namespace}" ] && namespace=default
[ -z "${service}" ] && service=monitor-injector
[ -z "${deployment}" ] && deployment=monitor-injector-server
[ -z "${secret}" ] && secret=monitor-injector-tls

kubectl delete MutatingWebhookConfiguration ${webhook} || true
kubectl delete service -n ${namespace} ${service} || true
kubectl delete deployment -n ${namespace} ${deployment} || true
kubectl delete secret -n ${namespace} ${secret} || true
