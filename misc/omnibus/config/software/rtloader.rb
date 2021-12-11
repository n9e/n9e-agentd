# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache License Version 2.0.
# This product includes software developed at Datadog (https:#www.datadoghq.com/).
# Copyright 2016-present Datadog, Inc.

require './lib/ostools.rb'
require 'pathname'

name 'rtloader'

dependency "python2" if with_python_runtime? "2"
dependency "python3" if with_python_runtime? "3"

license "Apache-2.0"

source path: '../rtloader'

build do
  # set GOPATH on the omnibus source dir for this software
  env = {
      "CFLAGS" => "-I#{install_dir}/embedded/include -O2 -g -pipe",
      "LDFLAGS" => "-Wl,-rpath,#{install_dir}/embedded/lib -L#{install_dir}/embedded/lib",
  }

  command "cmake -DBUILD_DEMO=OFF -DCMAKE_INSTALL_PREFIX=#{install_dir}/embedded -DPython3_EXECUTABLE=#{install_dir}/embedded/bin/python3 -DDISABLE_PYTHON2=ON -DCMAKE_INSTALL_LIBDIR=lib .", :env => env
  command "make", :env => env
  command "make install", :env => env

end
