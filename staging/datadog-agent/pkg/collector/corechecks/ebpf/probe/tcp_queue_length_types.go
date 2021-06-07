package probe

// TCPQueueLengthStatsKey is the type of the `TCPQueueLengthStats` map key: the container ID
type TCPQueueLengthStatsKey struct {
	ContainerID string `json:"containerid"`
}

// TCPQueueLengthStatsValue is the type of the `TCPQueueLengthStats` map value: the maximum fill rate of busiest read and write buffers
type TCPQueueLengthStatsValue struct {
	ReadBufferMaxUsage  uint32 `json:"readBufferMaxUsage"`
	WriteBufferMaxUsage uint32 `json:"writeBufferMaxUsage"`
}

// TCPQueueLengthStats is the map of the maximum fill rate of the read and write buffers per container
type TCPQueueLengthStats map[string]TCPQueueLengthStatsValue
