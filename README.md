# n9e-agentd

This is a monitor agent for N9E

## Build from source

```shell
$ make pkg

$ ls build/
linux-amd64/  n9e-agentd  n9e-agentd-5.0.0-rc3.linux.amd64.tar.gz
```


## Install & Running

Install
```
tar xzvf n9e-agentd-X.X.X.linux.amd64.tar.gz
mkdir -p /opt/n9e
mv ./linux-amd64 /opt/n9e/agentd
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
```
/opt/n9e/agentd/bin/n9e-agentd -c /opt/n9e/agentd/etc/agentd.yaml
```

Use with systemd
```
cp -a misc/systemd/n9e-agentd.service /usr/lib/systemd/system/
systemctl enable n9e-agentd
systemctl start n9e-agentd
```
