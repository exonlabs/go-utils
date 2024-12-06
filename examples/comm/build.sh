#!/bin/bash
cd $(dirname $(readlink -f $0))

GO=go

BUILD_PATH=../../build/comm
if [ ! -z "${GO_BIN}" ] ;then
    GO=${GO_BIN}
    BUILD_PATH=${BUILD_PATH}/${GO_BIN}
fi

# clean build dirs
rm -rf ${BUILD_PATH}
mkdir -m 775 -p ${BUILD_PATH}

files="server client"
for n in $files ;do
    # linux build
    ${GO} build -o ${BUILD_PATH}/${n} ${n}/main.go

    # windows build
    GOOS=windows GOARCH=amd64 ${GO} build \
        -o ${BUILD_PATH}/${n}_64.exe ${n}/main.go
    GOOS=windows GOARCH=386 ${GO} build \
        -o ${BUILD_PATH}/${n}_32.exe ${n}/main.go
done
