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


if ${pre} || [[ "$RELEASE" == "rc"* ]]; then
	# pre-release
	ghr -u n9e -r n9e-agentd -delete --prerelease \
		-b "n9e-agentd v${VERSION}-${RELEASE} is a pre-release. It is to help gather feedback from n9e as well as give users a chance to test agentd in dev environments before v${VERSION} is officially released."\
		v${VERSION}-${RELEASE} ${dst_dir}
else
	# release
	ghr -u n9e -r n9e-agentd -delete v${VERSION}-${RELEASE} ${dst_dir}
fi
