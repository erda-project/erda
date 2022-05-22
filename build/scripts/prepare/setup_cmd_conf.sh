#!/bin/bash
set -eo pipefail

cmdMainGoPathFromProjRoot="$1"                               # such as: cmd/monitor/monitor/main.go
cmdDirFromProjRoot="$(dirname "$cmdMainGoPathFromProjRoot")" # such as: cmd/monitor/monitor

source_dir=""
depth=$(echo "${cmdMainGoPathFromProjRoot}" | grep -o / | wc -l)
for _ in $(seq 1 "${depth}"); do
  source_dir=../${source_dir}
done
source_dir="${source_dir}conf"

target_dir="${cmdDirFromProjRoot}/common-conf"

ln_cmd="ln -shf ${source_dir} ${target_dir}"
echo "${ln_cmd}"
eval "${ln_cmd}"

cmd_conf="${cmdDirFromProjRoot}/conf"
echo "${cmd_conf}"
mkdir -p "${cmd_conf}"
touch "${cmd_conf}/.gitkeep"
