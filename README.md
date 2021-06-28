# n9e-agentd

This is a monitor agent for N9E

## Build from source

```
make
```

Cross compiling
```shell
$ make pkgs

$ ls build/
agentd*        linux-amd64/  n9e-agentd-5.0.0-rc1.darwin.amd64.tar.gz  n9e-agentd-5.0.0-rc1.windows.amd64.tar.gz
darwin-amd64/  mocker*       n9e-agentd-5.0.0-rc1.linux.amd64.tar.gz   windows-amd64/
```


## Install & Running

Install
```
mkdir -p /opt/n9e/agentd/{bin,run,logs}
cp -a build/agentd /opt/n9e/agentd/bin/
cp -a misc/* /opt/n9e/agentd/ 
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
  configProviders:
    - name: http
      polling: true
      templateUrl: http://localhost:8000	# replace the server address here
```

Start agentd
```
/opt/n9e/agentd/bin/agentd -c /opt/n9e/agentd/etc/agentd.yaml
```

Use with systemd
```
cp -a misc/systemd/n9e-agentd.service /usr/lib/systemd/system/
systemctl enable n9e-agentd
systemctl start n9e-agentd
```

## Configure
```
# /opt/n9e/agentd/etc/agentd.yaml

```
