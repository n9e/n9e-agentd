#!/bin/bash
CWD=$(cd $(dirname $0)/; pwd)
cd $CWD/..

# https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
#lists="darwin/amd64 linux/amd64 linux/arm64 windows/amd64"
lists="darwin/amd64 linux/amd64 windows/amd64"
#lists="linux/arm64"


unset GOFLAGS
set -aex
cd build
for str in ${lists}; do
	arr=(${str//\// })
	os=${arr[0]}
	arch=${arr[1]}
	dir="${os}-${arch}"
	file=agentd
	cc=$(go env CC)
	cxx=$(go env CXX)

	if [[ ${os} == "windows" ]]; then
		file=agentd.exe
		cc=x86_64-w64-mingw32-gcc
		cxx=x86_64-w64-mingw32-g++
	fi

	if [[ ${os} == "darwin" ]]; then
		cc=o64-clang
		cxx=o64-clang++
	fi

	test -d ${dir} || mkdir -p ${dir}/{bin,misc}
	GO111MODULE=on CGO_ENABLED=1 GOOS=${os} GOARCH=${arch} \
		CC=${cc} CXX=${cxx} \
		go build -ldflags "${GO_BUILD_LDFLAGS}" \
		-mod vendor \
		-o ${dir}/bin/${file} ../cmd/agentd
	cp -a ../README.md ${dir}/
	cp -a ../misc/etc ${dir}/
	cp -a ../misc/conf.d ${dir}/
	cp -a ../misc/script.d ${dir}/
	cp -a ../misc/systemd ${dir}/misc/
	tar czvf n9e-agentd-${VERSION}-${RELEASE}.${os}.${arch}.tar.gz ${dir}
done
