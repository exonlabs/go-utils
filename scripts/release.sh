#!/bin/bash
cd $(dirname $(readlink -f $0))/..

VER_FILE=version

VERSION=$(cat ${VER_FILE} |head -n 1 |xargs |sed 's|\.dev.*||g')
RELEASE_TAG=v${VERSION}


echo -e "\n* Releasing: ${VERSION}"

# check previous versions tags
if git tag |grep -wq "${RELEASE_TAG}" ;then
    echo -e "\n-- Error!! tag '${RELEASE_TAG}' already exists\n"
    exit 1
fi

# adjust release version
echo -e "${VERSION}" > ${VER_FILE}

# setting release tag
git commit -m "Release '${VERSION}'" ${VER_FILE}
if ! git tag "${RELEASE_TAG}" ;then
    echo -e "\n-- Error!! failed commit and adding tag '${RELEASE_TAG}'\n"
    exit 1
fi

# bump new version
NEW_VER=$(echo "${VERSION}" \
    |awk -F. '{for(i=1;i<NF;i++){printf $i"."}{printf $NF+1".dev"}}')
echo -e "${NEW_VER}" > ${VER_FILE}
git commit -m "Bump version to '${NEW_VER}'" ${VER_FILE}

echo -e "\n* Released: ${VERSION}\n"
