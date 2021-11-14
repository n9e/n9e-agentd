#!/bin/bash
CWD=$(cd $(dirname $0)/..; pwd)
cd $CWD/..

APP_NAME=${APP_NAME:-n9e-agentd}

# https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
# go tool dist list
#lists="darwin/amd64 linux/amd64 linux/arm64 linux/arm windows/amd64"
lists="linux/amd64"

for str in ${lists}; do
	arr=(${str//\// })
	GOOS=${arr[0]}
	GOARCH=${arr[1]}
	mount="-v ${CWD}:/src"

	if [ -f ./build/envs ]; then
		sudo rm -f ./build/envs
	fi

	if [ -d /opt/data ]; then
		mount="${mount} -v /opt/data:/opt/data"
	fi
	echo ${mount}
	sudo docker run --rm \
		-w /src \
		${mount} \
		--name golang-cross-builder \
		--hostname golang-cross-builder \
		-e GOOS=${arr[0]} \
		-e GOARCH=${arr[1]} \
		-it ghcr.io/gythialy/golang-cross-builder:v1.16.2 \
		make pkg
done
