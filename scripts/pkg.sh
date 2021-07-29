#!/bin/bash
CWD=$(cd $(dirname $0)/; pwd)
cd $CWD/..

make || exit 1

cd build

os=$(go env GOOS)
arch=$(go env GOARCH)
file=n9e-agentd
dir="${os}-${arch}"
test -d ${dir} || mkdir -p ${dir}/{bin,misc,run,checks.d}
cp -a ../README.md ${dir}/
cp -a ../misc/etc ${dir}/
cp -a ../misc/conf.d ${dir}/
cp -a ../misc/scripts.d ${dir}/
cp -a ../misc/systemd ${dir}/misc/
cp -a ../build/n9e-agentd ${dir}/bin/

if [[ -n ${RELEASE_EXTRA} ]]; then
	tar czvf n9e-agentd-${VERSION}-${RELEASE}-${RELEASE_EXTRA}.${os}.${arch}.tar.gz ${dir}
else
	tar czvf n9e-agentd-${VERSION}-${RELEASE}.${os}.${arch}.tar.gz ${dir}
fi
