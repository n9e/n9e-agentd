require './lib/ostools.rb'
require 'pathname'

name 'n9e-agentd'

dependency "rtloader" if with_python_runtime? "3"

license "Apache-2.0"
license_file "./LICENSE"

source path: '../..'
relative_path 'src/github.com/n9e/n9e-agentd'

build do
  # set GOPATH on the omnibus source dir for this software
  gopath = Pathname.new(project_dir) + '../../../..'
  etc_dir = "/etc/n9e-agentd"
  if windows?
    env = {
        'GOPATH' => gopath.to_path,
        'PATH' => "#{gopath.to_path}/bin:#{ENV['PATH']}",
    }
    #major_version_arg = "%MAJOR_VERSION%"
    #py_runtimes_arg = "%PY_RUNTIMES%"
  else
    env = {
        'GOPATH' => gopath.to_path,
        'PATH' => "#{gopath.to_path}/bin:#{ENV['PATH']}",
        "INSTALL_DIR" => "#{install_dir}"
    }
    #major_version_arg = "$MAJOR_VERSION"
    #py_runtimes_arg = "$PY_RUNTIMES"
  end

  unless ENV["OMNIBUS_GOMODCACHE"].nil? || ENV["OMNIBUS_GOMODCACHE"].empty?
    gomodcache = Pathname.new(ENV["OMNIBUS_GOMODCACHE"])
    env["GOMODCACHE"] = gomodcache.to_path
  end

  # include embedded path (mostly for `pkg-config` binary)
  env = with_embedded_path(env)

  # we assume the go deps are already installed before running omnibus
  command "./scripts/build.sh", env: env

  conf_dir = "#{install_dir}/etc"
  mkdir conf_dir
  mkdir "#{install_dir}/bin"
  unless windows?
    mkdir "#{install_dir}/run/"
    mkdir "#{install_dir}/scripts/"
  end

  ## build the custom action library required for the install
  #if windows?
  #  platform = windows_arch_i386? ? "x86" : "x64"
  #  debug_customaction = ""
  #  if ENV['DEBUG_CUSTOMACTION'] and not ENV['DEBUG_CUSTOMACTION'].empty?
  #    debug_customaction = "--debug"
  #  end
  #  command "invoke customaction.build --major-version #{major_version_arg} #{debug_customaction} --arch=" + platform
  #  unless windows_arch_i386?
  #    command "invoke installcmd.build --major-version #{major_version_arg} --arch=" + platform
  #    command "invoke uninstallcmd.build --major-version #{major_version_arg} --arch=" + platform
  #  end
  #end

  # move around bin and config files
  copy 'misc/etc/agentd.yaml', "#{conf_dir}/agentd.yaml.example"
  copy 'misc/conf.d', "#{conf_dir}/"

  if windows? 
    copy 'build/agentd.exe', "#{install_dir}/bin/"
  else
    copy 'build/agentd', "#{install_dir}/bin/"
  end
  #unless windows?
  #  copy 'build/agentd', "#{install_dir}/bin/"
  #else
  #  copy 'bin/agent/ddtray.exe', "#{install_dir}/bin/agent"
  #  copy 'bin/agent/dist', "#{install_dir}/bin/agent"
  #  mkdir Omnibus::Config.package_dir() unless Dir.exists?(Omnibus::Config.package_dir())
  #  copy 'bin/agent/customaction.pdb', "#{Omnibus::Config.package_dir()}/"
  #end


  # Add SELinux policy for system-probe
  if debian? || redhat?
    command "./scripts/selinux.sh", env: env
    mkdir "#{conf_dir}/selinux"
    copy 'build/selinux/system_probe_policy.pp', "#{conf_dir}/selinux/"
  end

  if linux?
    if debian?
      erb source: "upstart_debian.conf.erb",
          dest: "#{install_dir}/scripts/n9e-agentd.conf",
          mode: 0644,
          vars: { install_dir: install_dir, etc_dir: etc_dir }
    elsif redhat? || suse?
      # Ship a different upstart job definition on RHEL to accommodate the old
      # version of upstart (0.6.5) that RHEL 6 provides.
      erb source: "upstart_redhat.conf.erb",
          dest: "#{install_dir}/scripts/n9e-agentd.conf",
          mode: 0644,
          vars: { install_dir: install_dir, etc_dir: etc_dir }
    end
    if suse?
      erb source: "sysvinit_suse.erb",
          dest: "#{install_dir}/scripts/n9e-agentd",
          mode: 0755,
          vars: { install_dir: install_dir, etc_dir: etc_dir }
    end

    erb source: "systemd.service.erb",
        dest: "#{install_dir}/scripts/n9e-agentd.service",
        mode: 0644,
        vars: { install_dir: install_dir, etc_dir: etc_dir }
  end

  # The file below is touched by software builds that don't put anything in the installation
  # directory (libgcc right now) so that the git_cache gets updated let's remove it from the
  # final package
  unless windows?
    delete "#{install_dir}/uselessfile"
  end
end
