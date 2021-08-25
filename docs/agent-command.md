# Howto use agentd as command line

To use command line, you must enable the apiserver module with agentd

## Configure

Add apiserver section into agentd.yaml

```yaml
# /opt/n9e/agentd/agentd.yaml
apiserver:
  enabled: true
  address: 127.0.0.1
  port: 8010
```

restart agentd
```
systemctl restart n9e-agentd
```

## Integration

```
source /opt/n9e/agentd/etc/agentd.rc

# list integration
agent integration list

# list integration
agent integration info [package]

# install integration
agent integration install [package]

# remove integration
agent integration remove [package]
```


## Check

Take `process` as an example

Configure

```yaml
## /opt/n9e/agentd/conf.d/process.d/conf.yaml
init_config:

instances:

 ## @param name - string - required
 ## Used to uniquely identify your metrics as they are tagged with this name in Datadog.
 #
   - name: ssh

 ## @param search_string - list of strings - optional
 ## If one of the elements in the list matches, it returns the count of
 ## all the processes that match the string exactly by default. Change this behavior with the
 ## parameter `exact_match: false`.
 ##
 ## Note: One and only one of search_string, pid or pid_file must be specified per instance.
 #
 search_string:
   - ssh
   - sshd
```

Check

```
source /opt/n9e/agentd/etc/agentd.rc
agent check process
```

#### See Also
 - https://docs.datadoghq.com/integrations/process/
