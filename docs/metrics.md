## metrics
[Datadog - Metrics Types](https://docs.datadoghq.com/developers/metrics/types/?tab=count)

## histogram

```yaml
agent:
  histogramAggregates: ["max", "median", "avg", "count"]
  histogramPercentiles: ["0.95"]
```

#### histogramAggregates
  - max
  - min
  - median
  - avg
  - sum
  - count


#### Different from open-falcon

In n9e, the meaning of count is quite different from falcon. You should use 
`monotonic_count` or `rate` to replace `count`

![metric-types](./metric-types.svg)

