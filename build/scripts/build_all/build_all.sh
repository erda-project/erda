#!/bin/bash
set -eo pipefail

echo "GO_BUILD_ENV: ${GO_BUILD_ENV}"
echo "VERSION_OPS: ${VERSION_OPS}"
echo "GO_BUILD_OPTIONS: ${GO_BUILD_OPTIONS}"
echo "PROJ_PATH: ${PROJ_PATH}"
echo "MODULE_PATH: ${MODULE_PATH:-}"

function build_one_module {
  module_path=$1
  cmd_build_dir=$2
  echo "gonna build module: $1"
  eval "${GO_BUILD_ENV}" go build "${VERSION_OPS}" "${GO_BUILD_OPTIONS}" -o "${PROJ_PATH}/$2/bin" "./$2"
}

specified_module_path="${MODULE_PATH}"

cmd_main_go_file_paths=$(find cmd -type f -name main.go)
for cmd_main_go_file_path in ${cmd_main_go_file_paths}; do # cmd/monitor/collector/main.go
  cmd_build_dir=$(dirname "${cmd_main_go_file_path}")      # cmd/monitor/collector
  module_path="${cmd_build_dir#cmd/}"                      # monitor/collector
  if [[ "${specified_module_path}" != "" ]]; then
    # only build one
    if [[ "${specified_module_path}" == "${module_path}" ]]; then
      build_one_module "${module_path}" "${cmd_build_dir}"
      break
    fi
    continue
  fi
  build_one_module "${module_path}" "${cmd_build_dir}"
done

echo "build all modules successfully!"

