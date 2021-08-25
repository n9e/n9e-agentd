# n9e-agentd

This is a monitor agent for N9E

[Download](https://github.com/n9e/n9e-agentd/releases)

## Build from source

```shell
$ make pkg

$ ls build/
linux-amd64/  n9e-agentd  n9e-agentd-5.1.0-rc1.linux.amd64.tar.gz
```

## Install & Running

Install
```shell
tar xzvf n9e-agentd-x.x.x.linux.amd64.tar.gz
mkdir -p /opt/n9e/agentd
mv ./linux-amd64/* /opt/n9e/agentd/
cp /opt/n9e/agentd/etc/agentd.yaml.default /opt/n9e/agentd/etc/agentd.yaml
```

Configure
```
## /opt/n9e/agentd/etc/agentd.yaml
agent:
  ident: $ip
  alias: $host
  endpoints:
    - http://localhost:8000			# replace the server address here
```

Start agentd
```shell
source /opt/n9e/agentd/etc/agentd.rc && \
/opt/n9e/agentd/bin/n9e-agentd start -f /opt/n9e/agentd/etc/agentd.yaml
```

Use with systemd
```shell
cp -a misc/systemd/n9e-agentd.service /usr/lib/systemd/system/
systemctl enable n9e-agentd
systemctl start n9e-agentd
```

## Run agentd in special directory

```shell
# refresh agentd.rc
cd {n9e-agentd-dir}

# generate env files
./bin/gen-envs.sh ./etc

source ./etc/agentd.rc
./bin/n9e-agentd start -f ./etc/agentd.yaml
```
