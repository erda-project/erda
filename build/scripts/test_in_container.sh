#!/bin/bash

set -o errexit -o pipefail
cd $(git rev-parse --show-toplevel);

make prepare
rm -rf coverage.txt 
rm -rf go.test.sum
make run-test
mv coverage.txt /go/src/output/coverage.txt
mv go.test.sum /go/src/output/go.test.sum