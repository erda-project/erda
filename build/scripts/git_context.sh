#!/usr/bin/env bash

set -o errexit -o pipefail

is_git_worktree() {
    local repo_root="${1:?repo root is required}"
    [[ -f "${repo_root}/.git" ]] && grep -q '^gitdir: ' "${repo_root}/.git"
}

prepare_git_build_context() {
    local repo_root="${1:?repo root is required}"
    local context_root="${2:?context root is required}"
    local container_repo_path="${3:?container repo path is required}"

    rm -rf "${context_root}"
    mkdir -p "${context_root}"

    rsync -a --delete --exclude '.git' "${repo_root}/" "${context_root}/"

    if ! is_git_worktree "${repo_root}"; then
        rsync -a "${repo_root}/.git/" "${context_root}/.git/"
        return
    fi

    local worktree_git_dir common_git_dir
    worktree_git_dir="$(git -C "${repo_root}" rev-parse --git-dir)"
    common_git_dir="$(git -C "${repo_root}" rev-parse --git-common-dir)"

    rsync -a "${common_git_dir}/" "${context_root}/.git-main/"
    rsync -a "${worktree_git_dir}/" "${context_root}/.git-worktree/"

    printf 'gitdir: .git-worktree\n' > "${context_root}/.git"
    printf '../.git-main\n' > "${context_root}/.git-worktree/commondir"
    printf '%s/.git\n' "${container_repo_path}" > "${context_root}/.git-worktree/gitdir"
}
