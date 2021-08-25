#!/bin/bash
if [[ -z ${1} || ! -d ${1} ]]; then
	echo "Usage:"
	echo "  ${0} <etc_dir_path>"
	echo
	echo "Examples:"
	echo "  ${0} /opt/n9e/agentd/etc"
	exit 1
fi

CWD=$(cd $(dirname $1); pwd)
cd $CWD

if [[ ! -f "${CWD}/bin/n9e-agentd" ]]; then
	echo "${CWD}/bin/n9e-agentd not found"
	exit 1
fi

envs=()
setenv() {
	eval ${1}='''${2}'''
	envs+=(${1} "${2}")
}

etc_dir=${CWD}/etc
lib_dir=${CWD}/embedded/lib
env_file=${CWD}/etc/agent.env

if [[ ! -d ${etc_dir} ]]; then
	echo "dir ${etc_dir} not found"
	exit 1
fi


if [[ -d ${lib_dir} ]]; then
	setenv LD_LIBRARY_PATH ${lib_dir}
fi

if [[ -d ${CWD}/embedded/bin ]]; then
	PATH="${CWD}/embedded/bin:${PATH}"
fi

setenv PATH "${CWD}/bin:${PATH}"

file="${CWD}/etc/agentd.rc"
echo "generating ${file}"
rm -f ${file}
for (( i=0; i<${#envs[@]}; i+=2 )); do
	k=${envs[i]}
	v="${envs[(($i+1))]}"
	echo "export $k=\"${v}\"" >> ${file}
done

file="${CWD}/etc/agentd.env"
echo "generating ${file}"
rm -f ${file}
for (( i=0; i<${#envs[@]}; i+=2 )); do
	k=${envs[i]}
	v="${envs[(($i+1))]}"
	echo "$k=\"${v}\"" >> ${file}
done
