#!/bin/bash
cd $(dirname $(readlink -f $0))/..

GO=go

SRC_PATH=./pkg
BUILD_PATH=build/tests

# GO_BIN env variable allow for building with another go binary
if [ ! -z "${GO_BIN}" ] ;then
    GO=${GO_BIN}
    BUILD_PATH=${BUILD_PATH}/${GO_BIN}
fi
echo -e "\n* Build using $(${GO} version)\n"

BUILD_LINUX_PATH=${BUILD_PATH}/linux
BUILD_WIN_PATH=${BUILD_PATH}/win

# clean build dirs
rm -rf ${BUILD_PATH}
mkdir -m 775 -p ${BUILD_LINUX_PATH} ${BUILD_WIN_PATH}_64 ${BUILD_WIN_PATH}_32

function build {
    path=$1
    name=$2
    if [ -z "${name}" ] ;then
        name=$(echo ${path} |tr '/' '_')
    fi

    out=${BUILD_LINUX_PATH}/${name}_test
    echo "  - ${out}"
    ${GO} test ${SRC_PATH}/${path} -c -o ${out}

    if [ -z "$3" ] ;then
        out=${BUILD_WIN_PATH}_64/${name}_test_64.exe
        echo "  - ${out}"
        GOOS=windows GOARCH=amd64 \
            ${GO} test ${SRC_PATH}/${path} -c -o ${out}

        out=${BUILD_WIN_PATH}_32/${name}_test_32.exe
        echo "  - ${out}"
        GOOS=windows GOARCH=386 \
            ${GO} test ${SRC_PATH}/${path} -c -o ${out}
    fi
}

# build tests

build abc/gx
build abc/mapx
build abc/slicex
build abc/fsx
build abc/numx
build abc/dictx

build logging
build events
build queue
build ciphering
build console
build jconfig


echo -e "\n* Done\n"
