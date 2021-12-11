name "procps-ng"
default_version "3.3.16"

ship_source true

source url:    "https://gitlab.com/procps-ng/procps/-/archive/v3.3.16/procps-v#{version}.tar.gz",
       sha256: "7f09945e73beac5b12e163a7ee4cae98bcdd9a505163b6a060756f462907ebbc"

relative_path "procps-v#{version}"

env = {
  "LDFLAGS" => "-L#{install_dir}/embedded/lib -I#{install_dir}/embedded/include",
  "CFLAGS" => "-L#{install_dir}/embedded/lib -I#{install_dir}/embedded/include",
  "LD_RUN_PATH" => "#{install_dir}/embedded/lib",
}

build do
  ship_license "./COPYING"

  # By default procps-ng will build with the 'UNKNOWN' version if not built
  # from a git repository and the '.tarball-version' file doesn't exist.
  # Setting the version in that file will allow binaries to return the correct
  # info from the '--version' command.
  File.open(".tarball-version", "w") do |f|
    f.puts "#{version}"
  end

  command("./autogen.sh", env: env)
  command(["./configure",
           "--prefix=#{install_dir}/embedded",
           "--without-ncurses",
           ""].join(" "),
    env: env)
  command "make -j #{workers}", env: { "LD_RUN_PATH" => "#{install_dir}/embedded/lib" }
  command "make install"
end
