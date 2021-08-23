#!/bin/bash

set -u

abort() {
  printf "%s\n" "$@"
  exit 1
}

if [ -z "${BASH_VERSION:-}" ]; then
  abort "Bash is required to interpret this script."
fi

# First check OS.
OS="$(uname)"
if [[ "$OS" == "Linux" ]]; then
  ON_LINUX=1
elif [[ "$OS" == "Darwin" ]]; then
  ON_MACOS=1
else
  abort "The script is only supported on macOS and Linux."
fi

if [[ -z "${ON_LINUX-}" ]]; then
  PERMISSION_FORMAT="%A"
  CHOWN="/usr/sbin/chown"
  CHGRP="/usr/bin/chgrp"
  GROUP="admin"
else
  PERMISSION_FORMAT="%a"
  CHOWN="/bin/chown"
  CHGRP="/bin/chgrp"
  GROUP="$(id -gn)"
fi

# string formatters
if [[ -t 1 ]]; then
  tty_escape() { printf "\033[%sm" "$1"; }
else
  tty_escape() { :; }
fi
tty_mkbold() { tty_escape "1;$1"; }
tty_underline="$(tty_escape "4;39")"
tty_blue="$(tty_mkbold 34)"
tty_red="$(tty_mkbold 31)"
tty_bold="$(tty_mkbold 39)"
tty_reset="$(tty_escape 0)"

have_sudo_access() {
  local -a args
  if [[ -n "${SUDO_ASKPASS-}" ]]; then
    args=("-A")
  elif [[ -n "${NONINTERACTIVE-}" ]]; then
    args=("-n")
  fi

  if [[ -z "${HAVE_SUDO_ACCESS-}" ]]; then
    if [[ -n "${args[*]-}" ]]; then
      SUDO="/usr/bin/sudo ${args[*]}"
    else
      SUDO="/usr/bin/sudo"
    fi
    if [[ -n "${NONINTERACTIVE-}" ]]; then
      ${SUDO} -l mkdir &>/dev/null
    else
      ${SUDO} -v && ${SUDO} -l mkdir &>/dev/null
    fi
    HAVE_SUDO_ACCESS="$?"
  fi

  if [[ -n "${ON_MACOS-}" ]] && [[ "$HAVE_SUDO_ACCESS" -ne 0 ]]; then
    abort "Need sudo access on macOS (e.g. the user $USER needs to be an Administrator)!"
  fi

  return "$HAVE_SUDO_ACCESS"
}

shell_join() {
  local arg
  printf "%s" "$1"
  shift
  for arg in "$@"; do
    printf " "
    printf "%s" "${arg// /\ }"
  done
}

ohai() {
  printf "${tty_blue}==>${tty_bold} %s${tty_reset}\n" "$(shell_join "$@")"
}

execute() {
  if ! "$@"; then
    abort "$(printf "Failed during: %s" "$(shell_join "$@")")"
  fi
}

execute_sudo() {
  local -a args=("$@")
  if have_sudo_access; then
    if [[ -n "${SUDO_ASKPASS-}" ]]; then
      args=("-A" "${args[@]}")
    fi
    ohai "/usr/bin/sudo" "${args[@]}"
    execute "/usr/bin/sudo" "${args[@]}"
  else
    ohai "${args[@]}"
    execute "${args[@]}"
  fi
}

# USER isn't always set so provide a fall back for the installer and subprocesses.
if [[ -z "${USER-}" ]]; then
  USER="$(chomp "$(id -un)")"
  export USER
fi

# Invalidate sudo timestamp before exiting (if it wasn't active before).
if ! /usr/bin/sudo -n -v 2>/dev/null; then
  trap '/usr/bin/sudo -k' EXIT
fi

# Things can fail later if `pwd` doesn't exist.
# Also sudo prints a warning message for no good reason
cd "/usr" || exit 1

######################################################script

if ! command -v git >/dev/null; then
    abort "$(cat <<EOABORT
You must install Git before installing Erda.
EOABORT
)"
fi

if ! command -v docker >/dev/null; then
    abort "$(cat <<EOABORT
You must install Docker before installing Erda.
EOABORT
)"
fi

if ! command -v docker-compose >/dev/null; then
    abort "$(cat <<EOABORT
You must install Docker-Compose before installing Erda.
EOABORT
)"
fi

if ! command -v netcat >/dev/null; then
    abort "$(cat <<EOABORT
You must install netcat before installing Erda.
EOABORT
)"
fi

INSTALL_LOCATION="/opt/erda-quickstart"
ERDA_REPOSITORY="https://github.com/erda-project/erda.git"
ERDA_VERSION="23936e9"

# shellcheck disable=SC2016
ohai 'Checking for `sudo` access (which may request your password).'
have_sudo_access

# chown
if ! [[ -d "${INSTALL_LOCATION}" ]]; then
  execute_sudo "/bin/mkdir" "-p" "${INSTALL_LOCATION}"
fi
execute_sudo "$CHOWN" "-R" "$USER:$GROUP" "${INSTALL_LOCATION}"

ohai "Start clone Erda[${ERDA_REPOSITORY}] to ${INSTALL_LOCATION}"
(
  cd "${INSTALL_LOCATION}" >/dev/null || return

  # we do it in four steps to avoid merge errors when reinstalling
  execute "git" "init" "-q"

  # "git remote add" will fail if the remote is defined in the global config
  execute "git" "config" "remote.origin.url" "${ERDA_REPOSITORY}"
  execute "git" "config" "remote.origin.fetch" "+refs/heads/*:refs/remotes/origin/*"

  # ensure we don't munge line endings on checkout
  execute "git" "config" "core.autocrlf" "false"

  execute "git" "fetch" "--force" "origin"
  execute "git" "fetch" "--force" "--tags" "origin"

  execute "git" "reset" "--hard" "${ERDA_VERSION}"

) || exit 1

ohai "Start setup Erda using ${INSTALL_LOCATION}/quick-start/docker-compose.yml"

cd "${INSTALL_LOCATION}/quick-start" || exit 1
execute "docker-compose" "up" "-d" "mysql" || exit 1

echo "waiting for mysql ready"
sleep 10
i=1
until nc -z localhost 3306
do
  sleep 10
  if ((i++ >= 100)); then
    echo "timeout waiting for mysql ready"
    exit 1
  fi
done

execute "docker-compose" "up" "erda-migration" || exit 1
execute "docker-compose" "up" "sysctl-init" || exit 1
execute "docker-compose" "up" "-d" "elasticsearch" || exit 1
execute "docker-compose" "up" "-d" "cassandra" || exit 1
execute "docker-compose" "up" "-d" "kafka" || exit 1
execute "docker-compose" "up" "-d" || exit 1

ohai "Setup local hosts"
(
  exists_in_host="$(grep -n erda.local /etc/hosts)"
  if [ -z "$exists_in_host" ]; then
    echo "127.0.0.1 erda.local one.erda.local" | execute_sudo "tee" "-a" "/etc/hosts"
  fi
) || exit 1

ohai "Erda has been started successfully using ${INSTALL_LOCATION}/quick-start/docker-compose.yml"

ohai "Next steps:"
echo "visit ${tty_underline}http://erda.local${tty_reset} to start your journey on Erda"
echo "visit ${tty_underline}https://github.com/erda-project/erda/blob/master/docs/guides/quickstart/quickstart-full.md#try-erda${tty_reset} for basic use of Erda"
echo "visit ${tty_underline}https://docs.erda.cloud${tty_reset} for full introduction of Erda"
echo "goto ${INSTALL_LOCATION}/quick-start/ dir to check and manage the docker-compose resources"
