# custom
set(PREFIX ${CMAKE_INSTALL_PREFIX})
set(VERSION ${CPACK_PACKAGE_VERSION})
set(RELEASE ${CPACK_PACKAGE_RELEASE})

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
