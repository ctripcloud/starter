#!/bin/bash
BRANCH=`git symbolic-ref -q --short HEAD || git describe --tags --exact-match`
HASH=`git rev-parse HEAD | cut -c 1-8`
if [[ $BRANCH =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]];then
	VERSION=$BRANCH
else
	VERSION=0.0.0-$BRANCH-$HASH
fi

docker build --build-arg BINARY=${BINARY} -t ${IMAGENAME}:$VERSION -f build/Dockerfile.release build
