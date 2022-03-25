#!/bin/bash
set -e

home=${home}
workspace=${workspace}

if [[ ${workspace} =~ ^"${home}".* ]];
then
    ws="/root${workspace#$home}"

    cd $ws
    /usr/bin/erda-cli $@
else
    echo error: workspace \'${workspace}\' is not in the user home \'${home}\'
    exit 1
fi
