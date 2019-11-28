#!/bin/bash
BRANCH=`git symbolic-ref -q --short HEAD || git describe --tags --exact-match`
HASH=`git rev-parse HEAD | cut -c 1-8`
GOVER=`go version`
BUILDTIME=`date +%FT%T%z`
if [[ $BRANCH =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]];then
	VERSION=$BRANCH-$HASH
else
	VERSION=0.0.0-$BRANCH-$HASH
fi
MODULE=`GOPROXY="https://mirrors.aliyun.com/goproxy/,direct" go list -m | tr -d '\n'`
GOPROXY="https://mirrors.aliyun.com/goproxy/,direct" go build -ldflags "-X '${MODULE}/pkg.Version=$VERSION' -X '${MODULE}/pkg.GoVersion=$GOVER' -X '${MODULE}/pkg.BuildTime=$BUILDTIME'" .
