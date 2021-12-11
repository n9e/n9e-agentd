name "n9e-agentd-modk"
description "mock required to the build"
default_version "1.0.0"

license "Apache-2.0"

build do
  block do
    FileUtils.mkdir_p(File.join(install_dir, "bin"))
    FileUtils.mkdir_p(File.join(install_dir, "etc", "conf.d", "load.d"))
    FileUtils.mkdir_p(File.join(install_dir, "etc", "conf.d", "io.d"))
    FileUtils.mkdir_p(File.join(install_dir, "etc", "selinux"))
    FileUtils.mkdir_p(File.join(install_dir, "scripts"))
    FileUtils.mkdir_p(File.join(install_dir, "run"))

    FileUtils.touch(File.join(install_dir, "etc", "agentd.yaml.example"))
    FileUtils.touch(File.join(install_dir, "etc", "conf.d", "io.d", "conf.yaml.default"))
    FileUtils.touch(File.join(install_dir, "etc", "conf.d", "load.d", "conf.yaml.default"))

    FileUtils.touch(File.join(install_dir, "bin", "agentd"))
    FileUtils.touch(File.join(install_dir, "etc", "selinux", "a.pp"))
    FileUtils.touch(File.join(install_dir, "scripts", "n9e-agentd"))
    FileUtils.touch(File.join(install_dir, "scripts", "n9e-agentd.conf"))
    FileUtils.touch(File.join(install_dir, "scripts", "n9e-agentd.service"))

  end
end
