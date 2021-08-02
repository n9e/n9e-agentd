cmake_minimum_required(VERSION 3.0)
project(n9e-agentd NONE)

set(CPACK_PACKAGE_VERSION "5.0.0" 	CACHE STRING "version serial number")
set(CPACK_PACKAGE_RELEASE "1"		CACHE STRING "release serial number")
set(CPACK_GENERATOR "RPM;DEB"		CACHE STRING "package generator")
set(APP_NAME "n9e-agentd")
set(CPACK_PACKAGE_NAME ${APP_NAME})
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

include(scripts/package.cmake)

install(PROGRAMS build/n9e-agentd DESTINATION bin)
#install(PROGRAMS build/agentdctl DESTINATION /usr/bin)

install(CODE "FILE(MAKE_DIRECTORY
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/logs
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/tmp
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/run
	\$ENV{DESTDIR}/${CMAKE_INSTALL_PREFIX}/checks.d
)")

install(DIRECTORY misc/etc USE_SOURCE_PERMISSIONS DESTINATION .)
install(DIRECTORY misc/licenses USE_SOURCE_PERMISSIONS DESTINATION .)
install(DIRECTORY misc/conf.d USE_SOURCE_PERMISSIONS DESTINATION .)
install(DIRECTORY misc/scripts.d USE_SOURCE_PERMISSIONS DESTINATION .)
install(FILES misc/systemd/n9e-agentd.service DESTINATION /usr/lib/systemd/system)

include(CPack)
