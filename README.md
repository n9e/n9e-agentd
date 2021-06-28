# n9e-agentd

This is a monitor agent for N9E

#### build from source

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


#### install

generic install
```
mkdir -p /opt/n9e/agentd/{bin,run,logs}
cp -a build/agentd /opt/n9e/agentd/bin/
cp -a misc/* /opt/n9e/agentd/
cp /opt/n9e/agentd/etc/agentd.yaml.default /opt/n9e/agentd/etc/agentd.yaml
```

systemd
```
cp -a misc/systemd/n9e-agentd.service /usr/lib/systemd/system/
systemctl enable n9e-agentd
systemctl start n9e-agentd
```


#### run

```
/opt/n9e/agentd/bin/agentd -c /opt/n9e/agentd/etc/agentd.yaml
```

