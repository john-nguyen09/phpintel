#!/bin/bash

ROOT_DIR=$(pwd)

BUILD_DIR=/tmp/build_leveldb
SNAPPY_DIR=$BUILD_DIR/snappy
LEVELDB_DIR=$BUILD_DIR/leveldb
PREFIX=/mingw64

rm -rf $BUILD_DIR/*

mkdir -p $BUILD_DIR

cd $BUILD_DIR

if [ ! -f $PREFIX/lib/libsnappy.a ]; then
    (wget https://github.com/google/snappy/archive/1.1.8.zip ; \
        unzip 1.1.8.zip && \
        cd ./snappy-1.1.8 && \
        mkdir build && \
        cd build && \
        cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$PREFIX -G "MinGW Makefiles" .. && \
        make && \
        make install)
else
    echo "skip install snappy"
fi

cd $BUILD_DIR

if [ ! -f $PREFIX/lib/libleveldb.a ]; then
    (wget https://github.com/google/leveldb/archive/1.22.zip ; \
        unzip 1.22.zip && \
        cd ./leveldb-1.22 && \
        mkdir build && \
        cd build && \
        cmake -DCMAKE_INSTALL_PREFIX=$PREFIX -G "MinGW Makefiles" .. && \
        make leveldb && \
        make install)
else
    echo "skip install leveldb"
fi

cd $ROOT_DIR