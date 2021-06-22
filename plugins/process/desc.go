package process

import (
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/registry/metrics"
)

var langStrings = map[string]map[string]string{
	"zh": map[string]string{
		"proc.num":                "进程数量",
		"proc.uptime":             "进程运行时间",
		"proc.createtime":         "进程启动时间",
		"proc.open_fd_count":      "进程文件句柄数量",
		"proc.mem.rss":            "进程常驻内存大小",
		"proc.mem.vms":            "进程虚拟内存大小",
		"proc.mem.swap":           "进程交换空间大小",
		"proc.mem.shared":         "进程共享内存大小",
		"proc.mem.text":           "进程Text内存大小",
		"proc.mem.lib":            "进程lib内存大小",
		"proc.mem.data":           "进程data内存大小",
		"proc.mem.dirty":          "进程dirty内存大小",
		"proc.cpu.total":          "进程cpu使用率",
		"proc.cpu.user":           "进程用户态cpu使用率",
		"proc.cpu.sys":            "进程系统态cpu使用率",
		"proc.cpu.threads":        "进程中线程数量",
		"proc.io.read_rate":       "进程io读取频率(hz)",
		"proc.io.write_rate":      "进程io写入频率(hz)",
		"proc.io.readbytes_rate":  "进程io读取速率(b/s)",
		"proc.io.writebytes_rate": "进程io写入速率(b/s)",
		"proc.net.conn_rate":      "进程网络连接频率(hz)",
		"proc.net.bytes_rate":     "进程网络传输率(b/s)",
	},
	"en": map[string]string{
		"proc.num":                "The number of the process",
		"proc.uptime":             "The uptime of the process(values in seconds)",
		"proc.createtime":         "The start time of the process",
		"proc.open_fd_count":      "The count of open file descriptors for the process",
		"proc.mem.rss":            "Resident set size (RSS) is the portion of memory occupied by a process that is held in main memory (RAM)",
		"proc.mem.vms":            "Virtual memory size",
		"proc.mem.swap":           "Swap space size",
		"proc.mem.shared":         "Shared memory size",
		"proc.mem.text":           "Text memory size",
		"proc.mem.lib":            "Lib memory size",
		"proc.mem.data":           "Data memory size",
		"proc.mem.dirty":          "Dirty meomry size",
		"proc.cpu.total":          "Total CPU usage in percentage",
		"proc.cpu.user":           "User CPU usage in percentage",
		"proc.cpu.sys":            "System CPU usage in percentage",
		"proc.cpu.threads":        "The number of threads in the process",
		"proc.io.read_rate":       "The rate of I/O read(hz)",
		"proc.io.write_rate":      "The rate of I/O write(hz)",
		"proc.io.readbytes_rate":  "The bytes rate of I/O read(b/s)",
		"proc.io.writebytes_rate": "The bytes rate of I/O write(b/s)",
		"proc.net.conn_rate":      "The rate of network connection(hz)",
		"proc.net.bytes_rate":     "The bytes rate of network connection(b/s)",
	},
}

func init() {
	metrics.Register("proc.num")
	metrics.Register("proc.uptime")
	metrics.Register("proc.createtime")
	metrics.Register("proc.open_fd_count")
	metrics.Register("proc.mem.rss")
	metrics.Register("proc.mem.vms")
	metrics.Register("proc.mem.swap")
	metrics.Register("proc.mem.shared")
	metrics.Register("proc.mem.text")
	metrics.Register("proc.mem.lib")
	metrics.Register("proc.mem.data")
	metrics.Register("proc.mem.dirty")
	metrics.Register("proc.cpu.total")
	metrics.Register("proc.cpu.user")
	metrics.Register("proc.cpu.sys")
	metrics.Register("proc.cpu.threads")
	metrics.Register("proc.io.read_rate")
	metrics.Register("proc.io.write_rate")
	metrics.Register("proc.io.readbytes_rate")
	metrics.Register("proc.io.writebytes_rate")
	metrics.Register("proc.net.conn_rate")
	metrics.Register("proc.net.bytes_rate")
	i18n.SetLangStrings(langStrings)
}
