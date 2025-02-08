#!/bin/bash
cd $(dirname $(readlink -f $0))

GO=go

BUILD_PATH=../../build/namedpipe
if [ ! -z "${GO_BIN}" ] ;then
    GO=${GO_BIN}
    BUILD_PATH=${BUILD_PATH}/${GO_BIN}
fi

# clean build dirs
rm -rf ${BUILD_PATH}
mkdir -m 775 -p ${BUILD_PATH}

files="send_recv"
for n in $files ;do
    # linux build
    ${GO} build -o ${BUILD_PATH}/${n} ${n}/main.go
done
