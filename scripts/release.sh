#!/bin/bash
cd $(dirname $(readlink -f $0))/..

COMMIT_FILE=go.mod

VERSION=$(grep 'version = ' ${COMMIT_FILE} \
    |head -n 1 |cut -d'"' -f2 |xargs |sed 's|\.dev.*||g')
RELEASE_TAG=v${VERSION}

NEW_VER=$(echo "${VERSION}" \
    |awk -F. '{for(i=1;i<NF;i++){printf $i"."}{printf $NF+1".dev"}}')


echo -e "\n* Releasing: ${RELEASE_TAG}"

# check previous versions tags
if git tag |grep -wq "${RELEASE_TAG}" ;then
    echo -e "\n-- Error!! tag '${RELEASE_TAG}' already exists\n"
    exit 1
fi

# adjust release version
sed -i "s|version = .*|version = \"${VERSION}\"|g" ${COMMIT_FILE}

# setting release tag
git commit -m "Release '${VERSION}'" ${COMMIT_FILE}
if ! git tag "${RELEASE_TAG}" ;then
    echo -e "\n-- Error!! failed adding tag '${RELEASE_TAG}'\n"
    exit 1
fi

# bump new version
sed -i "s|version = .*|version = \"${NEW_VER}\"|g" ${COMMIT_FILE}
git commit -m "Bump version to '${NEW_VER}'" ${COMMIT_FILE}

echo -e "\n* Released: ${VERSION}\n"
