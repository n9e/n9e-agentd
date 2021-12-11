require "./lib/ostools.rb"

name 'openssl'
package_name 'openssl'

build_version "v1.0.0"
install_dir '/opt/n9e-agentd'
dependency 'openssl'
