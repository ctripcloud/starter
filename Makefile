mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_prj := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLE=0
export IMAGENAME=${current_prj}
export BINARY=${current_prj}

test:
	docker run --rm -w /root/${current_prj} -e GOOS=${GOOS} -e GOARCH=${GOARCH} -e CGO_ENABLE=${CGO_ENABLE} -v ${CURDIR}:/root/${current_prj} golang:1.13.4 /bin/bash -c 'GOPROXY=https://mirrors.aliyun.com/goproxy/,direct CGO_ENABLE=0 go test ./...'

bin:
	docker run --rm -w /root/${current_prj} -e GOOS=${GOOS} -e GOARCH=${GOARCH} -e CGO_ENABLE=${CGO_ENABLE} -v ${CURDIR}:/root/${current_prj}:rw golang:1.13.4 /root/${current_prj}/build/buildbin.sh

image: bin
	cp ${current_prj} build/ && build/buildimage.sh && rm build/${current_prj}
