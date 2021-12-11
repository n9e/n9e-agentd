#!/bin/bash
CWD=$(cd $(dirname $0)/..; pwd)
cd $CWD

if command -v checkmodule >/dev/null 2>&1
then
	mkdir -p ./build/selinux && \
	checkmodule -M -m -o ./build/selinux/system_probe_policy.mod ./misc/selinux/system_probe_policy.te && \
	semodule_package -o ./build/selinux/system_probe_policy.pp -m ./build/selinux/system_probe_policy.mod
else
	echo "can't find command: checkmodle"
fi

