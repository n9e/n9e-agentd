#!/bin/bash
CWD=$(cd $(dirname $0)/..; pwd)
cd $CWD

MODE=${1:-echo}

envs=()
setenv() {
	if [[ ! -z ${2} ]]; then
		eval ${1}='''${2}'''
		envs+=(${1} "${2}")
	fi
}

setenv HTTP_PROXY               ${HTTP_PROXY}
setenv HTTPS_PROXY              ${HTTPS_PROXY}

# OMNIBUS.agentd
setenv MACOSX_DEPLOYMENT_TARGET ${MACOSX_DEPLOYMENT_TARGET}
setenv SKIP_SIGN_MAC            ${SKIP_SIGN_MAC:-false}
setenv HARDENED_RUNTIME_MAC     ${HARDENED_RUNTIME_MAC:-false}
setenv OMNIBUS_BASE_DIR         ${OMNIBUS_BASE_DIR:-/opt/omnibus}
setenv OMNIBUS_GOMODCACHE       ${OMNIBUS_GOMODCACHE:-${OMNIBUS_BASE_DIR}/gomod}
setenv DEB_SIGNING_PASSPHRASE   ${DEB_SIGNING_PASSPHRASE}
setenv DEB_GPG_KEY_NAME         ${DEB_GPG_KEY_NAME}
setenv RPM_SIGNING_PASSPHRASE   ${RPM_SIGNING_PASSPHRASE}
setenv SIGN_PFX                 ${SIGN_PFX}
setenv SIGN_PFX_PW              ${SIGN_PFX_PW}
setenv SIGN_WINDOWS             ${SIGN_WINDOWS}
setenv WINDOWS_DDNPM_DRIVER     ${WINDOWS_DDNPM_DRIVER}
setenv PY_RUNTIMES              ${PY_RUNTIMES}

# GO_BUILD_LDFLAGS
setenv GO_BUILD_LDFLAGS "$($(cd $(dirname $0)/; pwd)/go-build-ldflags.sh LDFLAG)"
