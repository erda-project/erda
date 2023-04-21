#!/bin/bash

set -o errexit -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

EXTERNAL_REPO_DIR="externalrepo"

# usage
usage() {
    echo "proto_fetcher.sh ACTION"
    echo "ACTION: "
    echo "    fetch        fetch all external proto files."
    echo "    cleanup      clean external repo dir (default: $EXTERNAL_REPO_DIR)."
    exit 1
}
if [ -z "$1" ]; then
    usage
fi

# common function: fetch code use git
fetch_code() {
   while [ "$#" -gt 0 ]; do
      case "$1" in
          --repo=*)
            repo_url="${1#*=}"
            ;;
          --mirror=*)
            mirror_url="${1#*=}"
            ;;
          --commit-id=*)
            commit_id="${1#*=}"
            ;;
          --branch=*)
            branch="${1#*=}"
            ;;
          *)
            echo "error: unknown option: $1" >&2
            exit 1
            ;;
      esac
      shift
  done

  if [[ -z $repo_url ]]; then
    echo "error: you must provide a repository URL using the -r option."
    usage
    exit 1
  fi

  echo "fetch code configuration:"
  echo "  Repo URL: $repo_url"
  [ -n "$mirror_url" ] && echo "  Repo mirror: $mirror_url" && repo_url=$mirror_url
  echo "  Output path: $EXTERNAL_REPO_DIR"
  [ -n "$commit_id" ] && echo "  Commit ID: $commit_id"
  [ -n "$branch" ] && echo "  Branch: $branch"

  if [[ -n $commit_id ]]; then
    mkdir $EXTERNAL_REPO_DIR && cd "$_"
    git init -q
    git remote add origin "$repo_url"
    git fetch -q --depth 1 origin "$commit_id"
    git checkout -q "$commit_id"
    cd ..
  else
    git clone -q ${branch:+--branch $branch} --depth 1 "$repo_url" $EXTERNAL_REPO_DIR
  fi
  echo "fetch code done"
}

# common function: cleanup external repo
cleanup_external_repo() {
  echo "cleanup external repo dir: $EXTERNAL_REPO_DIR"
  rm -rf $EXTERNAL_REPO_DIR
}

# (options) common function: add // +SKIP_GO-FORM annotation
add_skip_go_from_annotation() {
  target_path=$1
  skip_annotation="// +SKIP_GO-FORM"

  echo "add annotation: $skip_annotation, configuration: "
  echo "  Target path: $target_path"

  find "$target_path" -type f -name '*.proto' -print0 | while IFS= read -r -d '' file; do
    content=$(awk '/^message/ {gsub(/^message/,"// +SKIP_GO-FORM\n&")} 1' "$file")
    echo "$content" > "$file"
  done
}


# fetch open-telemetry proto files
fetch_opentelemetry_proto() {
  echo -e "\nstart fetch open-telemetry proto files"
  target_path="./opentelemetry"

  # cleanup
  echo "cleanup opentelemetry proto files"
  rm -rf $target_path

  # fetch code
  echo -e "\n#1 [otel] fetch code:"
  fetch_code --repo=https://github.com/open-telemetry/opentelemetry-proto.git \
             --commit-id=395c8422fe90080314c7d9b4114d701a0c049e1f \
             ${OTEL_PROTO_REPO_MIRROR:+--mirror=$OTEL_PROTO_REPO_MIRROR}
  echo "#1 done"

  # mv proto files
  source_path=$EXTERNAL_REPO_DIR/opentelemetry

  echo -e "\n#2 [otel] mv proto files to erda proto dir"
  echo "  Source: $source_path"
  echo "  Target: $target_path"
  mv $source_path $target_path
  echo "#2 done"

  # rename go_package
  echo -e "\n#3 [otel] rename proto go_package"
  find "$target_path" -type f -name '*.proto' -print0 | while IFS= read -r -d '' file; do
    awk_result=$(awk '
        /^option go_package = / {
            gsub(/go\.opentelemetry\.io\/proto\/otlp/, "github.com/erda-project/erda-proto-go/opentelemetry/proto")
            gsub(/\/v1";/, "/v1/pb\";")
        }
        { print }
    ' "$file")
    echo "$awk_result" > "$file"
  done
  echo "#3 done"

  # add // +SKIP_GO-FORM annotation annotation
  echo -e "\n#4 [otel] add // +SKIP_GO-FORM annotation"
  add_skip_go_from_annotation $target_path
  echo "#4 done"

  echo -e "\nfetch open-telemetry proto files done"
}


case "$1" in
    "fetch")
        cleanup_external_repo
        # fetch open-telemetry proto files
        fetch_opentelemetry_proto
        # register your fetch function here
        ;;
    "cleanup")
        cleanup_external_repo
        ;;
    *)
        usage
esac
