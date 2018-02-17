#!/usr/bin/env bash

# Adapted from: https://gocv.io/getting-started/linux/

set -ex

go get -u -d gocv.io/x/gocv
cd $GOPATH/src/gocv.io/x/gocv

make deps
make download
make build
make cleanup
source ./env.sh
