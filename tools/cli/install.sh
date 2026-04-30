#!/bin/bash
{
    set -e
    SUDO=''
    if [ "$(id -u)" != "0" ]; then
      SUDO='sudo'
      echo "This script requires superuser permissions."
      echo "You will be prompted for password by sudo."
      # clear any previous sudo permission
      sudo -k
    fi

    $SUDO bash -s -- "$@" <<'SCRIPT'
  set -e

  echoerr() { echo "$@" 1>&2; }

  if [[ ! ":$PATH:" == *":/usr/local/bin:"* ]]; then
    echoerr '$PATH does not contain /usr/local/bin, which need by this installer.'
    exit 1
  fi

  if ! command -v jq >/dev/null 2>&1; then
    echoerr "This installer needs 'jq' command"
    exit 1
  fi
  if ! command -v curl >/dev/null 2>&1; then
    echoerr "This installer needs 'curl' command"
    exit 1
  fi

  baseURL="${ERDA_CLI_BASE_URL:-https://erda-release.oss-cn-hangzhou.aliyuncs.com}"

  # Map runtime platform into the Go-style identifiers used by manifest paths.
  if [[ "$(uname)" == "Darwin" ]]; then
    goos="darwin"
  elif [[ "$(expr substr "$(uname -s)" 1 5)" == "Linux" ]]; then
    goos="linux"
  else
    echoerr "This installer is only supported on Linux and MacOS"
    exit 1
  fi

  case "$(uname -m)" in
    x86_64|amd64) goarch="amd64" ;;
    arm64|aarch64) goarch="arm64" ;;
    *)
      echoerr "Unsupported CPU architecture: $(uname -m)"
      exit 1
      ;;
  esac

  installFromArchive() {
    local archive_url="$1"
    local expected_sha256="$2"

    archive_tmp="$(mktemp)"
    extract_tmp="$(mktemp -d)"

    cleanup() {
      rm -f "$archive_tmp"
      rm -rf "$extract_tmp"
    }
    trap cleanup RETURN

    echo "Installing CLI from $archive_url"
    curl -f -sS -L -o "$archive_tmp" "$archive_url"

    if [[ -n "$expected_sha256" && "$expected_sha256" != "null" ]]; then
      if command -v sha256sum >/dev/null 2>&1; then
        echo "${expected_sha256}  ${archive_tmp}" | sha256sum -c - >/dev/null
      else
        actual_sha256="$(shasum -a 256 "$archive_tmp" | awk '{print $1}')"
        if [[ "$actual_sha256" != "$expected_sha256" ]]; then
          echoerr "sha256 mismatch for downloaded archive"
          return 1
        fi
      fi
    fi

    case "$archive_url" in
      *.tar.gz)
        if ! command -v tar >/dev/null 2>&1; then
          echoerr "This installer needs 'tar' to extract tar.gz archives"
          return 1
        fi
        tar -xzf "$archive_tmp" -C "$extract_tmp"
        ;;
      *.zip)
        if ! command -v unzip >/dev/null 2>&1; then
          echoerr "This installer needs 'unzip' to extract zip archives"
          return 1
        fi
        unzip -q "$archive_tmp" -d "$extract_tmp"
        ;;
      *)
        echoerr "Unknown archive type: $archive_url"
        return 1
        ;;
    esac

    exe_path=""
    if [[ -f "$extract_tmp/erda-cli" ]]; then
      exe_path="$extract_tmp/erda-cli"
    elif [[ -f "$extract_tmp/erda-cli.exe" ]]; then
      exe_path="$extract_tmp/erda-cli.exe"
    fi

    if [[ -z "$exe_path" ]]; then
      echoerr "Archive does not contain erda-cli executable"
      return 1
    fi

    cp "$exe_path" /usr/local/bin/erda-cli
    chmod +x /usr/local/bin/erda-cli
    return 0
  }

  fetchLatestManifestForChannel() {
    local channel="$1"
    versions_url="${baseURL}/cli/${goos}/${goarch}/${channel}-versions.json"
    versions_tmp="$(mktemp)"

    cleanup() {
      rm -f "$versions_tmp"
    }
    trap cleanup RETURN

    http_code="$(curl -sS -o "$versions_tmp" -w "%{http_code}" "$versions_url" || echo "000")"
    if [[ "$http_code" != "200" ]]; then
      return 1
    fi

    archive_url="$(jq -r '.versions[0].url' "$versions_tmp")"
    expected_sha256="$(jq -r '.versions[0].sha256' "$versions_tmp")"
    if [[ -z "$archive_url" || "$archive_url" == "null" ]]; then
      return 1
    fi

    return 0
  }

  fetchManifestByVersion() {
    local ver="$1"
    manifest_url="${baseURL}/cli/${goos}/${goarch}/erda-cli-${ver}.json"
    manifest_tmp="$(mktemp)"

    cleanup() {
      rm -f "$manifest_tmp"
    }
    trap cleanup RETURN

    http_code="$(curl -sS -o "$manifest_tmp" -w "%{http_code}" "$manifest_url" || echo "000")"
    if [[ "$http_code" != "200" ]]; then
      return 1
    fi

    archive_url="$(jq -r .url "$manifest_tmp")"
    expected_sha256="$(jq -r .sha256 "$manifest_tmp")"
    if [[ -z "$archive_url" || "$archive_url" == "null" ]]; then
      return 1
    fi

    return 0
  }

  mode="latest"
  channel=""
  Version=""

  if [[ $# -gt 0 ]]; then
    case "$1" in
      alpha) mode="latest"; channel="alpha" ;;
      beta) mode="latest"; channel="beta" ;;
      stable) mode="latest"; channel="stable" ;;
      *)
        mode="byVersion"
        Version="$1"
        Version="${Version#v}" # Accept versions like v2.4.0.
        ;;
    esac
  fi

  # Everything is resolved from OSS manifests.
  if [[ "$mode" == "latest" ]]; then
    # When no channel is specified, try channels in order.
    channels_try=()
    if [[ -z "$channel" ]]; then
      channels_try=("stable" "beta" "alpha")
    else
      channels_try=("$channel")
    fi

    for ch in "${channels_try[@]}"; do
      if fetchLatestManifestForChannel "$ch"; then
        installFromArchive "$archive_url" "$expected_sha256"
        exit 0
      fi
    done

    if [[ -z "$channel" ]]; then
      echoerr "Failed to resolve latest CLI manifest for channels stable/beta/alpha"
    else
      echoerr "Failed to resolve latest CLI manifest for channel '$channel'"
    fi
    exit 1
  fi

  if [[ "$mode" == "byVersion" ]]; then
    if fetchManifestByVersion "$Version"; then
      installFromArchive "$archive_url" "$expected_sha256"
      exit 0
    fi
  fi

  echoerr "Failed to install CLI"
  exit 1

SCRIPT
  # test the installed CLI
  LOCATION=$(command -v erda-cli)
  echo "erda-cli installed to $LOCATION"
  erda-cli version
}