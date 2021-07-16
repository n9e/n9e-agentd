## statsd

#### setup

1. config
```
agent:
  statsd:
    enabled: true
    # port - optional - default: 8125
    port: 8125
    # optional
    socket: /tmp/agent-statsd.socket
```

2. restart agentd
```
systemctl restart agentd
```


#### Code
  - [dogstatsd](https://docs.datadoghq.com/developers/dogstatsd)


#### Format

metric type
- guage `g`
- count `c`
- histogram `h`
- distribution `d`
- set `s`
- timing `ms`


```golang
type sample struct {
	Name string
	Value float64
	values []float64
	SetValue string
	MetricType metricType
	SampleRate float64
	Tags []string
}
```

#### samples
```
# service check
_sc|agent.up|0
_sc|agent.up|0|d:21|h:localhost|h:localhost2|d:22
_sc|agent.up|0|d:21|h:localhost|#tag1:test,tag2,dd.internal.entity_id:testID|m:this is fine
# event
_e{10,9}:test title|test text
_e{10,24}:test title|test\\line1\\nline2\\nline3
_e{10,24}:test|title|test\\line1\\nline2\\nline3
_e{10,9}:test title|test text|d:21
_e{10,9}:test title|test text|p:low
daemon:666|g|#
daemon:abc:def|s
daemon:3.5|d
daemon:3.5:4.5|d
# with tags
daemon:666|g|#sometag1:somevalue1,sometag2:somevalue2
# with rate
daemon:666|g|@0.21
```
