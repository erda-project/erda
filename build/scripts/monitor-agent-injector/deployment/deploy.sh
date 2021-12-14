#!/usr/bin/env bash

set -e pipefail

usage() {
    cat <<EOF
usage: ${0} <key directory> [OPTIONS]

OPTIONS:
       --service          Service name of webhook.
       --namespace        Namespace where webhook service and secret reside.
       --secret           Secret name for CA certificate and server certificate/key pair.
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
        *)
            usage
            ;;
    esac
    shift
done

[ -z "${service}" ] && service=monitor-injector
[ -z "${namespace}" ] && namespace=default
[ -z "${secret}" ] && secret=monitor-injector-tls

basedir="$(dirname "$0")"
keydir="$(mktemp -d)"

# Generate keys into a temporary directory.
echo "Generating TLS keys ..."
"${basedir}/generate-keys.sh" "$keydir" --service ${service} --namespace ${namespace}

# Create the TLS secret for the generated keys.
kubectl -n ${namespace} create secret tls ${secret} \
    --cert "${keydir}/${service}.pem" \
    --key "${keydir}/${service}-key.pem"

# Read the PEM-encoded CA certificate, base64 encode it, and replace the `${CA_PEM_B64}` placeholder in the YAML
# template with it. Then, create the Kubernetes resources.
ca_pem_b64="$(openssl base64 -A <"${keydir}/monitor-agent-ca.pem")"
sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"${basedir}/deployment.yaml.template" \
    | kubectl create -f -

# Delete the key directory to prevent abuse (DO NOT USE THESE KEYS ANYWHERE ELSE).
rm -rf "$keydir"

echo "The webhook server has been deployed and configured!"
