#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
source "${ROOT_DIR}/build/scripts/git_context.sh"

tmpdir=$(mktemp -d)
trap 'rm -rf "${tmpdir}"' EXIT

main_repo="${tmpdir}/main"
repo_root="${tmpdir}/worktree"
context_root="${tmpdir}/context"
worktree_name=$(basename "${repo_root}")
mkdir -p "${main_repo}"
git -C "${tmpdir}" init main >/dev/null 2>&1
git -C "${main_repo}" config user.name test
git -C "${main_repo}" config user.email test@example.com
git -C "${main_repo}" config commit.gpgSign false

cat > "${main_repo}/README.md" <<'EOF_README'
hello
EOF_README
mkdir -p "${main_repo}/src"
cat > "${main_repo}/src/file.txt" <<'EOF_FILE'
payload
EOF_FILE

git -C "${main_repo}" add README.md src/file.txt
git -C "${main_repo}" commit -m "init" >/dev/null 2>&1
git -C "${main_repo}" worktree add "${repo_root}" -b feature >/dev/null 2>&1

prepare_git_build_context "${repo_root}" "${context_root}" "/go/src/github.com/erda-project/erda"

test -f "${context_root}/README.md"
test -f "${context_root}/src/file.txt"
test -f "${context_root}/.git"
test -f "${context_root}/.git-main/worktrees/${worktree_name}/HEAD"
test -f "${context_root}/.git-worktree/HEAD"

grep -qx 'gitdir: .git-worktree' "${context_root}/.git"
grep -qx '../.git-main' "${context_root}/.git-worktree/commondir"
grep -qx '/go/src/github.com/erda-project/erda/.git' "${context_root}/.git-worktree/gitdir"

echo "git_context_test: PASS"
