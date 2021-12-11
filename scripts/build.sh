#!/bin/bash
CWD=$(cd $(dirname $0)/..; pwd)
cd $CWD

PACKAGE_VERSION=${PACKAGE_VERSION:-5.0.0}
PACKAGE_RELEASE=${PACKAGE_RELEASE:-1}
GOMODCACHE=${GOMODCACHE:-$(go env GOMODCACHE)}
GO111MODULE=${GO111MODULE:-on}
CGO_ENABLED=${CGO_ENABLED:-1}
GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}
GOPATH=${GOPATH:-$(go env GOPATH)}
INSTALL_DIR=${INSTALL_DIR:-./build/dist}
GO_BUILD_LDFLAGS="$($(cd $(dirname $0)/; pwd)/go-build-ldflags.sh LDFLAG)"
GOFLAGS="-tags=zlib,jmx,kubelet,secrets"

if [[ -d ${INSTALL_DIR}/embedded/lib ]]; then
	CGO_CFLAGS="-I${INSTALL_DIR}/embedded/include"
	CGO_LDFLAGS="-L${INSTALL_DIR}/embedded/lib -ldl"
	LDFLAGS="-Wl,-rpath,${INSTALL_DIR}/embedded/lib -L${INSTALL_DIR}/embedded/lib"
fi

if [[ ! -z ${PY_RUNTIMES} ]]; then
	GOFLAGS="${GOFLAGS},python"
fi

outfile=agentd

if [[ ${GOOS} != $(go env GOHOSTOS) || ${GOARCH} != $( go env GOHOSTARCH) ]]; then
case $GOOS in
windows)
	case $GOARCH in
	amd64)
		outfile=agentd.exe
		cc=x86_64-w64-mingw32-gcc
		cxx=x86_64-w64-mingw32-g++
		;;
	*)
		echo unsupported $GOOS-$GOARCH
		exit 1
		;;
	esac
	;;
darwin)
	case $GOARCH in
	amd64)
		cc=o64-clang
		cxx=o64-clang++
		;;
	arm64)
		cc=arm64-apple-darwin20.2-clang
		cxx=arm64-apple-darwin20.2-clang++
		;;
	*)
		echo unsupported $GOOS-$GOARCH
		exit 1
		;;
	esac
	;;
linux)
	case $GOARCH in
	arm)
		cc=arm-linux-gnueabi-gcc
		cxx=arm-linux-gnueabi-g++
		;;
	arm64)
		cc=aarch64-linux-gnu-gcc
		cxx=aarch64-linux-gnu-g++
		;;
	*)
		echo unsupported $GOOS-$GOARCH
		exit 1
		;;
	esac
	;;
esac
fi

OUTFILE=${CWD}/build/${outfile}


## prepare
if [[ -f ${OUTFILE} ]]; then
	rm -f ${OUTFILE} || exit 1
fi

if [[ -z ${GO_BUILD_LDFLAGS} || -z ${OUTFILE} ]]; then
	echo "unable to get ${GO_BUILD_LDFLAGS} or ${OUTFILE}"
	exit 1
fi

# check version
GIT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo unknown)

if [[ ${GIT_VERSION} != "v${PACKAGE_VERSION}-${PACKAGE_RELEASE}" ]]; then
	echo GIT_VERSION ${GIT_VERSION}
	echo Version ${PACKAGE_VERSION}
	echo Release ${PACKAGE_RELEASE}
	echo "Git version is out of sync. Please release the version first"
	echo "git tag v${PACKAGE_VERSION}-${PACKAGE_RELEASE}"
	exit 1
fi




envs="GOMODCACHE=\"${GOMODCACHE}\""
envs="${envs} GO111MODULE=\"${GO111MODULE}\""
envs="${envs} CGO_ENABLED=\"${CGO_ENABLED}\""
envs="${envs} GOOS=\"${GOOS}\""
envs="${envs} GOARCH=\"${GOARCH}\""
envs="${envs} GOPATH=\"${GOPATH}\""
envs="${envs} GOFLAGS=\"${GOFLAGS}\""

if [[ ! -z "$cc" ]]; then
	envs="${envs} cc=\"${cc}\""
fi

if [[ ! -z "$cxx" ]]; then
	envs="${envs} cxx=\"${cxx}\""
fi

cmd="${envs} go build -ldflags \"${GO_BUILD_LDFLAGS}\" -o ${OUTFILE} ./cmd/agent"

set -x
eval $cmd
set +x

if [ $? -ne 0 ];then
	exit 1
fi

if [[ $(go env GOHOSTOS) = $(go env GOOS) && $(go env GOARCH) = $(go env GOHOSTARCH) ]]; then
	${OUTFILE} version
fi
