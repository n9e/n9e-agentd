#!/bin/bash -x

src_dir=$(pwd)

install_dir=/tmp/rtloader
if [[ -d $install_dir ]]; then
	rm -rf $install_dir
fi
mkdir $install_dir

build_dir=/tmp/build
if [[ -d $build_dir ]]; then
	rm -rf $build_dir
fi
mkdir $build_dir

cd $build_dir

export LDFLAGS="-Wl,-rpath,/opt/n9e-agentd/embedded/lib -L/opt/n9e-agentd/embedded/lib"
export LD_LIBRARY_PATH="/opt/n9e-agentd/embedded/lib"
cmake -DCMAKE_INSTALL_LIBDIR=lib -DBUILD_DEMO=OFF -DCMAKE_INSTALL_PREFIX:PATH=${install_dir}/embedded -DPython3_EXECUTABLE=/root/miniconda3/bin/python3.9 -DDISABLE_PYTHON2=ON -DCMAKE_INSTALL_LIBDIR=lib ${src_dir}
make
make install

unset LDFLAGS
unset LD_LIBRARY_PATH

#ls ${install_dir}/embedded/lib
ldd ${install_dir}/embedded/lib/libdatadog-agent-three.so | grep not
