cmake_minimum_required(VERSION 3.0)

if(NOT EXISTS ${CMAKE_SOURCE_DIR}/scripts/${APP_NAME}.cmake)
	message( FATAL_ERROR "invaild APP_NAME ${APP_NAME}, CMake will exit." )
endif()

project(${APP_NAME} NONE)
include(${CMAKE_SOURCE_DIR}/scripts/${APP_NAME}.cmake)
