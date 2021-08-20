cmake_minimum_required(VERSION 3.0)

include(${CMAKE_SOURCE_DIR}/build/envs.cmake)

project(${APP_NAME} NONE)

set(CPACK_PACKAGE_NAME ${APP_NAME})
set(CPACK_PACKAGE_VERSION ${VERSION})
set(CPACK_PACKAGE_RELEASE ${RELEASE})
set(CPACK_GENERATOR ${GENERATOR})
set(CPACK_PACKAGE_CONTACT "steveyubo@didichuxing.com")
set(CPACK_PACKAGE_VENDOR "https://n9e.didiyun.com/")
set(CPACK_RPM_PACKAGE_URL "https://github.com/n9e/n9e-agentd")
set(CPACK_PACKAGE_DESCRIPTION "")
set(CPACK_PACKAGE_FILE_NAME "${CPACK_PACKAGE_NAME}-${CPACK_PACKAGE_VERSION}-${CPACK_PACKAGE_RELEASE}.${CMAKE_SYSTEM_NAME}.${CMAKE_SYSTEM_PROCESSOR}")
set(CPACK_RPM_PACKAGE_LICENSE "BSD-3")
set(CPACK_RPM_PACKAGE_REQUIRES "")
set(CPACK_DEBIAN_PACKAGE_DEPENDS "libc6 (>= 2.7)")
set(APP_USER root)
set(APP_GROUP root)
set(APP_UID 0)
set(APP_GID 0)
set(DD_CACHE_DIR "${CMAKE_SOURCE_DIR}/.cache/dd-${GOOS}-${GOARCH}")
set(CMAKE_INSTALL_PREFIX /opt/n9e/agentd)
set(CPACK_RPM_PACKAGE_GROUP "System Environment/Daemons")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "N9E Agent Deamon")
set(CPACK_RPM_EXCLUDE_FROM_AUTO_FILELIST_ADDITION 
	/opt
	/opt/n9e
	/opt/n9e/agentd
	/usr
	/usr/lib
	/usr/lib/systemd
	/usr/lib/systemd/system)

set(PREFIX ${CMAKE_INSTALL_PREFIX})

set(CPACK_PACKAGE_RELOCATABLE "")
set(CPACK_RESOURCE_FILE_LICENSE ${CMAKE_SOURCE_DIR}/LICENSE)
set(CPACK_RESOURCE_FILE_README ${CMAKE_SOURCE_DIR}/README.md)
set(CPACK_PACKAGING_INSTALL_PREFIX ${CMAKE_INSTALL_PREFIX})

set(CPACK_SYSTEM_NAME ${CMAKE_SYSTEM_NAME})
set(CPACK_RPM_PACKAGE_SUMMARY ${CPACK_PACKAGE_NAME})
set(CPACK_RPM_PACKAGE_DESCRIPTION ${CPACK_PACKAGE_DESCRIPTION_SUMMARY})
# This prevents the default %post from running which causes binaries to be
# striped. Without this, MaxCtrl will not work on all systems as the
# binaries will be stripped.
set(CPACK_RPM_SPEC_INSTALL_POST "/bin/true")
set(CPACK_RPM_PACKAGE_RELEASE ${CPACK_PACKAGE_RELEASE})
set(CPACK_RPM_PACKAGE_ARCHITECTURE ${CMAKE_SYSTEM_PROCESSOR})
set(CPACK_RPM_PRE_UNINSTALL_SCRIPT_FILE ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/rpm/prerm)
set(CPACK_RPM_POST_INSTALL_SCRIPT_FILE ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/rpm/postinst)
set(CPACK_RPM_POST_UNINSTALL_SCRIPT_FILE ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/rpm/postrm)

if(CMAKE_SYSTEM_PROCESSOR STREQUAL "x86_64")
	set(CPACK_DEBIAN_PACKAGE_ARCHITECTURE "amd64")
else()
	set(CPACK_DEBIAN_PACKAGE_ARCHITECTURE ${CMAKE_SYSTEM_PROCESSOR})
endif()
set(CPACK_DEBIAN_PACKAGE_DESCRIPTION "${CPACK_PACKAGE_DESCRIPTION}\n ${CPACK_PACKAGE_DESCRIPTION_SUMMARY}")
set(CPACK_DEBIAN_PACKAGE_CONTROL_EXTRA "${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/deb/postinst;${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/deb/prerm;${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/deb/postrm")

foreach(action "after_install" "after_remove" "after_upgrade" "before_remove")
	set(out "${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/${action}")
	if(EXISTS ${CMAKE_SOURCE_DIR}/misc/${APP_NAME}/${action}.sh.in)
		configure_file(${CMAKE_SOURCE_DIR}/misc/${APP_NAME}/${action}.sh.in ${out})
	elseif(${CMAKE_SOURCE_DIR}/misc/packaging/${action}.sh.in)
		configure_file(${CMAKE_SOURCE_DIR}/misc/packaging/${action}.sh.in ${out})
	endif()
endforeach()

if(EXISTS ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/after_install)
	file(READ ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/after_install AFTER_INSTALL_SCRIPT)
endif()
if(EXISTS ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/after_remove)
	file(READ ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/after_remove AFTER_REMOVE_SCRIPT)
endif()
if(EXISTS ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/after_upgrade)
	file(READ ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/after_upgrade AFTER_UPGRADE_SCRIPT)
endif()
if(EXISTS ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/before_remove)
	file(READ ${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/before_remove BEFORE_REMOVE_SCRIPT)
endif()

foreach(action "postinst" "prerm" "postrm")
	foreach(pkg "rpm" "deb")
		configure_file(${CMAKE_SOURCE_DIR}/misc/packaging/${pkg}/${action}.sh.in
			${CMAKE_CURRENT_BINARY_DIR}/packaging/${APP_NAME}/${pkg}/${action})
	endforeach(pkg)
endforeach(action)

install(CODE "FILE(MAKE_DIRECTORY
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/logs
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/tmp
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/run
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/checks.d
)")

install(FILES build/n9e-agentd misc/gen-envs.sh DESTINATION bin)
install(DIRECTORY misc/etc USE_SOURCE_PERMISSIONS DESTINATION .)
install(DIRECTORY misc/licenses USE_SOURCE_PERMISSIONS DESTINATION .)
install(DIRECTORY misc/scripts.d USE_SOURCE_PERMISSIONS DESTINATION .)
install(FILES misc/systemd/n9e-agentd.service DESTINATION /usr/lib/systemd/system)

if (DEFINED DD_ROOT)
	install(DIRECTORY ${DD_CACHE_DIR}/conf.d USE_SOURCE_PERMISSIONS DESTINATION .)
	install(DIRECTORY ${DD_CACHE_DIR}/embedded USE_SOURCE_PERMISSIONS DESTINATION .)
	install(FILES
		${DD_CACHE_DIR}/agent_requirements-py3.txt
		${DD_CACHE_DIR}/final_constraints-py3.txt
		${DD_CACHE_DIR}/requirements-agent-release.txt
		${DD_CACHE_DIR}/version-manifest.json
		DESTINATION .)
else()
	install(DIRECTORY misc/conf.d USE_SOURCE_PERMISSIONS DESTINATION .)
endif()

include(CPack)

