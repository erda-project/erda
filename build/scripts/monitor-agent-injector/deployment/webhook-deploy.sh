#!/usr/bin/env bash

set -e pipefail

basedir="$(dirname "$0")"
keydir="/netdata/dice-ops/dice-config/certificates"

# Read the PEM-encoded CA certificate, base64 encode it, and replace the `${CA_PEM_B64}` placeholder in the YAML
# template with it. Then, create the Kubernetes resources.
ca_pem_b64="$(openssl base64 -A <"${keydir}/monitor-agent-injector-ca.pem")"
sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"${basedir}/webhook.yaml.template" \
    | kubectl create -f -

echo "The webhook server has been deployed and configured!"
