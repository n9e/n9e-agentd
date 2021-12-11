#!/bin/bash

docker run --rm -it \
	--name="n9e-agentd-builder" \
	--hostname="n9e-agentd-builder" \
	-v "$PWD:/src/n9e-agentd" \
	-v "$PWD/../tmp/omnibus:/opt/omnibus" \
	-v "$PWD/../tmp/n9e-agentd:/opt/n9e-agentd" \
	--workdir=/src/n9e-agentd dd:rpm-x64 bash

