// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

/*
Package net provides core checks for networking

*/
package net

import (
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/registry/metrics"
)

var langStrings = map[string]map[string]string{
	"zh": map[string]string{
		"ntp.offset": "本地时钟与NTP参考时钟之间的时间差",
		"system.net.aws.ec2.bw_in_allowance_exceeded":     "由于入站聚合带宽超过了实例的入站聚合带宽而成的数据包数",
		"system.net.aws.ec2.bw_out_allowance_exceeded":    "数据包的数量,因为外向聚合带宽超过了实例的最大值",
		"system.net.aws.ec2.conntrack_allowance_exceeded": "数据包数,因为连接跟踪超出了实例的最大值,并且无法建立新连接",
		"system.net.aws.ec2.linklocal_allowance_exceeded": "包装的数量,因为流量的PPS到本地代理服务超过了网络接口的最大值",
		"system.net.aws.ec2.pps_allowance_exceeded":       "数据包的数量,因为Bidirectional PPS超出了实例的最大值",
		"system.net.bytes_rcvd":                           "每秒设备上收到的字节数",
		"system.net.bytes_sent":                           "从设备发送的字节数为每秒",
		"system.net.conntrack.acct":                       "boolean,启用连接跟踪流量计费。每流程64位字节和数据包计数器",
		"system.net.conntrack.buckets":                    "哈希表的大小",
		"system.net.conntrack.checksum":                   "boolean验证传入数据包的校验和",
		"system.net.conntrack.count":                      "ConnTrack表中存在的连接数",
		"system.net.conntrack.drop":                       "ConnTrack表中的跌幅数",
		"system.net.conntrack.early_drop":                 "Conntrack表中的早期跌落的数量",
		"system.net.conntrack.error":                      "Conntrack表中的错误数",
		"system.net.conntrack.events":                     "boolean启用连接跟踪代码将通过ctnetlink提供具有连接跟踪事件的用户空间",
		"system.net.conntrack.events_retry_timeout":       "events_retry_timeout",
		"system.net.conntrack.expect_max":                 "期望表的最大大小",
		"system.net.conntrack.found":                      "当前分配的流条目的数量",
		"system.net.conntrack.generic_timeout":            "默认为通用超时。这是指第4层未知/不支持的协议",
		"system.net.conntrack.helper":                     "boolean启用自动contrack辅助分配",
		"system.net.conntrack.icmp_timeout":               "默认为ICMP超时",
		"system.net.conntrack.ignore":                     "ConnTrack表中忽略的数量",
		"system.net.conntrack.insert":                     "ConnTrack表中的插入数",
		"system.net.conntrack.insert_failed":              "ConnTrack表中的插入失败的数量",
		"system.net.conntrack.invalid":                    "ConnTrack表中无效的数量",
		"system.net.conntrack.log_invalid":                "日志日志无效的数据包由值指定的类型",
		"system.net.conntrack.max":                        "Conntrack表最大容量",
		"system.net.conntrack.search_restart":             "搜索重启",
		"system.net.conntrack.tcp_be_liberal":             "boolean只能从窗口rst段标记为无效",
		"system.net.conntrack.tcp_loose":                  "boolean以启用拾取已经建立的连接",
		"system.net.conntrack.tcp_max_retrans":            "可以重新发送的最大数据包数,而无需从目的地接收（可接受的）ACK",
		"system.net.conntrack.tcp_timeout":                "TCP超时",
		"system.net.conntrack.tcp_timeout_close":          "",
		"system.net.conntrack.tcp_timeout_close_wait":     "",
		"system.net.conntrack.tcp_timeout_established":    "",
		"system.net.conntrack.tcp_timeout_fin_wait":       "",
		"system.net.conntrack.tcp_timeout_last_ack":       "",
		"system.net.conntrack.tcp_timeout_max_retrans":    "",
		"system.net.conntrack.tcp_timeout_stream":         "",
		"system.net.conntrack.tcp_timeout_syn_recv":       "",
		"system.net.conntrack.tcp_timeout_syn_sent":       "",
		"system.net.conntrack.tcp_timeout_time_wait":      "",
		"system.net.conntrack.tcp_timeout_unacknowledged": "",
		"system.net.conntrack.timestamp":                  "boolean启用连接跟踪流量时间戳",
		"system.net.packets_in.count":                     "接口接收的数据数据包数",
		"system.net.packets_in.error":                     "设备驱动程序检测到的数据包接收错误数",
		"system.net.packets_out.count":                    "接口传输的数据数据包数",
		"system.net.packets_out.error":                    "设备驱动程序检测到的数据包数量",
		"system.net.tcp.backlog_drops":                    "数据包的数量丢弃,因为TCP积压（仅限Linux）中没有空间。自代理V5.14.0以来",
		"system.net.tcp.backlog_drops.count":              "丢弃的数据包总数是因为TCP积压没有空间（仅限Linux）",
		"system.net.tcp.failed_retransmits.count":         "无法重新发送的j数据包总数（仅限Linux）",
		"system.net.tcp.in_segs":                          "收到的TCP段数（仅限Linux或Solaris）",
		"system.net.tcp.in_segs.count":                    "收到的TCP段的总数（仅限Linux或Solaris）",
		"system.net.tcp.listen_drops":                     "连接的次数已退出侦听（仅限Linux）。由于代理v5.14.0以来可用",
		"system.net.tcp.listen_drops.count":               "Connections的总次数已退出侦听（仅限Linux）",
		"system.net.tcp.listen_overflows":                 "连接的次数已溢出接受缓冲区（仅限Linux）。由于代理V5.14.0以来",
		"system.net.tcp.listen_overflows.count":           "连接的总次数已溢出接受缓冲区（仅限Linux）",
		"system.net.tcp.out_segs":                         "仅传输的TCP段数（仅限Linux或Solaris）",
		"system.net.tcp.out_segs.count":                   "仅传输的TCP段（仅限Linux或Solaris）",
		"system.net.tcp.rcv_packs":                        "收到的TCP数据包数（仅限BSD）",
		"system.net.tcp.recv_q.95percentile":              "TCP接收队列大小的第95百分位数",
		"system.net.tcp.recv_q.avg":                       "平均TCP接收队列大小",
		"system.net.tcp.recv_q.count":                     "连接速率",
		"system.net.tcp.recv_q.max":                       "最大TCP接收队列大小",
		"system.net.tcp.recv_q.median":                    "中位TCP接收队列大小",
		"system.net.tcp.retrans_packs":                    "TCP报文的数量重新发送（仅限BSD）",
		"system.net.tcp.retrans_segs":                     "TCP段的数量重传（仅限Linux或Solaris）",
		"system.net.tcp.retrans_segs.count":               "重传TCP段的总数（仅限Linux或Solaris）",
		"system.net.tcp.send_q.95percentile":              "TCP发送队列大小的第95百分位数",
		"system.net.tcp.send_q.avg":                       "平均TCP发送队列大小",
		"system.net.tcp.send_q.count":                     "连接速率",
		"system.net.tcp.send_q.max":                       "最大TCP发送队列大小",
		"system.net.tcp.send_q.median":                    "中位TCP发送队列大小",
		"system.net.tcp.sent_packs":                       "传输的TCP报文的数量（仅限BSD）",
		"system.net.tcp4.closing":                         "TCP IPv4关闭中连接的数量",
		"system.net.tcp4.established":                     "TCP IPv4建立的连接数",
		"system.net.tcp4.listening":                       "TCP IPv4监听连接的数量",
		"system.net.tcp4.opening":                         "TCP IPv4打开中连接的数量",
		"system.net.tcp6.closing":                         "TCP IPv6关闭中连接的数量",
		"system.net.tcp6.established":                     "TCP IPv6建立的连接数",
		"system.net.tcp6.listening":                       "TCP IPv6监听连接的数量",
		"system.net.tcp6.opening":                         "TCP IPv6打开中连接的数量",
		"system.net.udp.in_datagrams":                     "向UDP用户提供的UDP数据报的速率（仅限Linux）",
		"system.net.udp.in_datagrams.count":               "向UDP用户提供的UDP数据报总数（仅限Linux）",
		"system.net.udp.in_errors":                        "由于缺少目的端口（仅限Linux）缺少应用程序以外的原因无法提供的接收UDP数据报的速率",
		"system.net.udp.in_errors.count":                  "由于目的地端口缺少应用程序（仅限Linux）,无法出于缺少应用程序的原因无法提供的接收UDP数据报的总数",
		"system.net.udp.no_ports":                         "目的地端口没有应用程序的收到UDP数据报的速率（仅限Linux）",
		"system.net.udp.no_ports.count":                   "收到的UDP数据报总数在目标端口没有应用程序（仅限Linux）",
		"system.net.udp.out_datagrams":                    "从此实体发送的UDP数据报（仅限Linux）",
		"system.net.udp.out_datagrams.count":              "从此实体发送的UDP数据报总数（仅限Linux）",
		"system.net.udp.rcv_buf_errors":                   "丢失的UDP数据报速率因为接收缓冲区中没有空间（仅限Linux）",
		"system.net.udp.rcv_buf_errors.count":             "丢失的UDP数据报总数因接收缓冲区没有空间（仅限Linux）",
		"system.net.udp.snd_buf_errors":                   "udp数据报丢失的速率因为发送缓冲区中没有空间（仅限Linux）",
	},
	"en": map[string]string{
		"ntp.offset": "The time difference between the local clock and the NTP reference clock",
		"system.net.aws.ec2.bw_in_allowance_exceeded":     "The number of packets shaped because the inbound aggregate bandwidth exceeded the maximum for the instance",
		"system.net.aws.ec2.bw_out_allowance_exceeded":    "The number of packets shaped because the outbound aggregate bandwidth exceeded the maximum for the instance",
		"system.net.aws.ec2.conntrack_allowance_exceeded": "The number of packets shaped because connection tracking exceeded the maximum for the instance and new connections could not be established",
		"system.net.aws.ec2.linklocal_allowance_exceeded": "The number of packets shaped because the PPS of the traffic to local proxy services exceeded the maximum for the network interface",
		"system.net.aws.ec2.pps_allowance_exceeded":       "The number of packets shaped because the bidirectional PPS exceeded the maximum for the instance",
		"system.net.bytes_rcvd":                           "The number of bytes received on a device per second",
		"system.net.bytes_sent":                           "The number of bytes sent from a device per second",
		"system.net.conntrack.acct":                       "Boolean to enable connection tracking flow accounting. 64-bit byte and packet counters per flow are added",
		"system.net.conntrack.buckets":                    "Size of the hash table",
		"system.net.conntrack.checksum":                   "Boolean to verify checksum of incoming packets",
		"system.net.conntrack.count":                      "The number of connections present in the conntrack table",
		"system.net.conntrack.drop":                       "The number of drop in the conntrack table",
		"system.net.conntrack.early_drop":                 "The number of early drop in the conntrack table",
		"system.net.conntrack.error":                      "The number of error in the conntrack table",
		"system.net.conntrack.events":                     "Boolean to enable the connection tracking code will provide userspace with connection tracking events via ctnetlink",
		"system.net.conntrack.events_retry_timeout":       "events_retry_timeout",
		"system.net.conntrack.expect_max":                 "Maximum size of expectation table",
		"system.net.conntrack.found":                      "The number of currently allocated flow entries",
		"system.net.conntrack.generic_timeout":            "Default for generic timeout. This refers to layer 4 unknown/unsupported protocols",
		"system.net.conntrack.helper":                     "Boolean to enable automatic conntrack helper assignment",
		"system.net.conntrack.icmp_timeout":               "Default for ICMP timeout",
		"system.net.conntrack.ignore":                     "The number of ignored in the conntrack table",
		"system.net.conntrack.insert":                     "The number of insertion in the conntrack table",
		"system.net.conntrack.insert_failed":              "The number of failed insertion in the conntrack table",
		"system.net.conntrack.invalid":                    "The number of invalid in the conntrack table",
		"system.net.conntrack.log_invalid":                "Log invalid packets of a type specified by value",
		"system.net.conntrack.max":                        "Conntrack table max capacity",
		"system.net.conntrack.search_restart":             "search restart",
		"system.net.conntrack.tcp_be_liberal":             "Boolean to mark only out of window RST segments as INVALID",
		"system.net.conntrack.tcp_loose":                  "Boolean to enable picking up already established connections",
		"system.net.conntrack.tcp_max_retrans":            "Maximum number of packets that can be retransmitted without received an (acceptable) ACK from the destination",
		"system.net.conntrack.tcp_timeout":                "tcp timeout",
		"system.net.conntrack.tcp_timeout_close":          "",
		"system.net.conntrack.tcp_timeout_close_wait":     "",
		"system.net.conntrack.tcp_timeout_established":    "",
		"system.net.conntrack.tcp_timeout_fin_wait":       "",
		"system.net.conntrack.tcp_timeout_last_ack":       "",
		"system.net.conntrack.tcp_timeout_max_retrans":    "",
		"system.net.conntrack.tcp_timeout_stream":         "",
		"system.net.conntrack.tcp_timeout_syn_recv":       "",
		"system.net.conntrack.tcp_timeout_syn_sent":       "",
		"system.net.conntrack.tcp_timeout_time_wait":      "",
		"system.net.conntrack.tcp_timeout_unacknowledged": "",
		"system.net.conntrack.timestamp":                  "Boolean to enable connection tracking flow timestamping",
		"system.net.packets_in.count":                     "The number of packets of data received by the interface",
		"system.net.packets_in.error":                     "The number of packet receive errors detected by the device driver",
		"system.net.packets_out.count":                    "The number of packets of data transmitted by the interface",
		"system.net.packets_out.error":                    "The number of packet transmit errors detected by the device driver",
		"system.net.tcp.backlog_drops":                    "The number of packets dropped because there wasn't room in the TCP backlog (Linux only). Available since Agent v5.14.0",
		"system.net.tcp.backlog_drops.count":              "Total number of packets dropped because there wasn't room in the TCP backlog (Linux only)",
		"system.net.tcp.failed_retransmits.count":         "Total number of packets that failed to be retransmitted (Linux only)",
		"system.net.tcp.in_segs":                          "The number of TCP segments received (Linux or Solaris only)",
		"system.net.tcp.in_segs.count":                    "Total number of received TCP segments (Linux or Solaris only)",
		"system.net.tcp.listen_drops":                     "The number of times connections have dropped out of listen (Linux only). Available since Agent v5.14.0",
		"system.net.tcp.listen_drops.count":               "Total number of times connections have dropped out of listen (Linux only)",
		"system.net.tcp.listen_overflows":                 "The number of times connections have overflowed the accept buffer (Linux only). Available since Agent v5.14.0",
		"system.net.tcp.listen_overflows.count":           "Total number of times connections have overflowed the accept buffer (Linux only)",
		"system.net.tcp.out_segs":                         "The number of TCP segments transmitted (Linux or Solaris only)",
		"system.net.tcp.out_segs.count":                   "Total number of transmitted TCP segments (Linux or Solaris only)",
		"system.net.tcp.rcv_packs":                        "The number of TCP packets received (BSD only)",
		"system.net.tcp.recv_q.95percentile":              "The 95th percentile of the size of the TCP receive queue",
		"system.net.tcp.recv_q.avg":                       "The average TCP receive queue size",
		"system.net.tcp.recv_q.count":                     "The rate of connections",
		"system.net.tcp.recv_q.max":                       "The maximum TCP receive queue size",
		"system.net.tcp.recv_q.median":                    "The median TCP receive queue size",
		"system.net.tcp.retrans_packs":                    "The number of TCP packets retransmitted (BSD only)",
		"system.net.tcp.retrans_segs":                     "The number of TCP segments retransmitted (Linux or Solaris only)",
		"system.net.tcp.retrans_segs.count":               "Total number of retransmitted TCP segments (Linux or Solaris only)",
		"system.net.tcp.send_q.95percentile":              "The 95th percentile of the size of the TCP send queue",
		"system.net.tcp.send_q.avg":                       "The average TCP send queue size",
		"system.net.tcp.send_q.count":                     "The rate of connections",
		"system.net.tcp.send_q.max":                       "The maximum TCP send queue size",
		"system.net.tcp.send_q.median":                    "The median TCP send queue size",
		"system.net.tcp.sent_packs":                       "The number of TCP packets transmitted (BSD only)",
		"system.net.tcp4.closing":                         "The number of TCP IPv4 closing connections",
		"system.net.tcp4.established":                     "The number of TCP IPv4 established connections",
		"system.net.tcp4.listening":                       "The number of TCP IPv4 listening connections",
		"system.net.tcp4.opening":                         "The number of TCP IPv4 opening connections",
		"system.net.tcp6.closing":                         "The number of TCP IPv6 closing connections",
		"system.net.tcp6.established":                     "The number of TCP IPv6 established connections",
		"system.net.tcp6.listening":                       "The number of TCP IPv6 listening connections",
		"system.net.tcp6.opening":                         "The number of TCP IPv6 opening connections",
		"system.net.udp.in_datagrams":                     "The rate of UDP datagrams delivered to UDP users (Linux only)",
		"system.net.udp.in_datagrams.count":               "Total number of UDP datagrams delivered to UDP users (Linux only)",
		"system.net.udp.in_errors":                        "The rate of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port (Linux only)",
		"system.net.udp.in_errors.count":                  "Total number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port (Linux only)",
		"system.net.udp.no_ports":                         "The rate of received UDP datagrams for which there was no application at the destination port (Linux only)",
		"system.net.udp.no_ports.count":                   "Total number of received UDP datagrams for which there was no application at the destination port (Linux only)",
		"system.net.udp.out_datagrams":                    "The rate of UDP datagrams sent from this entity (Linux only)",
		"system.net.udp.out_datagrams.count":              "Total number of UDP datagrams sent from this entity (Linux only)",
		"system.net.udp.rcv_buf_errors":                   "The rate of UDP datagrams lost because there was no room in the receive buffer (Linux only)",
		"system.net.udp.rcv_buf_errors.count":             "Total number of UDP datagrams lost because there was no room in the receive buffer (Linux only)",
		"system.net.udp.snd_buf_errors":                   "The rate of UDP datagrams lost because there was no room in the send buffer (Linux only)",
	},
}

func init() {
	// https://docs.datadoghq.com/integrations/network/
	metrics.Register("ntp.offset", "gauge", "Shown as second")
	metrics.Register("system.net.aws.ec2.bw_in_allowance_exceeded", "gauge", "Shown as packet")
	metrics.Register("system.net.aws.ec2.bw_out_allowance_exceeded", "gauge", "Shown as packet")
	metrics.Register("system.net.aws.ec2.conntrack_allowance_exceeded", "gauge", "Shown as packet")
	metrics.Register("system.net.aws.ec2.linklocal_allowance_exceeded", "gauge", "Shown as packet")
	metrics.Register("system.net.aws.ec2.pps_allowance_exceeded", "gauge", "Shown as packet")
	metrics.Register("system.net.bytes_rcvd", "gauge", "Shown as byte")
	metrics.Register("system.net.bytes_sent", "gauge", "Shown as byte")
	metrics.Register("system.net.conntrack.acct", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.buckets", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.checksum", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.count", "gauge", "Shown as connection")
	metrics.Register("system.net.conntrack.drop", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.early_drop", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.error", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.events", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.events_retry_timeout", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.expect_max", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.found", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.generic_timeout", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.helper", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.icmp_timeout", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.ignore", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.insert", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.insert_failed", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.invalid", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.log_invalid", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.max", "gauge", "Shown as connection")
	metrics.Register("system.net.conntrack.search_restart", "count", "Shown as unit")
	metrics.Register("system.net.conntrack.tcp_be_liberal", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.tcp_loose", "gauge", "Shown as unit")
	metrics.Register("system.net.conntrack.tcp_max_retrans", "gauge", "Shown as packet")
	metrics.Register("system.net.conntrack.tcp_timeout", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_close", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_close_wait", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_established", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_fin_wait", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_last_ack", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_max_retrans", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_stream", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_syn_recv", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_syn_sent", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_time_wait", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.tcp_timeout_unacknowledged", "gauge", "Shown as second")
	metrics.Register("system.net.conntrack.timestamp", "gauge", "Shown as unit")
	metrics.Register("system.net.packets_in.count", "gauge", "Shown as packet")
	metrics.Register("system.net.packets_in.error", "gauge", "Shown as error")
	metrics.Register("system.net.packets_out.count", "gauge", "Shown as packet")
	metrics.Register("system.net.packets_out.error", "gauge", "Shown as error")
	metrics.Register("system.net.tcp.backlog_drops", "gauge", "Shown as packet")
	metrics.Register("system.net.tcp.backlog_drops.count", "count", "Shown as packet")
	metrics.Register("system.net.tcp.failed_retransmits.count", "count", "Shown as packet")
	metrics.Register("system.net.tcp.in_segs", "gauge", "Shown as segment")
	metrics.Register("system.net.tcp.in_segs.count", "count", "Shown as segment")
	metrics.Register("system.net.tcp.listen_drops", "gauge")
	metrics.Register("system.net.tcp.listen_drops.count", "count")
	metrics.Register("system.net.tcp.listen_overflows", "gauge")
	metrics.Register("system.net.tcp.listen_overflows.count", "count")
	metrics.Register("system.net.tcp.out_segs", "gauge", "Shown as segment")
	metrics.Register("system.net.tcp.out_segs.count", "count", "Shown as segment")
	metrics.Register("system.net.tcp.rcv_packs", "gauge", "Shown as packet")
	metrics.Register("system.net.tcp.recv_q.95percentile", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.recv_q.avg", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.recv_q.count", "rate", "Shown as connection")
	metrics.Register("system.net.tcp.recv_q.max", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.recv_q.median", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.retrans_packs", "gauge", "Shown as packet")
	metrics.Register("system.net.tcp.retrans_segs", "gauge", "Shown as segment")
	metrics.Register("system.net.tcp.retrans_segs.count", "count", "Shown as segment")
	metrics.Register("system.net.tcp.send_q.95percentile", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.send_q.avg", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.send_q.count", "rate", "Shown as connection")
	metrics.Register("system.net.tcp.send_q.max", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.send_q.median", "gauge", "Shown as byte")
	metrics.Register("system.net.tcp.sent_packs", "gauge", "Shown as packet")
	metrics.Register("system.net.tcp4.closing", "gauge", "Shown as connection")
	metrics.Register("system.net.tcp4.established", "gauge", "Shown as connection")
	metrics.Register("system.net.tcp4.listening", "gauge", "Shown as connection")
	metrics.Register("system.net.tcp4.opening", "gauge", "Shown as connection")
	metrics.Register("system.net.tcp6.closing", "gauge", "Shown as connection")
	metrics.Register("system.net.tcp6.established", "gauge", "Shown as connection")
	metrics.Register("system.net.tcp6.listening", "gauge", "Shown as connection")
	metrics.Register("system.net.tcp6.opening", "gauge", "Shown as connection")
	metrics.Register("system.net.udp.in_datagrams", "gauge", "Shown as datagram")
	metrics.Register("system.net.udp.in_datagrams.count", "count", "Shown as datagram")
	metrics.Register("system.net.udp.in_errors", "gauge", "Shown as datagram")
	metrics.Register("system.net.udp.in_errors.count", "count", "Shown as datagram")
	metrics.Register("system.net.udp.no_ports", "gauge", "Shown as datagram")
	metrics.Register("system.net.udp.no_ports.count", "count", "Shown as datagram")
	metrics.Register("system.net.udp.out_datagrams", "gauge", "Shown as datagram")
	metrics.Register("system.net.udp.out_datagrams.count", "count", "Shown as datagram")
	metrics.Register("system.net.udp.rcv_buf_errors", "gauge", "Shown as error")
	metrics.Register("system.net.udp.rcv_buf_errors.count", "count", "Shown as error")
	metrics.Register("system.net.udp.snd_buf_errors", "gauge", "Shown as error")

	i18n.SetLangStrings(langStrings)
}
