#!/bin/bash

string=$*
array=(${string//,/ })

for param in ${array[@]}
do
    wget $param -O extension$count.zip
    unzip extension$count.zip -d /tmp/dicehub-extension
    count=$[$count+1]
done
