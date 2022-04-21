#!/bin/bash

function do_merge() {
  dir=${1}
  env_branch=${2}
  need_push=${3}

  current_branch=$(git branch --show-current)
  if ! git diff-index --quiet HEAD --; then
    echo "You have uncommitted changes in your current branch. Please commit or stash them before continuing."
    exit 1
  fi

  # first line is base_ref
  base_ref=$(head -n 1 "$dir/$env_branch" | cut -d' ' -f1)
  if [ -z "$base_ref" ]; then
    echo "✨ No base ref found in $env_branch"
    exit 1
  fi

  # checkout to env_branch or create it
  fetch_branch "$env_branch"
  if ! git rev-parse --quiet --verify "$env_branch"; then
    git checkout -b "$env_branch" "$base_ref"
  else
    git checkout "$env_branch"
  fi

  # merge all follow branch
  NEW_REFS_FILE=$(mktemp /tmp/erdamerge.XXXXXX)
  while IFS= read -r line
  do
    if [ "$line" == "---" ]; then
      echo "---" >> "$NEW_REFS_FILE"
      continue
    fi

    to_merge_ref=$(echo "$line" | cut -d' ' -f1)

    fetch_branch "$to_merge_ref"
    git rev-parse --quiet --verify "$to_merge_ref"
    if [ $? -ne 0 ]; then
      echo "🔗 Merge ref $to_merge_ref not found"
      echo "$to_merge_ref [NOT FOUND]" >> "$NEW_REFS_FILE"
    else
      echo "✨ Merging $to_merge_ref into $env_branch"
      git merge "$to_merge_ref"
      if [ $? -ne 0 ]; then
        echo "✨ Merge $to_merge_ref failed"
        echo "$to_merge_ref [FAIL]" >> "$NEW_REFS_FILE"
        git merge --abort
        continue
      fi
      echo "$to_merge_ref [OK]" >> "$NEW_REFS_FILE"
    fi
  done < "$dir/$env_branch"

  mv "$NEW_REFS_FILE" "$dir/$env_branch"


  if $need_push; then
    echo "🚀 Pushing $env_branch"
    git push upstream "$env_branch:$env_branch" -f

    echo "🚀 Commit & Pushing config_info"
    pushd "$dir"
    if ! git diff-index --quiet HEAD --; then
      git commit -a -m "Update deploy-info" && \
      git push upstream deploy-info:deploy-info
    fi
    popd
  fi

  git checkout "$current_branch"

}

function fetch_branch() {
  ref=$1
  git fetch upstream "$ref:$ref" -f -u || true
}

function do_clear() {
  env_branch=${1}
  echo "✨ Clearing $env_branch"
  git checkout master
  git branch -D "$env_branch" || true
}

dir=./deploy/auto-heads
instruction=${1}
if [[ "${instruction}" != "merge" && "${instruction}" != "clear" ]]; then
  echo "Usage: $0 [ merge | clear ] <branch>"
  exit 1
fi
env_branch=${2}
if [ -z "${env_branch}" ]; then
  echo "Usage: $0 [ merge | clear ] <branch>"
  exit 1
fi
need_push=false
if [ "${3}" == "push" ]; then
  need_push=true
fi
echo "🪐 NEED PUSH: $need_push"


case ${instruction} in
"merge")
  do_merge $dir "$env_branch" $need_push
  ;;
"clear")
  do_clear "$env_branch"
  ;;
esac
