#!/usr/bin/env bash

set -euo pipefail

: ${1?'missing key directory'}
key_dir="$1"

usage() {
    cat <<EOF
usage: ${0} <key directory> [OPTIONS]

OPTIONS:
       --service          Service name of webhook.
       --namespace        Namespace where webhook service reside.
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
        *)
            usage
            ;;
    esac
    shift
done

[ -z "${service}" ] && service=webhook-svc
[ -z "${namespace}" ] && namespace=default

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

# Generate config
cd "$key_dir"
cat <<EOF >> ${service}.cert.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

# Generate the CA cert and private key
openssl req -nodes -new -x509 -keyout ${service}-ca-key.pem -out ${service}-ca.pem -subj "/CN=ca.${service}.${namespace}.svc"
# Generate the private key for the server
openssl genrsa -out ${service}-key.pem 2048
# Generate a Certificate Signing Request (CSR) for the private key, and sign it with the private key of the CA.
openssl req -new -key ${service}-key.pem -subj "/CN=${service}.${namespace}.svc" -config ${service}.cert.conf \
    | openssl x509 -req -CA ${service}-ca.pem -CAkey ${service}-ca-key.pem -CAcreateserial -out ${service}.pem -extensions v3_req -extfile ${service}.cert.conf


