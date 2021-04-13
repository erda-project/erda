#!/bin/bash

set -o errexit -o nounset -o pipefail
cd /tmp

if [[ -z "${NETDATA_SQLDATA_PATH}" ]]; then
    echo "no sqldata path"
    exit 1
fi

echo "$6"
s="$(echo "$6" | md5sum | awk '{print $1}')"
d="/tmp/sql-${s}"
rm -rf "${d}"
mkdir "${d}"
cd "${d}"
if [[ "$6" = file://* ]]; then
    tar -xzf "/netdata/${NETDATA_SQLDATA_PATH}/$(echo "$6" | sed 's#^file://##')"
else
    curl -o db.tar.gz "$6"
    tar -xzf db.tar.gz
fi
echo "$5" | while read db; do
    find . -name "${db}.sql" -type f | while read i; do
        echo "${i}"
        mysql -h "$1" -P "$2" -u "$3" --password="$4" "${db}" < "${i}"
    done
done
rm -rf "${d}"
