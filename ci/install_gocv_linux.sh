#!/usr/bin/env bash

# Adapted from: https://gocv.io/getting-started/linux/

set -ex

go get -u -d gocv.io/x/gocv
cd $GOPATH/src/gocv.io/x/gocv

make deps

# NOTE: We can't use `make download` and `make build` as described
# in the tutorial since they have hard-coded /tmp paths, and on Travis
# /tmp is a small tmpfs without enough room.
TMPDIR=/home/travis/opencv

# make download
mkdir -p $TMPDIR
cd $TMPDIR
wget -q -O opencv.zip https://github.com/opencv/opencv/archive/3.4.0.zip
unzip -q opencv.zip
wget -q -O opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/3.4.0.zip
unzip -q opencv_contrib.zip

# make build
cd $TMPDIR/opencv-3.4.0
mkdir build
cd build
cmake -D CMAKE_BUILD_TYPE=RELEASE -D CMAKE_INSTALL_PREFIX=/usr/local -D OPENCV_EXTRA_MODULES_PATH=$TMPDIR/opencv_contrib-3.4.0/modules -D BUILD_DOCS=OFF BUILD_EXAMPLES=OFF -D BUILD_TESTS=OFF -D BUILD_PERF_TESTS=OFF -D BUILD_opencv_java=OFF -D BUILD_opencv_python=OFF -D BUILD_opencv_python2=OFF -D BUILD_opencv_python3=OFF ..
make -j4
sudo make install
sudo ldconfig
