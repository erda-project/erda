#!/bin/bash
set -euo pipefail

projPath="$1"
modulePath="$2"
targetProjPath="$3"

if [[ "${modulePath}" != "" ]]; then
  echo "mode: one-app, modulePath: ${modulePath}"
  currentModuleCmdDir="${projPath}/cmd/${modulePath}"
  targetModuleCmdDir="${targetProjPath}/cmd/${modulePath}"
  targetModuleBaseDir="$(dirname "${targetModuleCmdDir}")"
  mkdir -p "${targetModuleBaseDir}"
  mv "${currentModuleCmdDir}" "${targetModuleBaseDir}"
else
  echo "mode: multi-apps"
  mkdir -p "${targetProjPath}"
  mv "${projPath}/cmd" "${targetProjPath}"
fi

# common-conf
mv "${projPath}/conf" "${targetProjPath}/conf"
