#!/bin/bash
CWD=$(cd $(dirname $0)/..; pwd)
cd $CWD

MODE=${1:-echo}

envs=()
setenv() {
	eval ${1}='''${2}'''
	envs+=(${1} "${2}")
}

setenv VERSION     ${VERSION:-5.0.0}
setenv RELEASE     ${RELEASE:-1}
setenv GENERATOR   ${GENERATOR:-RPM}
setenv APP_NAME    ${APP_NAME:-n9e-agentd}
setenv GO111MODULE ${GO111MODULE:-on}
setenv CGO_ENABLED ${CGO_ENABLED:-1}
setenv GOOS        ${GOOS:-$(go env GOOS)}
setenv GOARCH      ${GOARCH:-$(go env GOARCH)}

outfile=n9e-agentd

if [[ ${GOOS} != $(go env GOHOSTOS) || ${GOARCH} != $( go env GOHOSTARCH) ]]; then
case $GOOS in
windows)
	case $GOARCH in
	amd64)
		outfile=n9e-agentd.exe
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
	amd64)
		cc=$(go env CC)
		cxx=$(go env CXX)
		;;
	*)
		echo unsupported $GOOS-$GOARCH
		exit 1
		;;
	esac
	;;
esac
fi

setenv CC      ${cc}
setenv CXX     ${cxx}
setenv OUTFILE ${CWD}/build/${outfile}

dd_root=/opt/data/${GOOS}/${GOARCH}/datadog-agent
goflags="-tags=zlib,jmx,kubelet,secrets"

if [[ -d ${dd_root}/embedded/lib ]]; then
	goflags="${goflags},python"

	if [[ ! -d "${CWD}/.cache/dd-${GOOS}-${GOARCH}" ]]; then
		mkdir -p ${CWD}/.cache
		cd ${dd_root}
		# cp dd -> .cache
		mkdir -p ${CWD}/.cache/dd-${GOOS}-${GOARCH}
		cat ${CWD}/misc/packaging/dd-filelist.txt | sudo xargs tar cf - | tar xf - -C ${CWD}/.cache/dd-${GOOS}-${GOARCH}
		cd ${CWD}/misc
		# override conf.d
		tar cf - conf.d | tar xf - -C ${CWD}/.cache/dd-${GOOS}-${GOARCH}
		cd ${CWD}
	fi

	setenv CGO_CFLAGS      "-I${dd_root}/embedded/include"
	setenv CGO_LDFLAGS     "-L${dd_root}/embedded/lib -ldl"
	setenv DD_ROOT         "${dd_root}"
	setenv LD_LIBRARY_PATH "${dd_root}/embedded/lib"
fi

setenv GOFLAGS "${goflags}"


# GO_BUILD_LDFLAGS
setenv GO_BUILD_LDFLAGS "$($(cd $(dirname $0)/; pwd)/go-build-ldflags.sh LDFLAG)"

case $MODE in
"cmake")
	for (( i=0; i<${#envs[@]}; i+=2 )); do
		k=${envs[i]}
		v=${envs[(($i+1))]}
		echo "set(${k} \"${v}\")"
	done
	;;
"echo")
	for (( i=0; i<${#envs[@]}; i+=2 )); do
		k=${envs[i]}
		v="${envs[(($i+1))]}"
		echo "$k=\"${v}\""
	done
	;;
"export")
	for (( i=0; i<${#envs[@]}; i+=2 )); do
		k=${envs[i]}
		v="${envs[(($i+1))]}"
		echo "export $k=\"${v}\""
	done
	;;
*)
	echo "Usage: $0 <cmake|echo|export>"
	exit 1
	;;
esac
