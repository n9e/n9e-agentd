#
# Copyright:: Copyright (c) 2014 Opscode, Inc.
# License:: Apache License, Version 2.0
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

#
# libyaml 0.1.5 fixes a security vulnerability to 0.1.4.
# Since the rubyinstaller.org doesn't release ruby when a dependency gets
# patched, we are manually patching the dependency until we get a new
# ruby release on windows.
# See: https://github.com/oneclick/rubyinstaller/issues/210
# This component should be removed when libyaml 0.1.5 ships with ruby builds
# of rubyinstaller.org
#
name "libyaml-windows"
default_version "0.2.2"

version "0.2.2" do
  source sha256: "9d430d3788081027a2dcf13fb8823b5ee296b1c8fe0353c86339b4c7b4018441"
end

source url: "https://s3.amazonaws.com/dd-agent-omnibus/libyaml-#{version}-x64-windows.zip",
       extract: :seven_zip

build do
  temp_directory = File.join(Omnibus::Config.cache_dir, "libyaml-cache")

  # Ensure the directory exists
  mkdir temp_directory
  # First extract the zip file
  command "7z.exe x #{project_file} -o#{temp_directory} "
  # Now copy over libyaml-0-2.dll to the build dir
  copy("#{temp_directory}/bin/libyaml-0-2.dll", "#{windows_safe_path(python_2_embedded)}/DLLs/libyaml-0-2.dll")
end
