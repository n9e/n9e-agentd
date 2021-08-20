// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/log"
	coreConfig "github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/config/logs"
)

// LogsConfigKeys stores logs configuration keys stored in YAML configuration files
type LogsConfigKeys struct {
	prefix string
	config *coreConfig.Config
	c      *logs.Config
}

// defaultLogsConfigKeys defines the default YAML keys used to retrieve logs configuration
func defaultLogsConfigKeys() *LogsConfigKeys {
	return NewLogsConfigKeys("logs_config.", coreConfig.C)
}

// NewLogsConfigKeys returns a new logs configuration keys set
func NewLogsConfigKeys(configPrefix string, config *coreConfig.Config) *LogsConfigKeys {
	path := fmt.Sprintf("agent.%s", strings.Trim(configPrefix, "."))
	c := &logs.Config{}
	if err := config.Read(path, c); err != nil {
		log.Warnf("read logs_config at %s error %s", path, err)
	}

	return &LogsConfigKeys{prefix: path, config: config, c: c}
}

func (l *LogsConfigKeys) getConfig() *coreConfig.Config {
	if l.config != nil {
		return l.config
	}
	return coreConfig.C
}

func (l *LogsConfigKeys) getConfigKey(key string) string {
	return l.prefix + key
}

//func isSetAndNotEmpty(config coreConfig.Config, key string) bool {
//	return config.IsSet(key) && len(config.GetString(key)) > 0
//}

//func (l *LogsConfigKeys) isSetAndNotEmpty(key string) bool {
//	return isSetAndNotEmpty(l.getConfig(), key)
//}

func (l *LogsConfigKeys) ddURL443() string {
	return l.c.Url443
}

func (l *LogsConfigKeys) logsDDURL() (string, bool) {
	return l.c.LogsUrl, l.c.LogsUrl != ""
}

func (l *LogsConfigKeys) ddPort() int {
	return l.c.DDPort
}

func (l *LogsConfigKeys) isSocks5ProxySet() bool {
	return len(l.socks5ProxyAddress()) > 0
}

func (l *LogsConfigKeys) socks5ProxyAddress() string {
	return l.c.Socks5ProxyAddress
}

func (l *LogsConfigKeys) isForceTCPUse() bool {
	return l.c.UseTCP
}

func (l *LogsConfigKeys) usePort443() bool {
	return l.c.UsePort443
}

func (l *LogsConfigKeys) isForceHTTPUse() bool {
	return l.c.UseHTTP
}

func (l *LogsConfigKeys) logsNoSSL() bool {
	return !l.c.UseSSL
}

func (l *LogsConfigKeys) devModeNoSSL() bool {
	return l.c.DevModeNoSSL
}

func (l *LogsConfigKeys) devModeUseProto() bool {
	return l.c.DevModeUseProto
}

func (l *LogsConfigKeys) compressionLevel() int {
	return l.c.CompressionLevel
}

func (l *LogsConfigKeys) useCompression() bool {
	return l.c.UseCompression
}

func (l *LogsConfigKeys) hasAdditionalEndpoints() bool {
	return len(l.getAdditionalEndpoints()) > 0
}

// getLogsAPIKey provides the dd api key used by the main logs agent sender.
func (l *LogsConfigKeys) getLogsAPIKey() string {
	return l.c.APIKey
}

func (l *LogsConfigKeys) connectionResetInterval() time.Duration {
	return l.c.ConnectionResetInterval
}

func (l *LogsConfigKeys) getAdditionalEndpoints() []logs.Endpoint {
	return l.c.AdditionalEndpoints
}

func (l *LogsConfigKeys) expectedTagsDuration() time.Duration {
	return l.c.ExpectedTagsDuration
}

func (l *LogsConfigKeys) taggerWarmupDuration() time.Duration {
	return l.c.TaggerWarmupDuration
}

func (l *LogsConfigKeys) batchWait() time.Duration {
	return l.c.BatchWait
}

func (l *LogsConfigKeys) batchMaxConcurrentSend() int {
	return l.c.BatchMaxConcurrentSend
}

func (l *LogsConfigKeys) batchMaxSize() int {
	return l.c.BatchMaxSize
}

func (l *LogsConfigKeys) batchMaxContentSize() int {
	return l.c.BatchMaxContentSize
}

func (l *LogsConfigKeys) senderBackoffFactor() float64 {
	return l.c.SenderBackoffFactor
}

func (l *LogsConfigKeys) senderBackoffBase() float64 {
	return l.c.SenderBackoffBase
}

func (l *LogsConfigKeys) senderBackoffMax() float64 {
	return l.c.SenderBackoffMax
}

func (l *LogsConfigKeys) senderRecoveryInterval() int {
	return l.c.SenderRecoveryInterval
}

func (l *LogsConfigKeys) senderRecoveryReset() bool {
	return l.c.SenderRecoveryReset
}

// AggregationTimeout is used when performing aggregation operations
func (l *LogsConfigKeys) aggregationTimeout() time.Duration {
	return l.c.AggregationTimeout
}

func (l *LogsConfigKeys) useV2API() bool {
	return l.c.UseV2Api
}
