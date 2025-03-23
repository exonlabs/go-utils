#!/bin/bash
cd $(dirname $(readlink -f $0))/..

GO=go
ARGS=(-trimpath -ldflags='-s -w')

SRC_PATH=examples
BUILD_PATH=build/examples

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

    out=${BUILD_LINUX_PATH}/${name}
    echo "  - ${out}"
    ${GO} build -o ${out} "${ARGS[@]}" ${SRC_PATH}/${path}/*.go

    if [ -z "$3" ] ;then
        out=${BUILD_WIN_PATH}_64/${name}_64.exe
        echo "  - ${out}"
        GOOS=windows GOARCH=amd64 \
            ${GO} build -o ${out} "${ARGS[@]}" ${SRC_PATH}/${path}/*.go

        out=${BUILD_WIN_PATH}_32/${name}_32.exe
        echo "  - ${out}"
        GOOS=windows GOARCH=386 \
            ${GO} build -o ${out} "${ARGS[@]}" ${SRC_PATH}/${path}/*.go
    fi
}

function build_cgo {
    path=$1
    name=$2
    if [ -z "${name}" ] ;then
        name=$(echo ${path} |tr '/' '_')
    fi

    out=${BUILD_LINUX_PATH}/${name}
    echo "  - ${out}"
    CGO_ENABLED=1 \
        ${GO} build -o ${out} "${ARGS[@]}" ${SRC_PATH}/${path}/*.go

    if [ -z "$3" ] ;then
        out=${BUILD_WIN_PATH}_64/${name}_64.exe
        echo "  - ${out}"
        CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
            ${GO} build -o ${out} "${ARGS[@]}" ${SRC_PATH}/${path}/*.go

        out=${BUILD_WIN_PATH}_32/${name}_32.exe
        echo "  - ${out}"
        CC=i686-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=386 \
            ${GO} build -o ${out} "${ARGS[@]}" ${SRC_PATH}/${path}/*.go
    fi
}

# build examples

build logging/basic_loggers
build logging/custom_formatter
build logging/file_logging
build logging/tree_loggers

build jconfig/simple
build jconfig/with_backup

build proc/basic_process
build proc/routine_manager
build proc/managed_process
build proc/workers_extmanager
build proc/workers_manager "" no_win

build comm/server
build comm/client

build console/simple_inputs

build namedpipe/send_recv "" no_win


echo -e "\n* Done\n"
