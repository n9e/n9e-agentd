#!/bin/bash
CWD=$(cd $(dirname $0)/..; pwd)
cd $CWD

MODE=${1:-build}

## prepare
if [[ -f ${OUTFILE} ]]; then
	rm -f ${OUTFILE} || exit 1
fi

if [[ -z ${GO_BUILD_LDFLAGS} || -z ${OUTFILE} ]]; then
	echo "unable to get ${GO_BUILD_LDFLAGS} or ${OUTFILE}"
	exit 1
fi

go build -ldflags "${GO_BUILD_LDFLAGS}" -o ${OUTFILE} ./cmd/agent || exit 1

if [[ $(go env GOHOSTOS) = $(go env GOOS) && $(go env GOARCH) = $(go env GOHOSTARCH) ]]; then
	${OUTFILE} version
fi

if [ ${MODE} != "pkg" ]; then
	exit 0
fi

# build package
cd build

dir="${CWD}/build/${GOOS}-${GOARCH}"
test -d ${dir} || mkdir -p ${dir}/{bin,misc,run,checks.d,logs,tmp}
cp -a ${OUTFILE} ${dir}/bin/
cp -a ../misc/bin/* ${dir}/bin/
cp -a ../README.md ${dir}/
cp -a ../misc/etc ${dir}/
cp -a ../misc/conf.d ${dir}/
cp -a ../misc/scripts.d ${dir}/
cp -a ../misc/licenses ${dir}/
cp -a ../misc/systemd ${dir}/misc/
cp -a ../scripts/gen-envs.sh ${dir}/bin/

if [[ -n ${DD_ROOT} && -d ${CWD}/.cache/dd-${GOOS}-${GOARCH} ]]; then
	cd ${CWD}/.cache/dd-${GOOS}-${GOARCH}
	tar cf - * | tar xf - -C ${dir}
fi

pkg_name=n9e-agentd-${VERSION}-${RELEASE}.${GOOS}.${GOARCH}.tar.gz
if [[ -n ${RELEASE_EXTRA} ]]; then
	pkg_name=${pkg_name/%.tar.gz/.${RELEASE_EXTRA}.tar.gz}
fi

cd ${CWD}/build
tar czvf $pkg_name ${GOOS}-${GOARCH}
