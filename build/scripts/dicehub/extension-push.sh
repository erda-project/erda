#!/bin/bash

string=$*
array=(${string//,/ })

mkdir -p $GOPATH/src/extension

for param in ${array[@]}
do
    wget $param -O extension$count.zip
    unzip extension$count.zip -d $GOPATH/src/extension
    count=$[$count+1]
done
