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

INSTALL_LOCATION="/opt/erda-quickstart"
ERDA_REPOSITORY="https://github.com/erda-project/erda.git"

ohai "Start clone Erda[${ERDA_REPOSITORY}] to ${INSTALL_LOCATION}"
(
    execute_sudo "/bin/mkdir" "-p" "${INSTALL_LOCATION}"
    cd "${INSTALL_LOCATION}" >/dev/null || return

    # we do it in four steps to avoid merge errors when reinstalling
    execute_sudo "git" "init" "-q"

    # "git remote add" will fail if the remote is defined in the global config
    execute_sudo "git" "config" "remote.origin.url" "${ERDA_REPOSITORY}"
    execute_sudo "git" "config" "remote.origin.fetch" "+refs/heads/*:refs/remotes/origin/*"

    # ensure we don't munge line endings on checkout
    execute_sudo "git" "config" "core.autocrlf" "false"

    execute_sudo "git" "fetch" "--force" "origin"
    execute_sudo "git" "fetch" "--force" "--tags" "origin"

    execute_sudo "git" "reset" "--hard" "origin/master"

    cd "${INSTALL_LOCATION}/quick-start" >/dev/null

) || exit 1

ohai "Start setup Erda using ${INSTALL_LOCATION}/quick-start/docker-compose.yml"
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
echo "visit http://erda.local to start your trivial on Erda"
echo "visit https://github.com/erda-project/erda/blob/master/docs/guides/quickstart/quickstart-full.md#try-erda for basic use of Erda"
echo "visit https://docs.erda.cloud for full introduction of Erda"
echo "goto ${INSTALL_LOCATION}/quick-start/ dir to check and manage the docker-compose resources"
