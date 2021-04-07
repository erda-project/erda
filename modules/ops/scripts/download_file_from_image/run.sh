#!/bin/bash

set -o errexit -o nounset -o pipefail
source env.sh

if [[ -z "${NETDATA_AGENT_PATH}" ]]; then
    echo "no agent path"
    exit 1
fi

# args:
# IMAGE_NAME      镜像名
# IMAGE_FILE_PATH 镜像内文件路径
# FILE_NAME       复制到外部的文件名
# EXPECT_MD5      文件 MD5
# EXECUTABLE      文件是否可执行

tempFileName=tmpfile-$RANDOM
destPath="/netdata/${NETDATA_AGENT_PATH}/${FILE_NAME}"

# destPath 可能被误创建为目录或其他非 Regular File
if [[ ! -f "${destPath}" ]]; then
    rm -fr "${destPath}"
    mkdir -p $(dirname "${destPath}")
fi

# 如果目标文件已存在
if [[ -f "$destPath" ]]; then
    currentMD5="$(md5sum ${destPath} | cut -d ' ' -f1)"
    if [[ "${EXPECT_MD5}" == "${currentMD5}" ]]; then
        echo "current file MD5 matched, exit directly."
        exit 0
    fi
    echo "MD5 not match, expect: ${EXPECT_MD5} current: ${currentMD5}"
fi

echo "begin download file to ${destPath}"
tmpContainerID=$(docker create "${IMAGE_NAME}")
docker cp "${tmpContainerID}:${IMAGE_FILE_PATH}" ${tempFileName}
docker rm ${tmpContainerID}
mv -f "${tempFileName}" "${destPath}"

if [[ "${EXECUTABLE}" == "true" ]]; then
    chmod +x "${destPath}"
fi
