#!/usr/bin/env bash

set -euo pipefail

echoerr() { echo "$@" 1>&2; }

usage() {
  cat <<EOF
Usage:
  install.sh [--system] [stable|beta|alpha|<version>]

Examples:
  install.sh
  install.sh alpha
  install.sh 2.4.0-alpha.20260421142059
  install.sh --system stable

Environment:
  ERDA_CLI_BASE_URL      Release base URL. Default: https://erda-release.oss-cn-hangzhou.aliyuncs.com
  ERDA_CLI_INSTALL_DIR   User install directory. Default: \$HOME/.erda/bin
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echoerr "This installer needs '$1' command"
    exit 1
  fi
}

detect_os() {
  case "$(uname)" in
    Darwin)
      echo darwin
      ;;
    Linux*)
      echo linux
      ;;
    *)
      echoerr "This installer is only supported on Linux and MacOS"
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)
      echo amd64
      ;;
    arm64|aarch64)
      echo arm64
      ;;
    *)
      echoerr "Unsupported architecture: $(uname -m)"
      exit 1
      ;;
  esac
}

checksum_sha256() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    return 1
  fi
}

path_export_line() {
  local dir_expr=$INSTALL_DIR
  if [[ "$INSTALL_DIR" == "$HOME/"* ]]; then
    dir_expr="\$HOME/${INSTALL_DIR#"$HOME/"}"
  fi
  echo "export PATH=\"${dir_expr}:\$PATH\""
}

profile_append_command() {
  local profile_display=$1
  local export_line
  export_line=$(path_export_line)
  printf "grep -qxF '%s' %s 2>/dev/null || printf '%%s\\n' '' '# erda-cli' '%s' '' >> %s" "$export_line" "$profile_display" "$export_line" "$profile_display"
}

profile_path() {
  case "$(basename "${SHELL:-}")" in
  zsh)
    echo "$HOME/.zshrc"
    ;;
  bash)
    if [[ "$OS" == "darwin" ]]; then
      echo "$HOME/.bash_profile"
    else
      echo "$HOME/.bashrc"
    fi
    ;;
  esac
}

display_path() {
  local file=$1
  if [[ "$file" == "$HOME/"* ]]; then
    echo "~/${file#"$HOME/"}"
  else
    echo "$file"
  fi
}

json_field() {
  local field=$1
  sed -nE "/\"${field}\"[[:space:]]*:/ { s/.*\"${field}\"[[:space:]]*:[[:space:]]*\"([^\"]*)\".*/\1/p; q; }"
}

fetch_to_file() {
  local url=$1
  local output=$2
  local label=$3
  local status

  if ! status=$(curl -sSL -w "%{http_code}" -o "$output" "$url"); then
    echoerr "Failed to fetch ${label}: ${url}"
    exit 1
  fi

  case "$status" in
    2*)
      return 0
      ;;
    404)
      if [[ "$label" == "manifest" ]]; then
        if [[ -n "$TARGET_VERSION" ]]; then
          echoerr "No remote release found for version \"${TARGET_VERSION}\" on ${OS}/${ARCH} yet."
        else
          echoerr "No remote release found for channel \"${CHANNEL}\" on ${OS}/${ARCH} yet."
        fi
      else
        echoerr "Download package not found: ${url}"
      fi
      exit 1
      ;;
    *)
      echoerr "Failed to fetch ${label}: ${url} (status ${status})"
      exit 1
      ;;
  esac
}

SYSTEM_INSTALL=false
CHANNEL=stable
TARGET_VERSION=

while [[ $# -gt 0 ]]; do
  case "$1" in
    --system)
      SYSTEM_INSTALL=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    stable|beta|alpha)
      CHANNEL="$1"
      shift
      ;;
    *)
      if [[ -n "$TARGET_VERSION" ]]; then
        echoerr "Unexpected argument: $1"
        usage
        exit 1
      fi
      TARGET_VERSION="$1"
      shift
      ;;
  esac
done

require_cmd curl
require_cmd tar

OS=$(detect_os)
ARCH=$(detect_arch)
BASE_URL=${ERDA_CLI_BASE_URL:-https://erda-release.oss-cn-hangzhou.aliyuncs.com}

if [[ -n "$TARGET_VERSION" ]]; then
  MANIFEST_URL="${BASE_URL}/cli/${OS}/${ARCH}/erda-cli-${TARGET_VERSION}.json"
else
  MANIFEST_URL="${BASE_URL}/cli/${OS}/${ARCH}/${CHANNEL}.json"
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

MANIFEST_PATH="${TMP_DIR}/manifest.json"
fetch_to_file "$MANIFEST_URL" "$MANIFEST_PATH" "manifest"

URL=$(json_field url <"$MANIFEST_PATH")
SHA256=$(json_field sha256 <"$MANIFEST_PATH")
VERSION=$(json_field version <"$MANIFEST_PATH")

if [[ -z "$URL" ]]; then
  echoerr "Manifest does not contain download url: $MANIFEST_URL"
  exit 1
fi

if [[ -z "$VERSION" ]]; then
  echoerr "Manifest does not contain version: $MANIFEST_URL"
  exit 1
fi

if [[ "$SYSTEM_INSTALL" == "true" ]]; then
  INSTALL_DIR=/usr/local/bin
  INSTALL_PATH="${INSTALL_DIR}/erda-cli"
  SUDO=
  if [[ "$(id -u)" != "0" ]]; then
    require_cmd sudo
    SUDO=sudo
    echo "System install requires superuser permissions."
    sudo -k
  fi
else
  INSTALL_DIR="${ERDA_CLI_INSTALL_DIR:-${HOME}/.erda/bin}"
  INSTALL_PATH="${INSTALL_DIR}/erda-cli"
  SUDO=
fi

echo "Installing erda-cli ${VERSION} from $URL"

case "$URL" in
  *.tar.gz)
    ARCHIVE_PATH="${TMP_DIR}/erda-cli.tar.gz"
    ;;
  *.zip)
    ARCHIVE_PATH="${TMP_DIR}/erda-cli.zip"
    ;;
  *)
    echoerr "Unsupported archive format: $URL"
    exit 1
    ;;
esac
EXTRACTED_PATH="${TMP_DIR}/erda-cli"

fetch_to_file "$URL" "$ARCHIVE_PATH" "package"

if [[ -n "$SHA256" ]]; then
  if ! ACTUAL_SHA256=$(checksum_sha256 "$ARCHIVE_PATH"); then
    echoerr "This installer needs 'shasum' or 'sha256sum' command to verify checksum"
    exit 1
  fi
  if [[ "$ACTUAL_SHA256" != "$SHA256" ]]; then
    echoerr "Checksum mismatch for downloaded CLI"
    exit 1
  fi
fi

case "$URL" in
  *.tar.gz)
    tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"
    ;;
  *.zip)
    if ! command -v unzip >/dev/null 2>&1; then
      echoerr "This installer needs 'unzip' command to extract zip archives"
      exit 1
    fi
    unzip -q "$ARCHIVE_PATH" -d "$TMP_DIR"
    ;;
esac
if [[ ! -f "$EXTRACTED_PATH" ]]; then
  echoerr "CLI archive does not contain erda-cli"
  exit 1
fi

$SUDO mkdir -p "$INSTALL_DIR"
$SUDO mv "$EXTRACTED_PATH" "$INSTALL_PATH"
$SUDO chmod +x "$INSTALL_PATH"

echo "erda-cli installed to $INSTALL_PATH"
echo "Installed version:"
"$INSTALL_PATH" version

ACTIVE_PATH=$(command -v erda-cli 2>/dev/null || true)
echo
echo "Next step:"
if [[ "$ACTIVE_PATH" == "$INSTALL_PATH" ]]; then
  echo "  erda-cli is ready to use in this shell."
  echo "  Current erda-cli: $ACTIVE_PATH"
else
  if [[ -n "$ACTIVE_PATH" ]]; then
    echo "  Your shell currently resolves 'erda-cli' to $ACTIVE_PATH."
    echo "  Put $INSTALL_DIR before $(dirname "$ACTIVE_PATH") in PATH to use this installation."
  else
    echo "  Your shell cannot find 'erda-cli' yet."
  fi

  EXPORT_LINE=$(path_export_line)
  echo
  echo "For the current shell, run:"
  echo "  $EXPORT_LINE"

  PROFILE=$(profile_path)
  if [[ -n "$PROFILE" ]]; then
    PROFILE_DISPLAY=$(display_path "$PROFILE")
    echo
    echo "To make it persistent, run:"
    echo "  $(profile_append_command "$PROFILE_DISPLAY")"
    echo "  source $PROFILE_DISPLAY"
  fi
fi
