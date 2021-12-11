#
# Copyright 2012-2015 Chef Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

name "zlib"
default_version "1.2.11"

version "1.2.11" do
  source sha256: "c3e5e9fdd5004dcb542feda5ee4f0ff0744628baf8ed2dd5d66f8ca1197cb1a1"
end

version "1.2.8" do
  source md5: "44d667c142d7cda120332623eab69f40"
end
version "1.2.6" do
  source md5: "618e944d7c7cd6521551e30b32322f4a"
end

source url: "https://zlib.net/zlib-#{version}.tar.gz",
       extract: :seven_zip

relative_path "zlib-#{version}"

build do
  command %q(cat > ./LICENSE <<"EOF"
/* zlib.h -- interface of the 'zlib' general purpose compression library
  version 1.2.8, April 28th, 2013

  Copyright (C) 1995-2013 Jean-loup Gailly and Mark Adler

  This software is provided 'as-is', without any express or implied
  warranty.  In no event will the authors be held liable for any damages
  arising from the use of this software.

  Permission is granted to anyone to use this software for any purpose,
  including commercial applications, and to alter it and redistribute it
  freely, subject to the following restrictions:

  1. The origin of this software must not be misrepresented; you must not
     claim that you wrote the original software. If you use this software
     in a product, an acknowledgment in the product documentation would be
     appreciated but is not required.
  2. Altered source versions must be plainly marked as such, and must not be
     misrepresented as being the original software.
  3. This notice may not be removed or altered from any source distribution.

  Jean-loup Gailly        Mark Adler
  jloup@gzip.org          madler@alumni.caltech.edu

*/
EOF
)
  ship_license "./LICENSE"
  if windows?
    env = with_standard_compiler_flags(with_embedded_path, bfd_flags: true)

    patch source: "zlib-windows-relocate.patch", env: env

    # We can't use the top-level Makefile. Instead, the developers have made
    # an organic, artisanal, hand-crafted Makefile.gcc for us which takes a few
    # variables.
    env["BINARY_PATH"] = "/bin"
    env["LIBRARY_PATH"] = "/lib"
    env["INCLUDE_PATH"] = "/include"
    env["DESTDIR"] = "#{install_dir}/embedded"

    make_args = [
      "-fwin32/Makefile.gcc",
      "SHARED_MODE=1",
      "CFLAGS=\"#{env["CFLAGS"]} -Wall\"",
      "ASFLAGS=\"#{env["CFLAGS"]} -Wall\"",
      "LDFLAGS=\"#{env["LDFLAGS"]}\"",
      "ARFLAGS=\"rcs #{env["ARFLAGS"]}\"",
      "RCFLAGS=\"--define GCC_WINDRES #{env["RCFLAGS"]}\"",
    ]

    # On windows, msys make 3.81 doesn't support -j.
    make(*make_args, env: env)
    make("install", *make_args, env: env)
  else
    # We omit the omnibus path here because it breaks mac_os_x builds by picking
    # up the embedded libtool instead of the system libtool which the zlib
    # configure script cannot handle.
    # TODO: Do other OSes need this?  Is this strictly a mac thing?
    env = with_standard_compiler_flags
    if solaris_10?
      # For some reason zlib needs this flag on solaris (cargocult warning?)
      env["CFLAGS"] << " -DNO_VIZ"
    elsif freebsd?
      # FreeBSD 10+ gets cranky if zlib is not compiled in a
      # position-independent way.
      env["CFLAGS"] << " -fPIC"
    end

    configure env: env

    make "-j #{workers}", env: env
    make "-j #{workers} install", env: env
  end
end
