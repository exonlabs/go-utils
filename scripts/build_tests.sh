#!/bin/bash
cd $(dirname $(readlink -f $0))/..

GO=go

BUILD_PATH=build/tests
if [ ! -z "${GO_BIN}" ] ;then
    GO=${GO_BIN}
    BUILD_PATH=${BUILD_PATH}/${GO_BIN}
fi
echo -e "\n* Build using $(${GO} version)\n"

BUILD_LINUX_PATH=${BUILD_PATH}/linux
BUILD_WIN_PATH=${BUILD_PATH}/win

# clean build dirs
rm -rf ${BUILD_PATH}
mkdir -m 775 -p ${BUILD_LINUX_PATH} ${BUILD_WIN_PATH}

# linux build
echo "  linux target:"
for n in gx mapx slicex fsx numx dictx ;do
    out=${BUILD_LINUX_PATH}/${n}.test
    echo "  - ${out}"
    ${GO} test ./pkg/abc/${n} -c -o ${out}
done
for n in logging events queue ciphering console jconfig ;do
    out=${BUILD_LINUX_PATH}/${n}.test
    echo "  - ${out}"
    ${GO} test ./pkg/${n} -c -o ${out}
done

# windows build
echo "  windows target:"
for n in gx mapx slicex fsx numx dictx ;do
    out=${BUILD_WIN_PATH}/${n}.test_64.exe
    echo "  - ${out}"
    GOOS=windows GOARCH=amd64 ${GO} test ./pkg/abc/${n} -c -o ${out}
    out=${BUILD_WIN_PATH}/${n}.test_32.exe
    echo "  - ${out}"
    GOOS=windows GOARCH=386 ${GO} test ./pkg/abc/${n} -c -o ${out}
done
for n in logging events queue ciphering console jconfig ;do
    out=${BUILD_WIN_PATH}/${n}.test_64.exe
    echo "  - ${out}"
    GOOS=windows GOARCH=amd64 ${GO} test ./pkg/${n} -c -o ${out}
    out=${BUILD_WIN_PATH}/${n}.test_32.exe
    echo "  - ${out}"
    GOOS=windows GOARCH=386 ${GO} test ./pkg/${n} -c -o ${out}
done

echo -e "\n* Done\n"
