#!/bin/bash
CWD=$(cd $(dirname $0)/; pwd)
cd $CWD/..

docker run \
	--rm \
	-w /src \
	-v ${PWD}:/src \
	-e RELEASE_EXTRA=glib212 \
	-e PATH=/go/bin:/usr/bin:/bin \
	-e GOROOT=/go \
	--name golang-builder \
	--hostname glib212-builder \
	-it ybbbbasdf/centos:glib2.12 \
	make pkg
