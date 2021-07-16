package mysql

import (
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/registry/metrics"
)

var langStrings = map[string]map[string]string{
	"zh": map[string]string{
		"mysql.binlog.cache_disk_use":                     "使用临时二进制日志缓存但超过 binlog_cache_size 值并使用临时文件存储来自事务的语句的事务数",
		"mysql.binlog.cache_use":                          "使用二进制日志缓存的事务数",
		"mysql.binlog.disk_use":                           "二进制日志文件总大小",
		"mysql.galera.wsrep_cert_deps_distance":           "显示节点可能并行应用的最低和最高序列号或 seqno 值之间的平均距离",
		"mysql.galera.wsrep_cluster_size":                 "Galera 集群中的当前节点数",
		"mysql.galera.wsrep_flow_control_paused":          "显示自上次调用 FLUSH STATUS 以来节点因流量控制而暂停的时间比例",
		"mysql.galera.wsrep_flow_control_paused_ns":       "显示由于流量控制导致的暂停时间，以纳秒为单位",
		"mysql.galera.wsrep_flow_control_recv":            "显示 galera 节点收到来自其他人的暂停流控制消息的次数",
		"mysql.galera.wsrep_flow_control_sent":            "显示galera节点向其他人发送暂停流控制消息的次数",
		"mysql.galera.wsrep_local_recv_queue_avg":         "显示自上次状态查询以来本地接收队列的平均大小",
		"mysql.galera.wsrep_local_send_queue_avg":         "显示自上次 FLUSH STATUS 查询以来发送队列长度的平均值",
		"mysql.info.schema.size":                          "以 MiB 为单位的架构大小",
		"mysql.innodb.active_transactions":                "InnoDB 表上的活动事务数",
		"mysql.innodb.buffer_pool_data":                   "InnoDB 缓冲池中包含数据的总字节数",
		"mysql.innodb.buffer_pool_dirty":                  "InnoDB 缓冲池中脏页中保存的当前总字节数",
		"mysql.innodb.buffer_pool_free":                   "InnoDB 缓冲池中的空闲页面数",
		"mysql.innodb.buffer_pool_pages_data":             "InnoDB 缓冲池中包含数据的页数",
		"mysql.innodb.buffer_pool_pages_dirty":            "InnoDB 缓冲池中的当前脏页数",
		"mysql.innodb.buffer_pool_pages_flushed":          "从 InnoDB 缓冲池刷新页面的请求数",
		"mysql.innodb.buffer_pool_pages_free":             "InnoDB 缓冲池中的空闲页数",
		"mysql.innodb.buffer_pool_pages_total":            "InnoDB 缓冲池的总大小，以页为单位",
		"mysql.innodb.buffer_pool_read_ahead":             "预读后台线程读入 InnoDB 缓冲池的页数",
		"mysql.innodb.buffer_pool_read_ahead_evicted":     "由预读后台线程读入 InnoDB 缓冲池的页面数，这些页面随后在未被查询访问的情况下被驱逐",
		"mysql.innodb.buffer_pool_read_ahead_rnd":         "InnoDB 启动的随机预读次数",
		"mysql.innodb.buffer_pool_read_requests":          "逻辑读取请求的数量",
		"mysql.innodb.buffer_pool_reads":                  "InnoDB 无法从缓冲池中满足的逻辑读取数，必须直接从磁盘读取",
		"mysql.innodb.buffer_pool_total":                  "InnoDB 缓冲池中的总页数",
		"mysql.innodb.buffer_pool_used":                   "InnoDB 缓冲池中使用的页面数",
		"mysql.innodb.buffer_pool_utilization":            "InnoDB 缓冲池的利用率",
		"mysql.innodb.buffer_pool_wait_free":              "当 InnoDB 需要读取或创建页面并且没有可用的干净页面时，InnoDB 首先刷新一些脏页面并等待该操作完成",
		"mysql.innodb.buffer_pool_write_requests":         "写入 InnoDB 缓冲池的次数",
		"mysql.innodb.checkpoint_age":                     "检查点年龄如 SHOW ENGINE INNODB STATUS 输出的 LOG 部分所示",
		"mysql.innodb.current_row_locks":                  "当前行锁的数量",
		"mysql.innodb.current_transactions":               "当前 InnoDB 事务",
		"mysql.innodb.data_fsyncs":                        "每秒 fsync() 操作数",
		"mysql.innodb.data_pending_fsyncs":                "当前挂起的 fsync() 操作数",
		"mysql.innodb.data_pending_reads":                 "当前挂起的读取数",
		"mysql.innodb.data_pending_writes":                "当前挂起的写入数",
		"mysql.innodb.data_read":                          "每秒读取的数据量",
		"mysql.innodb.data_reads":                         "数据读取速率",
		"mysql.innodb.data_writes":                        "数据写入速率",
		"mysql.innodb.data_written":                       "每秒写入的数据量",
		"mysql.innodb.dblwr_pages_written":                "每秒写入双写缓冲区的页数",
		"mysql.innodb.dblwr_writes":                       "每秒执行的双写操作数",
		"mysql.innodb.hash_index_cells_total":             "自适应哈希索引的单元格总数",
		"mysql.innodb.hash_index_cells_used":              "自适应哈希索引的使用单元格数",
		"mysql.innodb.history_list_length":                "历史列表长度如 SHOW ENGINE INNODB STATUS 输出的 TRANSACTIONS 部分所示",
		"mysql.innodb.ibuf_free_list":                     "插入缓冲区空闲列表，如 SHOW ENGINE INNODB STATUS 输出的 INSERT BUFFER AND ADAPTIVE HASH INDEX 部分所示",
		"mysql.innodb.ibuf_merged":                        "插入缓冲区和自适应哈希索引合并",
		"mysql.innodb.ibuf_merged_delete_marks":           "插入缓冲区和自适应哈希索引合并删除标记",
		"mysql.innodb.ibuf_merged_deletes":                "插入缓冲区和自适应哈希索引合并删除",
		"mysql.innodb.ibuf_merged_inserts":                "插入缓冲区和自适应哈希索引合并插入",
		"mysql.innodb.ibuf_merges":                        "插入缓冲区和自适应哈希索引合并",
		"mysql.innodb.ibuf_segment_size":                  "插入缓冲区段大小，如 SHOW ENGINE INNODB STATUS 输出的 INSERT BUFFER AND ADAPTIVE HASH INDEX 部分所示",
		"mysql.innodb.ibuf_size":                          "插入缓冲区大小，如 SHOW ENGINE INNODB STATUS 输出的 INSERT BUFFER AND ADAPTIVE HASH INDEX 部分所示",
		"mysql.innodb.lock_structs":                       "锁结构",
		"mysql.innodb.log_waits":                          "日志缓冲区太小并且在继续之前需要等待刷新的次数",
		"mysql.innodb.log_write_requests":                 "InnoDB 重做日志的写入请求数",
		"mysql.innodb.log_writes":                         "物理写入 InnoDB 重做日志文件的次数",
		"mysql.innodb.lsn_current":                        "日志序列号，如 SHOW ENGINE INNODB STATUS 输出的 LOG 部分所示",
		"mysql.innodb.lsn_flushed":                        "刷新到日志序列号，如 SHOW ENGINE INNODB STATUS 输出的 LOG 部分所示",
		"mysql.innodb.lsn_last_checkpoint":                "记录序列号最后一个检查点，如 SHOW ENGINE INNODB STATUS 输出的 LOG 部分所示",
		"mysql.innodb.mem_adaptive_hash":                  "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mem_additional_pool":                "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mem_dictionary":                     "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mem_file_system":                    "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mem_lock_system":                    "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mem_page_hash":                      "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mem_recovery_system":                "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mem_total":                          "如 SHOW ENGINE INNODB STATUS 输出的 BUFFER POOL AND MEMORY 部分所示",
		"mysql.innodb.mutex_os_waits":                     "互斥锁 OS 等待的速率",
		"mysql.innodb.mutex_spin_rounds":                  "互斥自旋轮的速率",
		"mysql.innodb.mutex_spin_waits":                   "互斥量自旋等待的速率",
		"mysql.innodb.os_file_fsyncs":                     "(Delta) InnoDB 执行的 fsync() 操作总数",
		"mysql.innodb.os_file_reads":                      "(Delta) InnoDB 中读取线程执行的文件读取总数",
		"mysql.innodb.os_file_writes":                     "(Delta) InnoDB 中写入线程执行的文件写入总数",
		"mysql.innodb.os_log_fsyncs":                      "fsync 写入日志文件的速率",
		"mysql.innodb.os_log_pending_fsyncs":              "待处理的 InnoDB 日志 fsync（同步到磁盘）请求的数量",
		"mysql.innodb.os_log_pending_writes":              "挂起的 InnoDB 日志写入的数量",
		"mysql.innodb.os_log_written":                     "写入 InnoDB 日志的字节数",
		"mysql.innodb.pages_created":                      "创建的 InnoDB 页数",
		"mysql.innodb.pages_read":                         "读取的 InnoDB 页数",
		"mysql.innodb.pages_written":                      "写入的 InnoDB 页数",
		"mysql.innodb.pending_aio_log_ios":                "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_aio_sync_ios":               "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_buffer_pool_flushes":        "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_checkpoint_writes":          "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_ibuf_aio_reads":             "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_log_flushes":                "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_log_writes":                 "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_normal_aio_reads":           "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.pending_normal_aio_writes":          "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.queries_inside":                     "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.queries_queued":                     "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.read_views":                         "如 SHOW ENGINE INNODB STATUS 输出的 FILE I/O 部分所示",
		"mysql.innodb.row_lock_current_waits":             "InnoDB 表上的操作当前正在等待的行锁数",
		"mysql.innodb.row_lock_time":                      "获取行锁所花费的时间比例 (ms/s)",
		"mysql.innodb.row_lock_waits":                     "每秒必须等待行锁的次数",
		"mysql.innodb.rows_deleted":                       "从 InnoDB 表中删除的行数",
		"mysql.innodb.rows_inserted":                      "插入 InnoDB 表的行数",
		"mysql.innodb.rows_read":                          "从 InnoDB 表读取的行数",
		"mysql.innodb.rows_updated":                       "InnoDB 表中更新的行数",
		"mysql.innodb.s_lock_os_waits":                    "如 SHOW ENGINE INNODB STATUS 输出的 SEMAPHORES 部分所示",
		"mysql.innodb.s_lock_spin_rounds":                 "如 SHOW ENGINE INNODB STATUS 输出的 SEMAPHORES 部分所示",
		"mysql.innodb.s_lock_spin_waits":                  "如 SHOW ENGINE INNODB STATUS 输出的 SEMAPHORES 部分所示",
		"mysql.innodb.x_lock_os_waits":                    "如 SHOW ENGINE INNODB STATUS 输出的 SEMAPHORES 部分所示",
		"mysql.innodb.x_lock_spin_rounds":                 "如 SHOW ENGINE INNODB STATUS 输出的 SEMAPHORES 部分所示",
		"mysql.innodb.x_lock_spin_waits":                  "如 SHOW ENGINE INNODB STATUS 输出的 SEMAPHORES 部分所示",
		"mysql.myisam.key_buffer_bytes_unflushed":         "MyISAM 密钥缓冲区字节未刷新",
		"mysql.myisam.key_buffer_bytes_used":              "使用的 MyISAM 密钥缓冲区字节",
		"mysql.myisam.key_buffer_size":                    "用于索引块的缓冲区大小",
		"mysql.myisam.key_read_requests":                  "从 MyISAM 密钥缓存中读取密钥块的请求数",
		"mysql.myisam.key_reads":                          "从磁盘物理读取密钥块到 MyISAM 密钥缓存的次数",
		"mysql.myisam.key_write_requests":                 "将密钥块写入 MyISAM 密钥缓存的请求数",
		"mysql.myisam.key_writes":                         "密钥块从 MyISAM 密钥缓存物理写入磁盘的次数",
		"mysql.net.aborted_clients":                       "由于客户端在没有正确关闭连接的情况下死亡而中止的连接数",
		"mysql.net.aborted_connects":                      "尝试连接到 MySQL 服务器的失败次数",
		"mysql.net.connections":                           "与服务器的连接速率",
		"mysql.net.max_connections":                       "自服务器启动以来同时使用的最大连接数",
		"mysql.net.max_connections_available":             "允许的最大并发客户端连接数",
		"mysql.performance.bytes_received":                "从所有客户端接收的字节数",
		"mysql.performance.bytes_sent":                    "发送到所有客户端的字节数",
		"mysql.performance.com_delete":                    "删除语句的速率",
		"mysql.performance.com_delete_multi":              "delete-multi 语句的速率",
		"mysql.performance.com_insert":                    "插入语句的速率",
		"mysql.performance.com_insert_select":             "插入-选择语句的速率",
		"mysql.performance.com_load":                      "加载语句的速率",
		"mysql.performance.com_replace":                   "替换语句的速率",
		"mysql.performance.com_replace_select":            "替换-选择语句的速率",
		"mysql.performance.com_select":                    "选择语句的速率",
		"mysql.performance.com_update":                    "更新语句的速率",
		"mysql.performance.com_update_multi":              "更新倍率",
		"mysql.performance.cpu_time":                      "MySQL 花费的 CPU 时间百分比",
		"mysql.performance.created_tmp_disk_tables":       "服务器在执行语句时每秒创建的内部磁盘临时表的速率",
		"mysql.performance.created_tmp_files":             "每秒创建临时文件的速率",
		"mysql.performance.created_tmp_tables":            "服务器在执行语句时每秒创建的内部临时表的速率",
		"mysql.performance.digest_95th_percentile.avg_us": "每个模式的查询响应时间第 95 个百分位",
		"mysql.performance.handler_commit":                "内部 COMMIT 语句的数量",
		"mysql.performance.handler_delete":                "内部 DELETE 语句的数量",
		"mysql.performance.handler_prepare":               "内部 PREPARE 语句的数量",
		"mysql.performance.handler_read_first":            "内部 READ_FIRST 语句的数量",
		"mysql.performance.handler_read_key":              "内部 READ_KEY 语句的数量",
		"mysql.performance.handler_read_next":             "内部 READ_NEXT 语句的数量",
		"mysql.performance.handler_read_prev":             "内部 READ_PREV 语句的数量",
		"mysql.performance.handler_read_rnd":              "内部 READ_RND 语句的数量",
		"mysql.performance.handler_read_rnd_next":         "内部 READ_RND_NEXT 语句的数量",
		"mysql.performance.handler_rollback":              "内部 ROLLBACK 语句的数量",
		"mysql.performance.handler_update":                "内部 UPDATE 语句的数量",
		"mysql.performance.handler_write":                 "内部 WRITE 语句的数量",
		"mysql.performance.kernel_time":                   "MySQL 在内核空间中花费的 CPU 时间百分比",
		"mysql.performance.key_cache_utilization":         "键缓存利用率",
		"mysql.performance.open_files":                    "打开的文件数",
		"mysql.performance.open_tables":                   "打开的表的数量",
		"mysql.performance.opened_tables":                 "已打开的表数",
		"mysql.performance.qcache.utilization":            "当前正在使用的查询缓存内存的一部分",
		"mysql.performance.qcache_free_blocks":            "查询缓存中的空闲内存块数",
		"mysql.performance.qcache_free_memory":            "查询缓存的可用内存量",
		"mysql.performance.qcache_hits":                   "查询缓存命中率",
		"mysql.performance.qcache_inserts":                "添加到查询缓存的查询数",
		"mysql.performance.qcache_lowmem_prunes":          "由于内存不足而从查询缓存中删除的查询数",
		"mysql.performance.qcache_not_cached":             "非缓存查询的数量（不可缓存，或由于 query_cache_type 设置而未缓存）",
		"mysql.performance.qcache_queries_in_cache":       "在查询缓存中注册的查询数",
		"mysql.performance.qcache_size":                   "为缓存查询结果分配的内存量",
		"mysql.performance.qcache_total_blocks":           "查询缓存中的块总数",
		"mysql.performance.queries":                       "查询率",
		"mysql.performance.query_run_time.avg":            "每个架构的平均查询响应时间",
		"mysql.performance.questions":                     "服务器执行语句的速率",
		"mysql.performance.select_full_join":              "由于不使用索引而执行表扫描的连接数",
		"mysql.performance.select_full_range_join":        "对引用表使用范围搜索的连接数",
		"mysql.performance.select_range":                  "在第一个表上使用范围的连接数",
		"mysql.performance.select_range_check":            "在每行之后检查键使用情况的没有键的连接数",
		"mysql.performance.select_scan":                   "对第一个表进行完整扫描的连接数",
		"mysql.performance.slow_queries":                  "慢查询的速度",
		"mysql.performance.sort_merge_passes":             "排序算法必须执行的合并传递次数",
		"mysql.performance.sort_range":                    "使用范围完成的排序数",
		"mysql.performance.sort_rows":                     "已排序的行数",
		"mysql.performance.sort_scan":                     "通过扫描表完成的排序数",
		"mysql.performance.table_cache_hits":              "打开表缓存查找的命中数",
		"mysql.performance.table_cache_misses":            "打开表缓存查找的未命中数",
		"mysql.performance.table_locks_immediate":         "可以立即授予表锁请求的次数",
		"mysql.performance.table_locks_immediate.rate":    "可以立即授予表锁请求的次数",
		"mysql.performance.table_locks_waited":            "无法立即授予表锁请求并需要等待的总次数",
		"mysql.performance.table_locks_waited.rate":       "无法立即授予表锁请求并需要等待的次数",
		"mysql.performance.table_open_cache":              "所有线程的打开表数",
		"mysql.performance.thread_cache_size":             "服务器应该缓存多少线程以供重用",
		"mysql.performance.threads_cached":                "线程缓存中的线程数",
		"mysql.performance.threads_connected":             "当前打开的连接数",
		"mysql.performance.threads_created":               "为处理连接而创建的线程数",
		"mysql.performance.threads_running":               "未休眠的线程数",
		"mysql.performance.user_time":                     "MySQL 在用户空间中花费的 CPU 时间百分比",
		"mysql.queries.count":                             "每个模式每个查询执行的查询总数",
		"mysql.queries.errors":                            "每个模式每个查询运行时出现错误的查询总数",
		"mysql.queries.lock_time":                         "每个模式每个查询等待锁定所花费的总时间",
		"mysql.queries.no_good_index_used":                "每个模式每个查询使用次优索引的查询总数",
		"mysql.queries.no_index_used":                     "每个模式的每个查询不使用索引的查询总数",
		"mysql.queries.rows_affected":                     "每个模式每个查询发生变异的行数",
		"mysql.queries.rows_sent":                         "每个模式每个查询发送的行数",
		"mysql.queries.select_full_join":                  "每个模式每个查询对连接表进行全表扫描的总数",
		"mysql.queries.select_scan":                       "每个模式每个查询的第一个表上的全表扫描总数",
		"mysql.queries.time":                              "每个模式每个查询的总查询执行时间",
		"mysql.replication.replicas_connected":            "连接到复制源的副本数",
		"mysql.replication.seconds_behind_master":         "master 和 slave 之间的延迟（以秒为单位）",
		"mysql.replication.seconds_behind_source":         "源和副本之间的延迟（以秒为单位）",
		"mysql.replication.slave_running":                 "显示此服务器是否是正在运行的复制从属/主服务器",
		"mysql.replication.slaves_connected":              "连接到复制主机的从属设备数",
	},
	"en": map[string]string{
		"mysql.binlog.cache_disk_use":                     "The number of transactions that used the temporary binary log cache but that exceeded the value of binlog_cache_size and used a temporary file to store statements from the transaction",
		"mysql.binlog.cache_use":                          "The number of transactions that used the binary log cache",
		"mysql.binlog.disk_use":                           "Total binary log file size",
		"mysql.galera.wsrep_cert_deps_distance":           "Shows the average distance between the lowest and highest sequence number, or seqno, values that the node can possibly apply in parallel",
		"mysql.galera.wsrep_cluster_size":                 "The current number of nodes in the Galera cluster",
		"mysql.galera.wsrep_flow_control_paused":          "Shows the fraction of the time, since FLUSH STATUS was last called, that the node paused due to Flow Control",
		"mysql.galera.wsrep_flow_control_paused_ns":       "Shows the pause time due to Flow Control, in nanoseconds",
		"mysql.galera.wsrep_flow_control_recv":            "Shows the number of times the galera node has received a pausing Flow Control message from others",
		"mysql.galera.wsrep_flow_control_sent":            "Shows the number of times the galera node has sent a pausing Flow Control message to others",
		"mysql.galera.wsrep_local_recv_queue_avg":         "Shows the average size of the local received queue since the last status query",
		"mysql.galera.wsrep_local_send_queue_avg":         "Show an average for the send queue length since the last FLUSH STATUS query",
		"mysql.info.schema.size":                          "Size of schemas in MiB",
		"mysql.innodb.active_transactions":                "The number of active transactions on InnoDB tables",
		"mysql.innodb.buffer_pool_data":                   "The total number of bytes in the InnoDB buffer pool containing data. The number includes both dirty and clean pages",
		"mysql.innodb.buffer_pool_dirty":                  "The total current number of bytes held in dirty pages in the InnoDB buffer pool",
		"mysql.innodb.buffer_pool_free":                   "The number of free pages in the InnoDB Buffer Pool",
		"mysql.innodb.buffer_pool_pages_data":             "The number of pages in the InnoDB buffer pool containing data. The number includes both dirty and clean pages",
		"mysql.innodb.buffer_pool_pages_dirty":            "The current number of dirty pages in the InnoDB buffer pool",
		"mysql.innodb.buffer_pool_pages_flushed":          "The number of requests to flush pages from the InnoDB buffer pool",
		"mysql.innodb.buffer_pool_pages_free":             "The number of free pages in the InnoDB buffer pool",
		"mysql.innodb.buffer_pool_pages_total":            "The total size of the InnoDB buffer pool, in pages",
		"mysql.innodb.buffer_pool_read_ahead":             "The number of pages read into the InnoDB buffer pool by the read-ahead background thread",
		"mysql.innodb.buffer_pool_read_ahead_evicted":     "The number of pages read into the InnoDB buffer pool by the read-ahead background thread that were subsequently evicted without having been accessed by queries",
		"mysql.innodb.buffer_pool_read_ahead_rnd":         "The number of random read-aheads initiated by InnoDB. This happens when a query scans a large portion of a table but in random order",
		"mysql.innodb.buffer_pool_read_requests":          "The number of logical read requests",
		"mysql.innodb.buffer_pool_reads":                  "The number of logical reads that InnoDB could not satisfy from the buffer pool, and had to read directly from disk",
		"mysql.innodb.buffer_pool_total":                  "The total number of pages in the InnoDB Buffer Pool",
		"mysql.innodb.buffer_pool_used":                   "The number of used pages in the InnoDB Buffer Pool",
		"mysql.innodb.buffer_pool_utilization":            "The utilization of the InnoDB Buffer Pool",
		"mysql.innodb.buffer_pool_wait_free":              "When InnoDB needs to read or create a page and no clean pages are available, InnoDB flushes some dirty pages first and waits for that operation to finish. This counter counts instances of these waits",
		"mysql.innodb.buffer_pool_write_requests":         "The number of writes done to the InnoDB buffer pool",
		"mysql.innodb.checkpoint_age":                     "Checkpoint age as shown in the LOG section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.current_row_locks":                  "The number of current row locks",
		"mysql.innodb.current_transactions":               "Current InnoDB transactions",
		"mysql.innodb.data_fsyncs":                        "The number of fsync() operations per second",
		"mysql.innodb.data_pending_fsyncs":                "The current number of pending fsync() operations",
		"mysql.innodb.data_pending_reads":                 "The current number of pending reads",
		"mysql.innodb.data_pending_writes":                "The current number of pending writes",
		"mysql.innodb.data_read":                          "The amount of data read per second",
		"mysql.innodb.data_reads":                         "The rate of data reads",
		"mysql.innodb.data_writes":                        "The rate of data writes",
		"mysql.innodb.data_written":                       "The amount of data written per second",
		"mysql.innodb.dblwr_pages_written":                "The number of pages written per second to the doublewrite buffer",
		"mysql.innodb.dblwr_writes":                       "The number of doublewrite operations performed per second",
		"mysql.innodb.hash_index_cells_total":             "Total number of cells of the adaptive hash index",
		"mysql.innodb.hash_index_cells_used":              "Number of used cells of the adaptive hash index",
		"mysql.innodb.history_list_length":                "History list length as shown in the TRANSACTIONS section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.ibuf_free_list":                     "Insert buffer free list, as shown in the INSERT BUFFER AND ADAPTIVE HASH INDEX section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.ibuf_merged":                        "Insert buffer and adaptative hash index merged",
		"mysql.innodb.ibuf_merged_delete_marks":           "Insert buffer and adaptative hash index merged delete marks",
		"mysql.innodb.ibuf_merged_deletes":                "Insert buffer and adaptative hash index merged delete",
		"mysql.innodb.ibuf_merged_inserts":                "Insert buffer and adaptative hash index merged inserts",
		"mysql.innodb.ibuf_merges":                        "Insert buffer and adaptative hash index merges",
		"mysql.innodb.ibuf_segment_size":                  "Insert buffer segment size, as shown in the INSERT BUFFER AND ADAPTIVE HASH INDEX section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.ibuf_size":                          "Insert buffer size, as shown in the INSERT BUFFER AND ADAPTIVE HASH INDEX section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.lock_structs":                       "Lock structs",
		"mysql.innodb.log_waits":                          "The number of times that the log buffer was too small and a wait was required for it to be flushed before continuing",
		"mysql.innodb.log_write_requests":                 "The number of write requests for the InnoDB redo log",
		"mysql.innodb.log_writes":                         "The number of physical writes to the InnoDB redo log file",
		"mysql.innodb.lsn_current":                        "Log sequence number as shown in the LOG section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.lsn_flushed":                        "Flushed up to log sequence number as shown in the LOG section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.lsn_last_checkpoint":                "Log sequence number last checkpoint as shown in the LOG section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_adaptive_hash":                  "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_additional_pool":                "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_dictionary":                     "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_file_system":                    "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_lock_system":                    "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_page_hash":                      "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_recovery_system":                "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mem_total":                          "As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.mutex_os_waits":                     "The rate of mutex OS waits",
		"mysql.innodb.mutex_spin_rounds":                  "The rate of mutex spin rounds",
		"mysql.innodb.mutex_spin_waits":                   "The rate of mutex spin waits",
		"mysql.innodb.os_file_fsyncs":                     "(Delta) The total number of fsync() operations performed by InnoDB",
		"mysql.innodb.os_file_reads":                      "(Delta) The total number of files reads performed by read threads within InnoDB",
		"mysql.innodb.os_file_writes":                     "(Delta) The total number of file writes performed by write threads within InnoDB",
		"mysql.innodb.os_log_fsyncs":                      "The rate of fsync writes to the log file",
		"mysql.innodb.os_log_pending_fsyncs":              "Number of pending InnoDB log fsync (sync-to-disk) requests",
		"mysql.innodb.os_log_pending_writes":              "Number of pending InnoDB log writes",
		"mysql.innodb.os_log_written":                     "Number of bytes written to the InnoDB log",
		"mysql.innodb.pages_created":                      "Number of InnoDB pages created",
		"mysql.innodb.pages_read":                         "Number of InnoDB pages read",
		"mysql.innodb.pages_written":                      "Number of InnoDB pages written",
		"mysql.innodb.pending_aio_log_ios":                "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_aio_sync_ios":               "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_buffer_pool_flushes":        "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_checkpoint_writes":          "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_ibuf_aio_reads":             "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_log_flushes":                "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_log_writes":                 "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_normal_aio_reads":           "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.pending_normal_aio_writes":          "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.queries_inside":                     "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.queries_queued":                     "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.read_views":                         "As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.row_lock_current_waits":             "The number of row locks currently being waited for by operations on InnoDB tables",
		"mysql.innodb.row_lock_time":                      "Fraction of time spent (ms/s) acquiring row locks",
		"mysql.innodb.row_lock_waits":                     "The number of times per second a row lock had to be waited for",
		"mysql.innodb.rows_deleted":                       "Number of rows deleted from InnoDB tables",
		"mysql.innodb.rows_inserted":                      "Number of rows inserted into InnoDB tables",
		"mysql.innodb.rows_read":                          "Number of rows read from InnoDB tables",
		"mysql.innodb.rows_updated":                       "Number of rows updated in InnoDB tables",
		"mysql.innodb.s_lock_os_waits":                    "As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.s_lock_spin_rounds":                 "As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.s_lock_spin_waits":                  "As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.x_lock_os_waits":                    "As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.x_lock_spin_rounds":                 "As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output",
		"mysql.innodb.x_lock_spin_waits":                  "As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output",
		"mysql.myisam.key_buffer_bytes_unflushed":         "MyISAM key buffer bytes unflushed",
		"mysql.myisam.key_buffer_bytes_used":              "MyISAM key buffer bytes used",
		"mysql.myisam.key_buffer_size":                    "Size of the buffer used for index blocks",
		"mysql.myisam.key_read_requests":                  "The number of requests to read a key block from the MyISAM key cache",
		"mysql.myisam.key_reads":                          "The number of physical reads of a key block from disk into the MyISAM key cache. If Key_reads is large, then your key_buffer_size value is probably too small. The cache miss rate can be calculated as Key_reads/Key_read_requests",
		"mysql.myisam.key_write_requests":                 "The number of requests to write a key block to the MyISAM key cache",
		"mysql.myisam.key_writes":                         "The number of physical writes of a key block from the MyISAM key cache to disk",
		"mysql.net.aborted_clients":                       "The number of connections that were aborted because the client died without closing the connection properly",
		"mysql.net.aborted_connects":                      "The number of failed attempts to connect to the MySQL server",
		"mysql.net.connections":                           "The rate of connections to the server",
		"mysql.net.max_connections":                       "The maximum number of connections that have been in use simultaneously since the server started",
		"mysql.net.max_connections_available":             "The maximum permitted number of simultaneous client connections",
		"mysql.performance.bytes_received":                "The number of bytes received from all clients",
		"mysql.performance.bytes_sent":                    "The number of bytes sent to all clients",
		"mysql.performance.com_delete":                    "The rate of delete statements",
		"mysql.performance.com_delete_multi":              "The rate of delete-multi statements",
		"mysql.performance.com_insert":                    "The rate of insert statements",
		"mysql.performance.com_insert_select":             "The rate of insert-select statements",
		"mysql.performance.com_load":                      "The rate of load statements",
		"mysql.performance.com_replace":                   "The rate of replace statements",
		"mysql.performance.com_replace_select":            "The rate of replace-select statements",
		"mysql.performance.com_select":                    "The rate of select statements",
		"mysql.performance.com_update":                    "The rate of update statements",
		"mysql.performance.com_update_multi":              "The rate of update-multi",
		"mysql.performance.cpu_time":                      "Percentage of CPU time spent by MySQL",
		"mysql.performance.created_tmp_disk_tables":       "The rate of internal on-disk temporary tables created by second by the server while executing statements",
		"mysql.performance.created_tmp_files":             "The rate of temporary files created by second",
		"mysql.performance.created_tmp_tables":            "The rate of internal temporary tables created by second by the server while executing statements",
		"mysql.performance.digest_95th_percentile.avg_us": "Query response time 95th percentile per schema",
		"mysql.performance.handler_commit":                "The number of internal COMMIT statements",
		"mysql.performance.handler_delete":                "The number of internal DELETE statements",
		"mysql.performance.handler_prepare":               "The number of internal PREPARE statements",
		"mysql.performance.handler_read_first":            "The number of internal READ_FIRST statements",
		"mysql.performance.handler_read_key":              "The number of internal READ_KEY statements",
		"mysql.performance.handler_read_next":             "The number of internal READ_NEXT statements",
		"mysql.performance.handler_read_prev":             "The number of internal READ_PREV statements",
		"mysql.performance.handler_read_rnd":              "The number of internal READ_RND statements",
		"mysql.performance.handler_read_rnd_next":         "The number of internal READ_RND_NEXT statements",
		"mysql.performance.handler_rollback":              "The number of internal ROLLBACK statements",
		"mysql.performance.handler_update":                "The number of internal UPDATE statements",
		"mysql.performance.handler_write":                 "The number of internal WRITE statements",
		"mysql.performance.kernel_time":                   "Percentage of CPU time spent in kernel space by MySQL",
		"mysql.performance.key_cache_utilization":         "The key cache utilization ratio",
		"mysql.performance.open_files":                    "The number of open files",
		"mysql.performance.open_tables":                   "The number of of tables that are open",
		"mysql.performance.opened_tables":                 "The number of tables that have been opened. If Opened_tables is big, your table_open_cache value is probably too small",
		"mysql.performance.qcache.utilization":            "Fraction of the query cache memory currently being used",
		"mysql.performance.qcache_free_blocks":            "The number of free memory blocks in the query cache",
		"mysql.performance.qcache_free_memory":            "The amount of free memory for the query cache",
		"mysql.performance.qcache_hits":                   "The rate of query cache hits",
		"mysql.performance.qcache_inserts":                "The number of queries added to the query cache",
		"mysql.performance.qcache_lowmem_prunes":          "The number of queries that were deleted from the query cache because of low memory",
		"mysql.performance.qcache_not_cached":             "The number of noncached queries (not cacheable, or not cached due to the query_cache_type setting)",
		"mysql.performance.qcache_queries_in_cache":       "The number of queries registered in the query cache",
		"mysql.performance.qcache_size":                   "The amount of memory allocated for caching query results",
		"mysql.performance.qcache_total_blocks":           "The total number of blocks in the query cache",
		"mysql.performance.queries":                       "The rate of queries",
		"mysql.performance.query_run_time.avg":            "Avg query response time per schema",
		"mysql.performance.questions":                     "The rate of statements executed by the server",
		"mysql.performance.select_full_join":              "The number of joins that perform table scans because they do not use indexes. If this value is not 0, you should carefully check the indexes of your tables",
		"mysql.performance.select_full_range_join":        "The number of joins that used a range search on a reference table",
		"mysql.performance.select_range":                  "The number of joins that used ranges on the first table. This is normally not a critical issue even if the value is quite large",
		"mysql.performance.select_range_check":            "The number of joins without keys that check for key usage after each row. If this is not 0, you should carefully check the indexes of your tables",
		"mysql.performance.select_scan":                   "The number of joins that did a full scan of the first table",
		"mysql.performance.slow_queries":                  "The rate of slow queries",
		"mysql.performance.sort_merge_passes":             "The number of merge passes that the sort algorithm has had to do. If this value is large, you should consider increasing the value of the sort_buffer_size system variable",
		"mysql.performance.sort_range":                    "The number of sorts that were done using ranges",
		"mysql.performance.sort_rows":                     "The number of sorted rows",
		"mysql.performance.sort_scan":                     "The number of sorts that were done by scanning the table",
		"mysql.performance.table_cache_hits":              "The number of hits for open tables cache lookups",
		"mysql.performance.table_cache_misses":            "The number of misses for open tables cache lookups",
		"mysql.performance.table_locks_immediate":         "The number of times that a request for a table lock could be granted immediately",
		"mysql.performance.table_locks_immediate.rate":    "The rate of times that a request for a table lock could be granted immediately",
		"mysql.performance.table_locks_waited":            "The total number of times that a request for a table lock could not be granted immediately and a wait was needed",
		"mysql.performance.table_locks_waited.rate":       "The rate of times that a request for a table lock could not be granted immediately and a wait was needed",
		"mysql.performance.table_open_cache":              "The number of open tables for all threads. Increasing this value increases the number of file descriptors that mysqld requires",
		"mysql.performance.thread_cache_size":             "How many threads the server should cache for reuse. When a client disconnects, the client's threads are put in the cache if there are fewer than thread_cache_size threads there",
		"mysql.performance.threads_cached":                "The number of threads in the thread cache",
		"mysql.performance.threads_connected":             "The number of currently open connections",
		"mysql.performance.threads_created":               "The number of threads created to handle connections. If Threads_created is big, you may want to increase the thread_cache_size value",
		"mysql.performance.threads_running":               "The number of threads that are not sleeping",
		"mysql.performance.user_time":                     "Percentage of CPU time spent in user space by MySQL",
		"mysql.queries.count":                             "The total count of executed queries per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.errors":                            "The total count of queries run with an error per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.lock_time":                         "The total time spent waiting on locks per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.no_good_index_used":                "The total count of queries which used a sub-optimal index per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.no_index_used":                     "The total count of queries which do not use an index per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.rows_affected":                     "The number of rows mutated per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.rows_sent":                         "The number of rows sent per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.select_full_join":                  "The total count of full table scans on a joined table per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.select_scan":                       "The total count of full table scans on the first table per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.queries.time":                              "The total query execution time per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA",
		"mysql.replication.replicas_connected":            "Number of replicas connected to a replication source",
		"mysql.replication.seconds_behind_master":         "The lag in seconds between the master and the slave",
		"mysql.replication.seconds_behind_source":         "The lag in seconds between the source and the replica",
		"mysql.replication.slave_running":                 "A boolean showing if this server is a replication slave / master that is running",
		"mysql.replication.slaves_connected":              "Number of slaves connected to a replication master",
	},
}

func registerMetric() {
	m := metrics.GetMetricGroup("process")

	m.Register("mysql.binlog.cache_disk_use", "gauge", "Shown as transaction")
	m.Register("mysql.binlog.cache_use", "gauge", "Shown as transaction")
	m.Register("mysql.binlog.disk_use", "gauge", "Shown as byte")
	m.Register("mysql.galera.wsrep_cert_deps_distance", "gauge")
	m.Register("mysql.galera.wsrep_cluster_size", "gauge", "Shown as node")
	m.Register("mysql.galera.wsrep_flow_control_paused", "gauge", "Shown as fraction")
	m.Register("mysql.galera.wsrep_flow_control_paused_ns", "count", "Shown as nanosecond")
	m.Register("mysql.galera.wsrep_flow_control_recv", "count")
	m.Register("mysql.galera.wsrep_flow_control_sent", "count")
	m.Register("mysql.galera.wsrep_local_recv_queue_avg", "gauge")
	m.Register("mysql.galera.wsrep_local_send_queue_avg", "gauge")
	m.Register("mysql.info.schema.size", "gauge", "Shown as mebibyte")
	m.Register("mysql.innodb.active_transactions", "gauge", "Shown as operation")
	m.Register("mysql.innodb.buffer_pool_data", "gauge", "Shown as byte")
	m.Register("mysql.innodb.buffer_pool_dirty", "gauge", "Shown as byte")
	m.Register("mysql.innodb.buffer_pool_free", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_pages_data", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_pages_dirty", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_pages_flushed", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_pages_free", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_pages_total", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_read_ahead", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_read_ahead_evicted", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_read_ahead_rnd", "gauge", "Shown as operation")
	m.Register("mysql.innodb.buffer_pool_read_requests", "gauge", "Shown as read")
	m.Register("mysql.innodb.buffer_pool_reads", "gauge", "Shown as read")
	m.Register("mysql.innodb.buffer_pool_total", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_used", "gauge", "Shown as page")
	m.Register("mysql.innodb.buffer_pool_utilization", "gauge", "Shown as fraction")
	m.Register("mysql.innodb.buffer_pool_wait_free", "count", "Shown as wait")
	m.Register("mysql.innodb.buffer_pool_write_requests", "gauge", "Shown as write")
	m.Register("mysql.innodb.checkpoint_age", "gauge")
	m.Register("mysql.innodb.current_row_locks", "gauge", "Shown as lock")
	m.Register("mysql.innodb.current_transactions", "gauge", "Shown as transaction")
	m.Register("mysql.innodb.data_fsyncs", "gauge", "Shown as operation")
	m.Register("mysql.innodb.data_pending_fsyncs", "gauge", "Shown as operation")
	m.Register("mysql.innodb.data_pending_reads", "gauge", "Shown as read")
	m.Register("mysql.innodb.data_pending_writes", "gauge", "Shown as write")
	m.Register("mysql.innodb.data_read", "gauge", "Shown as byte")
	m.Register("mysql.innodb.data_reads", "gauge", "Shown as read")
	m.Register("mysql.innodb.data_writes", "gauge", "Shown as write")
	m.Register("mysql.innodb.data_written", "gauge", "Shown as byte")
	m.Register("mysql.innodb.dblwr_pages_written", "gauge", "Shown as page")
	m.Register("mysql.innodb.dblwr_writes", "gauge", "Shown as byte")
	m.Register("mysql.innodb.hash_index_cells_total", "gauge")
	m.Register("mysql.innodb.hash_index_cells_used", "gauge")
	m.Register("mysql.innodb.history_list_length", "gauge")
	m.Register("mysql.innodb.ibuf_free_list", "gauge")
	m.Register("mysql.innodb.ibuf_merged", "gauge", "Shown as operation")
	m.Register("mysql.innodb.ibuf_merged_delete_marks", "gauge", "Shown as operation")
	m.Register("mysql.innodb.ibuf_merged_deletes", "gauge", "Shown as operation")
	m.Register("mysql.innodb.ibuf_merged_inserts", "gauge", "Shown as operation")
	m.Register("mysql.innodb.ibuf_merges", "gauge", "Shown as operation")
	m.Register("mysql.innodb.ibuf_segment_size", "gauge")
	m.Register("mysql.innodb.ibuf_size", "gauge")
	m.Register("mysql.innodb.lock_structs", "gauge", "Shown as operation")
	m.Register("mysql.innodb.log_waits", "gauge", "Shown as wait")
	m.Register("mysql.innodb.log_write_requests", "gauge", "Shown as write")
	m.Register("mysql.innodb.log_writes", "gauge", "Shown as write")
	m.Register("mysql.innodb.lsn_current", "gauge")
	m.Register("mysql.innodb.lsn_flushed", "gauge")
	m.Register("mysql.innodb.lsn_last_checkpoint", "gauge")
	m.Register("mysql.innodb.mem_adaptive_hash", "gauge", "Shown as byte")
	m.Register("mysql.innodb.mem_additional_pool", "gauge", "Shown as byte")
	m.Register("mysql.innodb.mem_dictionary", "gauge", "Shown as byte")
	m.Register("mysql.innodb.mem_file_system", "gauge")
	m.Register("mysql.innodb.mem_lock_system", "gauge")
	m.Register("mysql.innodb.mem_page_hash", "gauge")
	m.Register("mysql.innodb.mem_recovery_system", "gauge")
	m.Register("mysql.innodb.mem_total", "gauge", "Shown as byte")
	m.Register("mysql.innodb.mutex_os_waits", "gauge", "Shown as event")
	m.Register("mysql.innodb.mutex_spin_rounds", "gauge", "Shown as event")
	m.Register("mysql.innodb.mutex_spin_waits", "gauge", "Shown as event")
	m.Register("mysql.innodb.os_file_fsyncs", "gauge", "Shown as operation")
	m.Register("mysql.innodb.os_file_reads", "gauge", "Shown as operation")
	m.Register("mysql.innodb.os_file_writes", "gauge", "Shown as operation")
	m.Register("mysql.innodb.os_log_fsyncs", "gauge", "Shown as write")
	m.Register("mysql.innodb.os_log_pending_fsyncs", "gauge", "Shown as operation")
	m.Register("mysql.innodb.os_log_pending_writes", "gauge", "Shown as write")
	m.Register("mysql.innodb.os_log_written", "gauge", "Shown as byte")
	m.Register("mysql.innodb.pages_created", "gauge", "Shown as page")
	m.Register("mysql.innodb.pages_read", "gauge", "Shown as page")
	m.Register("mysql.innodb.pages_written", "gauge", "Shown as page")
	m.Register("mysql.innodb.pending_aio_log_ios", "gauge")
	m.Register("mysql.innodb.pending_aio_sync_ios", "gauge")
	m.Register("mysql.innodb.pending_buffer_pool_flushes", "gauge", "Shown as flush")
	m.Register("mysql.innodb.pending_checkpoint_writes", "gauge")
	m.Register("mysql.innodb.pending_ibuf_aio_reads", "gauge")
	m.Register("mysql.innodb.pending_log_flushes", "gauge", "Shown as flush")
	m.Register("mysql.innodb.pending_log_writes", "gauge", "Shown as write")
	m.Register("mysql.innodb.pending_normal_aio_reads", "gauge", "Shown as read")
	m.Register("mysql.innodb.pending_normal_aio_writes", "gauge", "Shown as write")
	m.Register("mysql.innodb.queries_inside", "gauge", "Shown as query")
	m.Register("mysql.innodb.queries_queued", "gauge", "Shown as query")
	m.Register("mysql.innodb.read_views", "gauge")
	m.Register("mysql.innodb.row_lock_current_waits", "gauge")
	m.Register("mysql.innodb.row_lock_time", "gauge", "Shown as fraction")
	m.Register("mysql.innodb.row_lock_waits", "gauge", "Shown as event")
	m.Register("mysql.innodb.rows_deleted", "gauge", "Shown as row")
	m.Register("mysql.innodb.rows_inserted", "gauge", "Shown as row")
	m.Register("mysql.innodb.rows_read", "gauge", "Shown as row")
	m.Register("mysql.innodb.rows_updated", "gauge", "Shown as row")
	m.Register("mysql.innodb.s_lock_os_waits", "gauge")
	m.Register("mysql.innodb.s_lock_spin_rounds", "gauge")
	m.Register("mysql.innodb.s_lock_spin_waits", "gauge", "Shown as wait")
	m.Register("mysql.innodb.x_lock_os_waits", "gauge", "Shown as wait")
	m.Register("mysql.innodb.x_lock_spin_rounds", "gauge")
	m.Register("mysql.innodb.x_lock_spin_waits", "gauge", "Shown as wait")
	m.Register("mysql.myisam.key_buffer_bytes_unflushed", "gauge", "Shown as byte")
	m.Register("mysql.myisam.key_buffer_bytes_used", "gauge", "Shown as byte")
	m.Register("mysql.myisam.key_buffer_size", "gauge", "Shown as byte")
	m.Register("mysql.myisam.key_read_requests", "gauge", "Shown as read")
	m.Register("mysql.myisam.key_reads", "gauge", "Shown as read")
	m.Register("mysql.myisam.key_write_requests", "gauge", "Shown as write")
	m.Register("mysql.myisam.key_writes", "gauge", "Shown as write")
	m.Register("mysql.net.aborted_clients", "gauge", "Shown as connection")
	m.Register("mysql.net.aborted_connects", "gauge", "Shown as connection")
	m.Register("mysql.net.connections", "gauge", "Shown as connection")
	m.Register("mysql.net.max_connections", "gauge", "Shown as connection")
	m.Register("mysql.net.max_connections_available", "gauge", "Shown as connection")
	m.Register("mysql.performance.bytes_received", "gauge", "Shown as byte")
	m.Register("mysql.performance.bytes_sent", "gauge", "Shown as byte")
	m.Register("mysql.performance.com_delete", "gauge", "Shown as query")
	m.Register("mysql.performance.com_delete_multi", "gauge", "Shown as query")
	m.Register("mysql.performance.com_insert", "gauge", "Shown as query")
	m.Register("mysql.performance.com_insert_select", "gauge", "Shown as query")
	m.Register("mysql.performance.com_load", "gauge", "Shown as query")
	m.Register("mysql.performance.com_replace", "gauge", "Shown as query")
	m.Register("mysql.performance.com_replace_select", "gauge", "Shown as query")
	m.Register("mysql.performance.com_select", "gauge", "Shown as query")
	m.Register("mysql.performance.com_update", "gauge", "Shown as query")
	m.Register("mysql.performance.com_update_multi", "gauge", "Shown as query")
	m.Register("mysql.performance.cpu_time", "gauge", "Shown as percent")
	m.Register("mysql.performance.created_tmp_disk_tables", "gauge", "Shown as table")
	m.Register("mysql.performance.created_tmp_files", "gauge", "Shown as file")
	m.Register("mysql.performance.created_tmp_tables", "gauge", "Shown as table")
	m.Register("mysql.performance.digest_95th_percentile.avg_us", "gauge", "Shown as microsecond")
	m.Register("mysql.performance.handler_commit", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_delete", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_prepare", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_read_first", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_read_key", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_read_next", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_read_prev", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_read_rnd", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_read_rnd_next", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_rollback", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_update", "gauge", "Shown as operation")
	m.Register("mysql.performance.handler_write", "gauge", "Shown as operation")
	m.Register("mysql.performance.kernel_time", "gauge", "Shown as percent")
	m.Register("mysql.performance.key_cache_utilization", "gauge", "Shown as fraction")
	m.Register("mysql.performance.open_files", "gauge", "Shown as file")
	m.Register("mysql.performance.open_tables", "gauge", "Shown as table")
	m.Register("mysql.performance.opened_tables", "gauge", "Shown as table")
	m.Register("mysql.performance.qcache.utilization", "gauge", "Shown as fraction")
	m.Register("mysql.performance.qcache_free_blocks", "gauge", "Shown as block")
	m.Register("mysql.performance.qcache_free_memory", "gauge", "Shown as byte")
	m.Register("mysql.performance.qcache_hits", "gauge", "Shown as hit")
	m.Register("mysql.performance.qcache_inserts", "gauge", "Shown as query")
	m.Register("mysql.performance.qcache_lowmem_prunes", "gauge", "Shown as query")
	m.Register("mysql.performance.qcache_not_cached", "gauge", "Shown as query")
	m.Register("mysql.performance.qcache_queries_in_cache", "gauge", "Shown as query")
	m.Register("mysql.performance.qcache_size", "gauge", "Shown as byte")
	m.Register("mysql.performance.qcache_total_blocks", "gauge", "Shown as block")
	m.Register("mysql.performance.queries", "gauge", "Shown as query")
	m.Register("mysql.performance.query_run_time.avg", "gauge", "Shown as microsecond")
	m.Register("mysql.performance.questions", "gauge", "Shown as query")
	m.Register("mysql.performance.select_full_join", "gauge", "Shown as operation")
	m.Register("mysql.performance.select_full_range_join", "gauge", "Shown as operation")
	m.Register("mysql.performance.select_range", "gauge", "Shown as operation")
	m.Register("mysql.performance.select_range_check", "gauge", "Shown as operation")
	m.Register("mysql.performance.select_scan", "gauge", "Shown as operation")
	m.Register("mysql.performance.slow_queries", "gauge", "Shown as query")
	m.Register("mysql.performance.sort_merge_passes", "gauge", "Shown as operation")
	m.Register("mysql.performance.sort_range", "gauge", "Shown as operation")
	m.Register("mysql.performance.sort_rows", "gauge", "Shown as operation")
	m.Register("mysql.performance.sort_scan", "gauge", "Shown as operation")
	m.Register("mysql.performance.table_cache_hits", "gauge", "Shown as hit")
	m.Register("mysql.performance.table_cache_misses", "gauge", "Shown as miss")
	m.Register("mysql.performance.table_locks_immediate", "gauge")
	m.Register("mysql.performance.table_locks_immediate.rate", "gauge")
	m.Register("mysql.performance.table_locks_waited", "gauge")
	m.Register("mysql.performance.table_locks_waited.rate", "gauge")
	m.Register("mysql.performance.table_open_cache", "gauge")
	m.Register("mysql.performance.thread_cache_size", "gauge", "Shown as byte")
	m.Register("mysql.performance.threads_cached", "gauge", "Shown as thread")
	m.Register("mysql.performance.threads_connected", "gauge", "Shown as connection")
	m.Register("mysql.performance.threads_created", "count", "Shown as thread")
	m.Register("mysql.performance.threads_running", "gauge", "Shown as thread")
	m.Register("mysql.performance.user_time", "gauge", "Shown as percent")
	m.Register("mysql.queries.count", "count", "Shown as query")
	m.Register("mysql.queries.errors", "count", "Shown as error")
	m.Register("mysql.queries.lock_time", "count", "Shown as nanosecond")
	m.Register("mysql.queries.no_good_index_used", "count", "Shown as query")
	m.Register("mysql.queries.no_index_used", "count", "Shown as query")
	m.Register("mysql.queries.rows_affected", "count", "Shown as row")
	m.Register("mysql.queries.rows_sent", "count", "Shown as row")
	m.Register("mysql.queries.select_full_join", "count")
	m.Register("mysql.queries.select_scan", "count")
	m.Register("mysql.queries.time", "count", "Shown as nanosecond")
	m.Register("mysql.replication.replicas_connected", "gauge")
	m.Register("mysql.replication.seconds_behind_master", "gauge", "Shown as second")
	m.Register("mysql.replication.seconds_behind_source", "gauge", "Shown as second")
	m.Register("mysql.replication.slave_running", "gauge")
	m.Register("mysql.replication.slaves_connected", "gauge")

}

func init() {
	registerMetric()
	i18n.SetLangStrings(langStrings)
}
