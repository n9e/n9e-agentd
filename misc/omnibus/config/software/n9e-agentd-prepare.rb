name "n9e-agentd-prepare"
description "steps required to preprare the build"
default_version "1.0.0"

license "Apache-2.0"

build do
  block do
    %w{embedded/lib embedded/bin bin}.each do |dir|
      dir_fullpath = File.expand_path(File.join(install_dir, dir))
      FileUtils.mkdir_p(dir_fullpath)
      FileUtils.touch(File.join(dir_fullpath, ".gitkeep"))
    end
  end
end

if windows?
  build do
    block do
      FileUtils.mkdir_p(File.expand_path(File.join(Omnibus::Config.source_dir(), "n9e-agentd", "src", "github.com", "n9e", "n9e-agentd", "bin", "agentd")))
      FileUtils.mkdir_p(File.expand_path(File.join(install_dir, "bin", "agentd")))

    end
  end
end
