#!/usr/bin/env bash

set -e pipefail

usage() {
    cat <<EOF
usage: ${0} <key directory> [OPTIONS]

OPTIONS:
       --webhook            MutatingWebhookConfiguration name.
EOF
    exit 1
}

while [[ $# -gt 1 ]]; do
    case ${2} in
        --webhook)
            webhook="$3"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

[ -z "${webhook}" ] && webhook=monitor-injector

kubectl delete MutatingWebhookConfiguration ${webhook} || true
