#!/bin/bash

function do_merge() {
  env_branch=${1}

  # first line is base_ref
  base_ref=$(head -n 1 $dir/"$env_branch" | cut -d' ' -f1)
  if [ -z "$base_ref" ]; then
    echo "✨ No base ref found in $env_branch"
    exit 1
  fi

  # checkout to env_branch or create it
  if git rev-parse --quiet --verify "$env_branch"; then
    git checkout "$env_branch"
  else
    git checkout -b "$env_branch" "$base_ref"
  fi

  # merge all follow branch
  NEW_REFS_FILE=$(mktemp /tmp/erda-multi-merge-XXXXX)
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

  git checkout master

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

case ${instruction} in
"merge")
  do_merge "$env_branch"
  ;;
"clear")
  do_clear "$env_branch"
  ;;
esac

need_push=false
if [ "${3}" == "push" ]; then
  need_push=true
fi
echo "🪐 NEED PUSH: $need_push"

if $need_push; then
  echo "🚀 Pushing $env_branch"
  git push upstream "$env_branch:$env_branch" -f
fi
