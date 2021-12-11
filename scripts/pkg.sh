#!/bin/bash
CWD=$(cd $(dirname $0)/..; pwd)
cd $CWD

. ./scripts/build.sh

# build package
cd build

dir="${CWD}/build/${GOOS}-${GOARCH}"
test -d ${dir} || mkdir -p ${dir}/{bin,dist,misc,run,checks.d,logs,tmp}
cp -a ${OUTFILE} ${dir}/bin/
cp -a ../misc/bin/* ${dir}/bin/
cp -a ../README.md ${dir}/
cp -a ../misc/etc ${dir}/
cp -a ../misc/conf.d ${dir}/
cp -a ../misc/scripts.d ${dir}/

pkg_name=n9e-agentd-${PACKAGE_VERSION}-${PACKAGE_RELEASE}.${GOOS}.${GOARCH}.tar.gz
if [[ -n ${PACKAGE_RELEASE_EXTRA} ]]; then
	pkg_name=${pkg_name/%.tar.gz/.${PACKAGE_RELEASE_EXTRA}.tar.gz}
fi

cd ${CWD}/build
tar czvf $pkg_name ${GOOS}-${GOARCH}
