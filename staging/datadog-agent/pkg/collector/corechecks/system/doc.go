// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

/*
Package system provides core checks for OS-level system metrics

*/
package system

import (
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/registry/metrics"
)

var langStrings = map[string]map[string]string{
	"zh": map[string]string{
		"system.cpu.context_switches":             "cpu上下文交换次数",
		"system.cpu.guest":                        "CPU运行虚拟处理器的时间百分比。仅适用于虚拟机监控程序",
		"system.cpu.idle":                         "CPU处于空闲状态的时间百分比",
		"system.cpu.interrupt":                    "处理器在处理中断上花费的时间百分比",
		"system.cpu.iowait":                       "CPU等待IO操作完成所花费的时间百分比",
		"system.cpu.num_cores":                    "CPU核心数",
		"system.cpu.stolen":                       "虚拟CPU等待虚拟机监控程序为另一个虚拟CPU提供服务所用的时间百分比。仅适用于虚拟机",
		"system.cpu.system":                       "CPU运行内核的时间百分比",
		"system.cpu.user":                         "CPU用于运行用户空间进程的时间百分比",
		"system.cpu.used":                         "CPU使用率",
		"system.load.1":                           "1 分钟的平均系统负载(（仅限Linux）",
		"system.load.15":                          "15 分钟的平均系统负载(（仅限Linux）",
		"system.load.5":                           "5 分钟的平均系统负载(（仅限Linux）",
		"system.load.norm.1":                      "1 分钟内的平均系统负载由CPU数量标准化（仅限Linux）",
		"system.load.norm.15":                     "15 分钟内的平均系统负载由CPU数量标准化（仅限Linux）",
		"system.load.norm.5":                      "5 分钟内的平均系统负载由CPU数量标准化（仅限Linux）",
		"system.fs.file_handles.allocated":        "系统上分配的文件句柄数",
		"system.fs.file_handles.allocated_unused": "系统上未使用的已分配文件句柄数",
		"system.fs.file_handles.in_use":           "已使用的已分配文件句柄数量超过系统最大值",
		"system.fs.file_handles.max":              "系统上分配的最大文件句柄",
		"system.fs.file_handles.used":             "系统使用的已分配文件句柄数",
		"system.fs.inodes.free":                   "空闲 inode 的数量",
		"system.fs.inodes.in_use":                 "正在使用的 inode 数量占总数的百分比",
		"system.fs.inodes.total":                  "inode 总数",
		"system.fs.inodes.used":                   "正在使用的 inode 数量",
		"system.io.avg_q_sz":                      "发送到设备的请求的平均队列大小",
		"system.io.avg_rq_sz":                     "向设备发出的请求的平均大小（仅限 Linux）",
		"system.io.await":                         "向设备发出 I/O 请求的平均时间。这包括队列中的请求所花费的时间和为它们提供服务所花费的时间（仅限 Linux）",
		"system.io.r_await":                       "向设备发出读取请求的平均时间。这包括队列中请求所花费的时间和为它们提供服务所花费的时间（仅限 Linux）",
		"system.io.r_s":                           "每秒向设备发出的读取请求数",
		"system.io.rkb_s":                         "每秒从设备读取的千字节数",
		"system.io.rrqm_s":                        "每秒合并到设备队列的读取请求数（仅限 Linux）",
		"system.io.svctm":                         "向设备发出请求的平均服务时间（仅限 Linux）",
		"system.io.util":                          "向设备发出 I/O 请求的 CPU 时间百分比（仅限 Linux）",
		"system.io.w_await":                       "向设备发出写入请求的平均时间。这包括队列中请求所花费的时间和为它们提供服务所花费的时间（仅限 Linux）",
		"system.io.w_s":                           "每秒向设备发出的写请求数",
		"system.io.wkb_s":                         "每秒写入设备的千字节数",
		"system.io.wrqm_s":                        "每秒合并到设备中的写入请求数（仅限 Linux）",
		"system.mem.buffered":                     "用于文件缓冲区的物理 RAM 量",
		"system.mem.cached":                       "用作缓存内存的物理 RAM 量",
		"system.mem.commit_limit":                 "系统当前可分配的内存总量，基于过量使用率。（仅适用于 Linux）",
		"system.mem.committed":                    "已在磁盘分页文件上保留空间的物理内存量，以防必须将其写回磁盘",
		"system.mem.committed_as":                 "系统上当前分配的内存量，即使它尚未被进程'使用'。（仅适用于 Linux）",
		"system.mem.free":                         "空闲内存的数量",
		"system.mem.nonpaged":                     "操作系统用于对象的物理内存量，这些对象不能写入磁盘，但只要分配了它们就必须保留在物理内存中",
		"system.mem.page_free":                    "空闲页面文件的数量",
		"system.mem.page_tables":                  "专用于最低页表级别的内存量",
		"system.mem.page_total":                   "页面文件的总大小",
		"system.mem.page_used":                    "正在使用的页面文件的数量",
		"system.mem.paged":                        "操作系统为对象使用的物理内存量，这些对象在不使用时可以写入磁盘",
		"system.mem.pagefile.free":                "空闲页面文件的数量",
		"system.mem.pagefile.pct_free":            "免费的页面文件数量占总数的百分比",
		"system.mem.pagefile.total":               "页面文件的总大小",
		"system.mem.pagefile.used":                "正在使用的页面文件的数量",
		"system.mem.pct_usable":                   "可用物理 RAM 的数量占总量的百分比",
		"system.mem.pct_used":                     "已使用物理 RAM 的数量占总量的百分比",
		"system.mem.shared":                       "用作共享内存的物理 RAM 量",
		"system.mem.slab":                         "内核用来缓存数据结构供自己使用的内存量",
		"system.mem.total":                        "物理内存总量(mb)",
		"system.mem.usable":                       "如果存在 /proc/meminfo 中 MemAvailable 的值，但如果不存在，则回退到添加空闲 + 缓冲 + 缓存内存(mb)",
		"system.mem.used":                         "正在使用的 RAM 量(mb)",
		"system.proc.count":                       "进程数（仅限 Windows）",
		"system.proc.queue_length":                "在处理器就绪队列中观察到延迟并等待执行的线程数（仅限 Windows）",
		"system.swap.cached":                      "用作缓存的交换空间",
		"system.swap.free":                        "可用交换空间的数量",
		"system.swap.pct_free":                    "未使用的交换空间量占总数的比例(0~1)",
		"system.swap.total":                       "交换空间总量",
		"system.swap.used":                        "正在使用的交换空间量",
		"system.uptime":                           "系统运行和可用的时间",
		"system.disk.free":                        "可用的磁盘空间量",
		"system.disk.in_use":                      "正在使用的磁盘空间量占总数的百分比",
		"system.disk.read_time":                   "每台设备阅读所花费的时间，以毫秒为单位",
		"system.disk.read_time_pct":               "从磁盘读取时间所占的百分比",
		"system.disk.total":                       "磁盘空间总量",
		"system.disk.used":                        "正在使用的磁盘空间量",
		"system.disk.write_time":                  "每台设备写入所花费的时间，以毫秒为单位",
		"system.disk.write_time_pct":              "写入磁盘的时间百分比",
	},
	// https://www.kernel.org/doc/Documentation/filesystems/proc.txt
	// https://www.opsdash.com/blog/cpu-usage-linux.html
	"en": map[string]string{
		"system.cpu.context_switches":             "Count of the number of context switches",
		"system.cpu.guest":                        "The percent of time the CPU spent running the virtual processor. Only applies to hypervisors",
		"system.cpu.idle":                         "Percent of time the CPU spent in an idle state",
		"system.cpu.interrupt":                    "The percentage of time that the processor is spending on handling Interrupts",
		"system.cpu.iowait":                       "The percent of time the CPU spent waiting for IO operations to complete",
		"system.cpu.num_cores":                    "The number of CPU cores",
		"system.cpu.stolen":                       "The percent of time the virtual CPU spent waiting for the hypervisor to service another virtual CPU. Only applies to virtual machines",
		"system.cpu.system":                       "The percent of time the CPU spent running the kernel",
		"system.cpu.user":                         "The percent of time the CPU spent running user space processes",
		"system.cpu.used":                         "The percent of time the CPU used",
		"system.fs.file_handles.allocated":        "Number of allocated file handles over the system",
		"system.fs.file_handles.allocated_unused": "Number of allocated file handles unused over the system",
		"system.fs.file_handles.in_use":           "The percentage of used allocated file handles over the system max",
		"system.fs.file_handles.max":              "Maximum of allocated files handles over the system",
		"system.fs.file_handles.used":             "Number of allocated file handles used over the system",
		"system.fs.inodes.free":                   "The number of free inodes",
		"system.fs.inodes.in_use":                 "The number of inodes in use as a percent of the total",
		"system.fs.inodes.total":                  "The total number of inodes",
		"system.fs.inodes.used":                   "The number of inodes in use",
		"system.io.avg_q_sz":                      "The average queue size of requests issued to the device",
		"system.io.avg_rq_sz":                     "The average size of requests issued to the device (Linux only)",
		"system.io.await":                         "The average time for I/O requests issued to the device to be served. This includes the time spent by the requests in queue and the time spent servicing them (Linux only)",
		"system.io.r_await":                       "The average time for read requests issued to the device to be served. This includes the time spent by the requests in queue and the time spent servicing them (Linux only)",
		"system.io.r_s":                           "The number of read requests issued to the device per second",
		"system.io.rkb_s":                         "The number of kibibytes read from the device per second",
		"system.io.rrqm_s":                        "The number of read requests merged per second that were queued to the device (Linux only)",
		"system.io.svctm":                         "The average service time for requests issued to the device (Linux only)",
		"system.io.util":                          "The percent of CPU time during which I/O requests were issued to the device (Linux only)",
		"system.io.w_await":                       "The average time for write requests issued to the device to be served. This includes the time spent by the requests in queue and the time spent servicing them (Linux only)",
		"system.io.w_s":                           "The number of write requests issued to the device per second",
		"system.io.wkb_s":                         "The number of kibibytes written to the device per second",
		"system.io.wrqm_s":                        "The number of write requests merged per second that were queued to the device (Linux only)",
		"system.load.1":                           "The average system load over one minute. (Linux only)",
		"system.load.15":                          "The average system load over fifteen minutes. (Linux only)",
		"system.load.5":                           "The average system load over five minutes. (Linux only)",
		"system.load.norm.1":                      "The average system load over one minute normalized by the number of CPUs. (Linux only)",
		"system.load.norm.15":                     "The average system load over fifteen minutes normalized by the number of CPUs. (Linux only)",
		"system.load.norm.5":                      "The average system load over five minutes normalized by the number of CPUs. (Linux only)",
		"system.mem.buffered":                     "The amount of physical RAM used for file buffers",
		"system.mem.cached":                       "The amount of physical RAM used as cache memory",
		"system.mem.commit_limit":                 "The total amount of memory currently available to be allocated on the system, based on the overcommit ratio. (Linux only)",
		"system.mem.committed":                    "The amount of physical memory for which space has been reserved on the disk paging file in case it must be written back to disk",
		"system.mem.committed_as":                 "The amount of memory presently allocated on the system, even if it has not been 'used' by processes as of yet. (Linux only)",
		"system.mem.free":                         "The amount of free RAM",
		"system.mem.nonpaged":                     "The amount of physical memory used by the OS for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated",
		"system.mem.page_free":                    "The amount of the page file that's free",
		"system.mem.page_tables":                  "The amount of memory dedicated to the lowest page table level",
		"system.mem.page_total":                   "The total size of the page file",
		"system.mem.page_used":                    "The amount of the page file in use",
		"system.mem.paged":                        "The amount of physical memory used by the OS for objects that can be written to disk when they are not in use",
		"system.mem.pagefile.free":                "The amount of the page file that's free",
		"system.mem.pagefile.pct_free":            "The amount of the page file that's free as a percent of the total",
		"system.mem.pagefile.total":               "The total size of the page file",
		"system.mem.pagefile.used":                "The amount of the page file in use",
		"system.mem.pct_usable":                   "The amount of usable physical RAM as a percent of the total",
		"system.mem.pct_used":                     "The amount of used physical RAM as a percent of the total",
		"system.mem.shared":                       "The amount of physical RAM used as shared memory",
		"system.mem.slab":                         "The amount of memory used by the kernel to cache data structures for its own use",
		"system.mem.total":                        "The total amount of physical RAM",
		"system.mem.usable":                       "Value of MemAvailable from /proc/meminfo if present, but falls back to adding free + buffered + cached memory if not",
		"system.mem.used":                         "The amount of RAM in use",
		"system.proc.count":                       "The number of processes (Windows only)",
		"system.proc.queue_length":                "The number of threads that are observed as delayed in the processor ready queue and are waiting to be executed (Windows only)",
		"system.swap.cached":                      "The amount of swap used as cache memory",
		"system.swap.free":                        "The amount of free swap space",
		"system.swap.pct_free":                    "The amount of swap space not in use as a percent of the total",
		"system.swap.total":                       "The total amount of swap space",
		"system.swap.used":                        "The amount of swap space in use",
		"system.uptime":                           "The amount of time the system has been working and available",
		"system.disk.free":                        "The amount of disk space that is free",
		"system.disk.in_use":                      "The amount of disk space in use as a percent of the total",
		"system.disk.read_time":                   "The time in ms spent reading per device",
		"system.disk.read_time_pct":               "Percent of time spent reading from disk",
		"system.disk.total":                       "The total amount of disk space",
		"system.disk.used":                        "The amount of disk space in use",
		"system.disk.write_time":                  "The time in ms spent writing per device",
		"system.disk.write_time_pct":              "Percent of time spent writing to disk",
	},
}

func init() {
	// https://docs.datadoghq.com/integrations/system/
	// https://docs.datadoghq.com/integrations/disk/
	metrics.Register("system.cpu.context_switches", "count")
	metrics.Register("system.cpu.guest", "gauge", "Shown as percent")
	metrics.Register("system.cpu.idle", "gauge", "Shown as percent")
	metrics.Register("system.cpu.interrupt", "gauge", "Shown as percent")
	metrics.Register("system.cpu.iowait", "gauge", "Shown as percent")
	metrics.Register("system.cpu.num_cores", "gauge")
	metrics.Register("system.cpu.stolen", "gauge", "Shown as percent")
	metrics.Register("system.cpu.system", "gauge", "Shown as percent")
	metrics.Register("system.cpu.user", "gauge", "Shown as percent")
	metrics.Register("system.fs.file_handles.allocated", "gauge", "Shown as file")
	metrics.Register("system.fs.file_handles.allocated_unused", "gauge", "Shown as file")
	metrics.Register("system.fs.file_handles.in_use", "gauge", "Shown as percent")
	metrics.Register("system.fs.file_handles.max", "gauge", "Shown as file")
	metrics.Register("system.fs.file_handles.used", "gauge", "Shown as file")
	metrics.Register("system.fs.inodes.free", "gauge", "Shown as inode")
	metrics.Register("system.fs.inodes.in_use", "gauge", "Shown as percent")
	metrics.Register("system.fs.inodes.total", "gauge", "Shown as inode")
	metrics.Register("system.fs.inodes.used", "gauge", "Shown as inode")
	metrics.Register("system.io.avg_q_sz", "gauge", "Shown as request")
	metrics.Register("system.io.avg_rq_sz", "gauge", "Shown as sector")
	metrics.Register("system.io.await", "gauge", "Shown as millisecond")
	metrics.Register("system.io.r_await", "gauge", "Shown as millisecond")
	metrics.Register("system.io.r_s", "gauge", "Shown as request")
	metrics.Register("system.io.rkb_s", "gauge", "Shown as kibibyte")
	metrics.Register("system.io.rrqm_s", "gauge", "Shown as request")
	metrics.Register("system.io.svctm", "gauge", "Shown as millisecond")
	metrics.Register("system.io.util", "gauge", "Shown as percent")
	metrics.Register("system.io.w_await", "gauge", "Shown as millisecond")
	metrics.Register("system.io.w_s", "gauge", "Shown as request")
	metrics.Register("system.io.wkb_s", "gauge", "Shown as kibibyte")
	metrics.Register("system.io.wrqm_s", "gauge", "Shown as request")
	metrics.Register("system.load.1", "gauge")
	metrics.Register("system.load.15", "gauge")
	metrics.Register("system.load.5", "gauge")
	metrics.Register("system.load.norm.1", "gauge")
	metrics.Register("system.load.norm.15", "gauge")
	metrics.Register("system.load.norm.5", "gauge")
	metrics.Register("system.mem.buffered", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.cached", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.commit_limit", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.committed", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.committed_as", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.free", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.nonpaged", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.page_free", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.page_tables", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.page_total", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.page_used", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.paged", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.pagefile.free", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.pagefile.pct_free", "gauge", "Shown as percent")
	metrics.Register("system.mem.pagefile.total", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.pagefile.used", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.pct_usable", "gauge", "Shown as percent")
	metrics.Register("system.mem.pct_used", "gauge", "Shown as percent")
	metrics.Register("system.mem.shared", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.slab", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.total", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.usable", "gauge", "Shown as megabyte")
	metrics.Register("system.mem.used", "gauge", "Shown as megabyte")
	metrics.Register("system.proc.count", "gauge", "Shown as process")
	metrics.Register("system.proc.queue_length", "gauge", "Shown as thread")
	metrics.Register("system.swap.cached", "gauge", "Shown as megabyte")
	metrics.Register("system.swap.free", "gauge", "Shown as megabyte")
	metrics.Register("system.swap.pct_free", "gauge", "Shown as percent")
	metrics.Register("system.swap.total", "gauge", "Shown as megabyte")
	metrics.Register("system.swap.used", "gauge", "Shown as megabyte")
	metrics.Register("system.uptime", "gauge", "Shown as second")
	metrics.Register("system.disk.free", "gauge", "Shown as kibibyte")
	metrics.Register("system.disk.in_use", "gauge", "Shown as percent")
	metrics.Register("system.disk.read_time", "count", "Shown as millisecond")
	metrics.Register("system.disk.read_time_pct", "gauge", "Shown as percent")
	metrics.Register("system.disk.total", "gauge", "Shown as kibibyte")
	metrics.Register("system.disk.used", "gauge", "Shown as kibibyte")
	metrics.Register("system.disk.write_time", "count", "Shown as millisecond")
	metrics.Register("system.disk.write_time_pct", "gauge", "Shown as percent")
	metrics.Register("system.fs.inodes.free", "gauge", "Shown as inode")
	metrics.Register("system.fs.inodes.in_use", "gauge", "Shown as percent")
	metrics.Register("system.fs.inodes.total", "gauge", "Shown as inode")

	i18n.SetLangStrings(langStrings)
}
