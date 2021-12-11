# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache License Version 2.0.
# This product includes software developed at Datadog (https:#www.datadoghq.com/).
# Copyright 2016-present Datadog, Inc.

name "jmxfetch"

jmxfetch_version = '0.42.0'
jmxfetch_hash = '3102eb76aa973e4ac08f9727130979d95eda734876424ca4d92b1c4c43081a1f'

if jmxfetch_version.nil? || jmxfetch_version.empty? || jmxfetch_hash.nil? || jmxfetch_hash.empty?
  raise "Please specify JMXFETCH_VERSION and JMXFETCH_HASH env vars to build."
end

default_version jmxfetch_version
source sha256: jmxfetch_hash

source url: "https://oss.sonatype.org/service/local/repositories/releases/content/com/datadoghq/jmxfetch/#{version}/jmxfetch-#{version}-jar-with-dependencies.jar",
       target_filename: "jmxfetch.jar"


jar_dir = "#{install_dir}/bin/agent/dist/jmx"

relative_path "jmxfetch"

build do
  command %q(echo > ./LICENSE <<"EOF"
BSD License

Copyright (c) 2013-2016, Datadog <info@datadoghq.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without 
modification, are permitted provided that the following conditions are met:

    * Redistributions of source code must retain the above copyright notice, 
      this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright notice,
      this list of conditions and the following disclaimer in the documentation 
      and/or other materials provided with the distribution.
    * Neither the name of Datadog nor the names of its contributors 
      may be used to endorse or promote products derived from this software 
      without specific prior written permission.
      
THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" 
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE 
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE 
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE 
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL 
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR 
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER 
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, 
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE 
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
EOF)
  ship_license "./LICENSE"
  mkdir jar_dir

  if osx? && code_signing_identity
    # Also sign binaries and libraries inside the .jar, because they're detected by the Apple notarization service.
    command "unzip jmxfetch.jar -d ."
    delete "jmxfetch.jar"

    if ENV['HARDENED_RUNTIME_MAC'] == 'true'
      hardened_runtime = "-o runtime --entitlements #{entitlements_file} "
    else
      hardened_runtime = ""
    end

    command "find . -type f | grep -E '(\\.so|\\.dylib|\\.jnilib)' | xargs -I{} codesign #{hardened_runtime}--force --timestamp --deep -s '#{code_signing_identity}' '{}'"
    command "zip jmxfetch.jar -r ."
  end

  copy "jmxfetch.jar", "#{jar_dir}/jmxfetch.jar"
  block { File.chmod(0644, "#{jar_dir}/jmxfetch.jar") }
end
