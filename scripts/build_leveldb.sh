#!/bin/bash

ROOT_DIR=$(pwd)

BUILD_DIR=/tmp/build_leveldb
SNAPPY_DIR=$BUILD_DIR/snappy
LEVELDB_DIR=$BUILD_DIR/leveldb

mkdir -p $BUILD_DIR

cd $BUILD_DIR

if [ ! -f $SNAPPY_DIR/lib/libsnappy.a ]; then
    (wget https://github.com/google/snappy/archive/1.1.8.zip ; \
        unzip 1.1.8.zip && \
        cd ./snappy-1.1.8 && \
        mkdir build && \
        cd build && \
        cmake -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ -DCMAKE_SYSTEM_NAME=Windows .. && \
        make && \
        mkdir -p $SNAPPY_DIR/include && \
        mkdir -p $SNAPPY_DIR/lib && \
        cp -f ../*.h $SNAPPY_DIR/include && \
        cp -f *.h $SNAPPY_DIR/include && \
        cp -f libsnappy.* $SNAPPY_DIR/lib && \
        cd ..)
else
    echo "skip install snappy"
fi

cd $BUILD_DIR

if [ ! -f $LEVELDB_DIR/lib/libleveldb.a ]; then
    (wget https://github.com/google/leveldb/archive/1.22.zip ; \
        unzip 1.22.zip && \
        cd ./leveldb-1.22 && \
        mkdir build && \
        cd build && \
        cmake -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ -DCMAKE_SYSTEM_NAME=Windows .. && \
        make leveldb && \
        mkdir -p $LEVELDB_DIR/include/leveldb && \
        cp -f ../include/leveldb/*.h $LEVELDB_DIR/include/leveldb && \
        mkdir -p $LEVELDB_DIR/lib && \
        cp -f libleveldb.* $LEVELDB_DIR/lib &&\
        cd ..)
else
    echo "skip install leveldb"
fi

cd $ROOT_DIR