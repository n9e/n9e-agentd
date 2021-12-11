#!/bin/bash
CWD=$(cd $(dirname $0)/; pwd)
cd $CWD/..

pre=false
dst_dir=./build/release

if [[ -d ${dst_dir} ]]; then
	rm -rf ${dst_dir}
fi

ls build/*.tar.gz >/dev/null 2>&1 
if [[ $? != 0 ]]; then
	echo no tar.gz packages found
	exit 1
fi

mkdir -p ${dst_dir}
cp -a  build/*.tar.gz ${dst_dir}
cp -a  build/*.rpm ${dst_dir}


reg='^[0-9]+$'
if ${pre} || [[ ! $PACKAGE_RELEASE =~ $reg ]]; then
	# pre-release
	ghr -u n9e -r n9e-agentd -delete --prerelease \
		-b "n9e-agentd v${PACKAGE_VERSION}-${PACKAGE_RELEASE} is a pre-release. It is to help gather feedback from n9e as well as give users a chance to test agentd in dev environments before v${PACKAGE_VERSION} is officially released."\
		v${PACKAGE_VERSION}-${PACKAGE_RELEASE} ${dst_dir}
else
	# release
	ghr -u n9e -r n9e-agentd -delete v${PACKAGE_VERSION}-${PACKAGE_RELEASE} ${dst_dir}
fi

