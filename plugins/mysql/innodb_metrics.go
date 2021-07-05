package mysql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/klog"
)

func NewInnoDBMetrics(db *sql.DB) *InnoDBMetrics {
	return &InnoDBMetrics{db: db}
}

type InnoDBMetrics struct {
	db *sql.DB
}

func (p *InnoDBMetrics) getStatsFromInnodbStatus() (map[string]string, error) {
	var innodb_status_text string
	_, err := queryRow(p.db, "SHOW /*!50000 ENGINE*/ INNODB STATUS", func() []interface{} {
		return []interface{}{
			new(string),
			new(string),
			&innodb_status_text,
		}
	})

	if err != nil {
		klog.Warningf("Unicode error while getting INNODB status (typically harmless, but if this warning is frequent metric collection could be impacted): %s", err)
		return nil, err
	}

	results := map[string]int64{}

	// Here we now parse InnoDB STATUS one line at a time
	// This is heavily inspired by the Percona monitoring plugins work
	txn_seen := false
	prev_line := ""
	// Only return aggregated buffer pool metrics
	buffer_id := int64(-1)
	for _, line := range strings.Split(innodb_status_text, "\n") {
		line = strings.TrimSpace(line)
		row := strings.Fields(line)
		for k, v := range row {
			row[k] = strings.Trim(v, ",;[]")
		}

		if strings.HasPrefix(line, "---BUFFER POOL") {
			buffer_id = Int64(row[2])
		}

		// SEMAPHORES
		if strings.HasPrefix(line, "Mutex spin waits") {
			// Mutex spin waits 79626940, rounds 157459864, OS waits 698719
			// Mutex spin waits 0, rounds 247280272495, OS waits 316513438
			results["Innodb_mutex_spin_waits"] = Int64(row[3])
			results["Innodb_mutex_spin_rounds"] = Int64(row[5])
			results["Innodb_mutex_os_waits"] = Int64(row[8])
		} else if strings.HasPrefix(line, "RW-shared spins") && strings.Index(line, ";") > 0 {
			// RW-shared spins 3859028, OS waits 2100750; RW-excl spins
			// 4641946, OS waits 1530310
			results["Innodb_s_lock_spin_waits"] = Int64(row[2])
			results["Innodb_x_lock_spin_waits"] = Int64(row[8])
			results["Innodb_s_lock_os_waits"] = Int64(row[5])
			results["Innodb_x_lock_os_waits"] = Int64(row[11])
		} else if strings.HasPrefix(line, "RW-shared spins") && strings.Index(line, "; RW-excl spins") == -1 {
			// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
			// RW-shared spins 604733, rounds 8107431, OS waits 241268
			results["Innodb_s_lock_spin_waits"] = Int64(row[2])
			results["Innodb_s_lock_spin_rounds"] = Int64(row[4])
			results["Innodb_s_lock_os_waits"] = Int64(row[7])
		} else if strings.HasPrefix(line, "RW-excl spins") {
			// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
			// RW-excl spins 604733, rounds 8107431, OS waits 241268
			results["Innodb_x_lock_spin_waits"] = Int64(row[2])
			results["Innodb_x_lock_spin_rounds"] = Int64(row[4])
			results["Innodb_x_lock_os_waits"] = Int64(row[7])
		} else if strings.Index(line, "seconds the semaphore:") > 0 {
			// --Thread 907205 has waited at handler/ha_innodb.cc line 7156 for 1.00 seconds the semaphore:
			results["Innodb_semaphore_waits"] += 1
			results["Innodb_semaphore_wait_time"] += int64(Float64(row[9])) * 1000

			// TRANSACTIONS
		} else if strings.HasPrefix(line, "Trx id counter") {
			// The beginning of the TRANSACTIONS section: start counting
			// transactions
			// Trx id counter 0 1170664159
			// Trx id counter 861B144C
			txn_seen = true
		} else if strings.HasPrefix(line, "History list length") {
			// History list length 132
			results["Innodb_history_list_length"] = Int64(row[3])

		} else if txn_seen && strings.HasPrefix(line, "---TRANSACTION") {
			// ---TRANSACTION 0, not started, process no 13510, OS thread id 1170446656
			results["Innodb_current_transactions"] += 1
			if strings.Index(line, "ACTIVE") > 0 {
				results["Innodb_active_transactions"] += 1
			}
		} else if strings.Index(line, "read views open inside InnoDB") > 0 {
			// 1 read views open inside InnoDB
			results["Innodb_read_views"] = Int64(row[0])
		} else if strings.HasPrefix(line, "mysql tables in use") {
			// mysql tables in use 2, locked 2
			results["Innodb_tables_in_use"] += Int64(row[4])
			results["Innodb_locked_tables"] += Int64(row[6])
		} else if txn_seen && strings.Index(line, "lock struct(s)") > 0 {
			// 23 lock struct(s), heap size 3024, undo log entries 27
			// LOCK WAIT 12 lock struct(s), heap size 3024, undo log entries 5
			// LOCK WAIT 2 lock struct(s), heap size 368
			if strings.HasPrefix(line, "LOCK WAIT") {
				results["Innodb_lock_structs"] += Int64(row[2])
				results["Innodb_locked_transactions"] += 1
			} else if strings.HasPrefix(line, "ROLLING BACK") {
				// ROLLING BACK 127539 lock struct(s), heap size 15201832,
				// 4411492 row lock(s), undo log entries 1042488
				results["Innodb_lock_structs"] += Int64(row[2])
			} else {
				results["Innodb_lock_structs"] += Int64(row[0])
			}

			// FILE I/O
		} else if strings.Index(line, " OS file reads, ") > 0 {
			// 8782182 OS file reads, 15635445 OS file writes, 947800 OS
			// fsyncs
			results["Innodb_os_file_reads"] = Int64(row[0])
			results["Innodb_os_file_writes"] = Int64(row[4])
			results["Innodb_os_file_fsyncs"] = Int64(row[8])
		} else if strings.HasPrefix(line, "Pending normal aio reads:") {
			if len(row) == 8 {
				// (len(row) == 8)  Pending normal aio reads: 0, aio writes: 0,
				results["Innodb_pending_normal_aio_reads"] = Int64(row[4])
				results["Innodb_pending_normal_aio_writes"] = Int64(row[7])
			} else if len(row) == 14 {
				// (len(row) == 14) Pending normal aio reads: 0 [0, 0] , aio writes: 0 [0, 0] ,
				results["Innodb_pending_normal_aio_reads"] = Int64(row[4])
				results["Innodb_pending_normal_aio_writes"] = Int64(row[10])
			} else if len(row) == 16 {
				// (len(row) == 16) Pending normal aio reads: [0, 0, 0, 0] , aio writes: [0, 0, 0, 0] ,
				if _are_values_numeric(row[4:8]) && _are_values_numeric(row[11:15]) {
					results["Innodb_pending_normal_aio_reads"] = Int64(row[4]) + Int64(row[5]) + Int64(row[6]) + Int64(row[7])
					results["Innodb_pending_normal_aio_writes"] = Int64(row[11]) + Int64(row[12]) + Int64(row[13]) + Int64(row[14])

					// (len(row) == 16) Pending normal aio reads: 0 [0, 0, 0, 0] , aio writes: 0 [0, 0] ,
				} else if _are_values_numeric(row[4:9]) && _are_values_numeric(row[12:15]) {
					results["Innodb_pending_normal_aio_reads"] = Int64(row[4])
					results["Innodb_pending_normal_aio_writes"] = Int64(row[12])
				} else {
					klog.Warningf("Can't parse result line %s", line)
				}
			} else if len(row) == 18 {
				// (len(row) == 18) Pending normal aio reads: 0 [0, 0, 0, 0] , aio writes: 0 [0, 0, 0, 0] ,
				results["Innodb_pending_normal_aio_reads"] = Int64(row[4])
				results["Innodb_pending_normal_aio_writes"] = Int64(row[12])
			} else if len(row) == 22 {
				// (len(row) == 22)
				// Pending normal aio reads: 0 [0, 0, 0, 0, 0, 0, 0, 0] , aio writes: 0 [0, 0, 0, 0] ,
				results["Innodb_pending_normal_aio_reads"] = Int64(row[4])
				results["Innodb_pending_normal_aio_writes"] = Int64(row[16])
			}
		} else if strings.HasPrefix(line, "ibuf aio reads") {
			//  ibuf aio reads: 0, log i/o"s: 0, sync i/o"s: 0
			//  or ibuf aio reads:, log i/o"s:, sync i/o"s{
			if len(row) == 10 {
				results["Innodb_pending_ibuf_aio_reads"] = Int64(row[3])
				results["Innodb_pending_aio_log_ios"] = Int64(row[6])
				results["Innodb_pending_aio_sync_ios"] = Int64(row[9])
			} else if len(row) == 7 {
				results["Innodb_pending_ibuf_aio_reads"] = 0
				results["Innodb_pending_aio_log_ios"] = 0
				results["Innodb_pending_aio_sync_ios"] = 0
			}
		} else if strings.HasPrefix(line, "Pending flushes (fsync)") {
			// Pending flushes (fsync) log: 0; buffer pool: 0
			results["Innodb_pending_log_flushes"] = Int64(row[4])
			results["Innodb_pending_buffer_pool_flushes"] = Int64(row[7])

			// INSERT BUFFER AND ADAPTIVE HASH INDEX
		} else if strings.HasPrefix(line, "Ibuf for space 0: size ") {
			// Older InnoDB code seemed to be ready for an ibuf per tablespace.  It
			// had two lines in the output.  Newer has just one line, see below.
			// Ibuf for space 0: size 1, free list len 887, seg size 889, is not empty
			// Ibuf for space 0: size 1, free list len 887, seg size 889,
			results["Innodb_ibuf_size"] = Int64(row[5])
			results["Innodb_ibuf_free_list"] = Int64(row[9])
			results["Innodb_ibuf_segment_size"] = Int64(row[12])
		} else if strings.HasPrefix(line, "Ibuf: size ") {
			// Ibuf: size 1, free list len 4634, seg size 4636,
			results["Innodb_ibuf_size"] = Int64(row[2])
			results["Innodb_ibuf_free_list"] = Int64(row[6])
			results["Innodb_ibuf_segment_size"] = Int64(row[9])

			if strings.Index(line, "merges") > -1 {
				results["Innodb_ibuf_merges"] = Int64(row[10])
			}
		} else if strings.Index(line, ", delete mark ") > 0 && strings.HasPrefix(prev_line, "merged operations:") {
			// Output of show engine innodb status has changed in 5.5
			// merged operations{
			// insert 593983, delete mark 387006, delete 73092
			results["Innodb_ibuf_merged_inserts"] = Int64(row[1])
			results["Innodb_ibuf_merged_delete_marks"] = Int64(row[4])
			results["Innodb_ibuf_merged_deletes"] = Int64(row[6])
			results["Innodb_ibuf_merged"] = results["Innodb_ibuf_merged_inserts"] + results["Innodb_ibuf_merged_delete_marks"] + results["Innodb_ibuf_merged_deletes"]
		} else if strings.Index(line, " merged recs, ") > 0 {
			// 19817685 inserts, 19817684 merged recs, 3552620 merges
			results["Innodb_ibuf_merged_inserts"] = Int64(row[0])
			results["Innodb_ibuf_merged"] = Int64(row[2])
			results["Innodb_ibuf_merges"] = Int64(row[5])
		} else if strings.HasPrefix(line, "Hash table size ") {
			// In some versions of InnoDB, the used cells is omitted.
			// Hash table size 4425293, used cells 4229064, ....
			// Hash table size 57374437, node heap has 72964 buffer(s) <--
			// no used cells
			results["Innodb_hash_index_cells_total"] = Int64(row[3])
			if strings.Index(line, "used cells") > 0 {
				results["Innodb_hash_index_cells_used"] = Int64(row[6])
			} else {
				results["Innodb_hash_index_cells_used"] = 0
			}

			// LOG
		} else if strings.Index(line, " log i/o's done, ") > 0 {
			// 3430041 log i/o"s done, 17.44 log i/o"s/second
			// 520835887 log i/o"s done, 17.28 log i/o"s/second, 518724686
			// syncs, 2980893 checkpoints
			results["Innodb_log_writes"] = Int64(row[0])
		} else if strings.Index(line, " pending log writes, ") > 0 {
			// 0 pending log writes, 0 pending chkp writes
			results["Innodb_pending_log_writes"] = Int64(row[0])
			results["Innodb_pending_checkpoint_writes"] = Int64(row[4])
		} else if strings.HasPrefix(line, "Log sequence number") {
			// This number is NOT printed in hex in InnoDB plugin.
			// Log sequence number 272588624
			results["Innodb_lsn_current"] = Int64(row[3])
		} else if strings.HasPrefix(line, "Log flushed up to") {
			// This number is NOT printed in hex in InnoDB plugin.
			// Log flushed up to   272588624
			results["Innodb_lsn_flushed"] = Int64(row[4])
		} else if strings.HasPrefix(line, "Last checkpoint at") {
			// Last checkpoint at  272588624
			results["Innodb_lsn_last_checkpoint"] = Int64(row[3])

			// BUFFER POOL AND MEMORY
		} else if strings.HasPrefix(line, "Total memory allocated") && strings.Index(line, "in additional pool allocated") > 0 {
			// Total memory allocated 29642194944; in additional pool allocated 0
			// Total memory allocated by read views 96
			results["Innodb_mem_total"] = Int64(row[3])
			results["Innodb_mem_additional_pool"] = Int64(row[8])
		} else if strings.HasPrefix(line, "Adaptive hash index ") {
			//   Adaptive hash index 1538240664     (186998824 + 1351241840)
			results["Innodb_mem_adaptive_hash"] = Int64(row[3])
		} else if strings.HasPrefix(line, "Page hash           ") {
			//   Page hash           11688584
			results["Innodb_mem_page_hash"] = Int64(row[2])
		} else if strings.HasPrefix(line, "Dictionary cache    ") {
			//   Dictionary cache    145525560      (140250984 + 5274576)
			results["Innodb_mem_dictionary"] = Int64(row[2])
		} else if strings.HasPrefix(line, "File system         ") {
			//   File system         313848         (82672 + 231176)
			results["Innodb_mem_file_system"] = Int64(row[2])
		} else if strings.HasPrefix(line, "Lock system         ") {
			//   Lock system         29232616       (29219368 + 13248)
			results["Innodb_mem_lock_system"] = Int64(row[2])
		} else if strings.HasPrefix(line, "Recovery system     ") {
			//   Recovery system     0      (0 + 0)
			results["Innodb_mem_recovery_system"] = Int64(row[2])
		} else if strings.HasPrefix(line, "Threads             ") {
			//   Threads             409336         (406936 + 2400)
			results["Innodb_mem_thread_hash"] = Int64(row[1])
		} else if strings.HasPrefix(line, "Buffer pool size ") {
			// The " " after size is necessary to avoid matching the wrong line{
			// Buffer pool size        1769471
			// Buffer pool size, bytes 28991012864
			if buffer_id == -1 {
				results["Innodb_buffer_pool_pages_total"] = Int64(row[3])
			}
		} else if strings.HasPrefix(line, "Free buffers") {
			// Free buffers            0
			if buffer_id == -1 {
				results["Innodb_buffer_pool_pages_free"] = Int64(row[2])
			}
		} else if strings.HasPrefix(line, "Database pages") {
			// Database pages          1696503
			if buffer_id == -1 {
				results["Innodb_buffer_pool_pages_data"] = Int64(row[2])
			}

		} else if strings.HasPrefix(line, "Modified db pages") {
			// Modified db pages       160602
			if buffer_id == -1 {
				results["Innodb_buffer_pool_pages_dirty"] = Int64(row[3])
			}
		} else if strings.HasPrefix(line, "Pages read ahead") {
			// Must do this BEFORE the next test, otherwise it"ll get fooled by this
			// line from the new plugin{
			// Pages read ahead 0.00/s, evicted without access 0.06/s
			klog.Infof("-")
		} else if strings.HasPrefix(line, "Pages read") {
			// Pages read 15240822, created 1770238, written 21705836
			if buffer_id == -1 {
				results["Innodb_pages_read"] = Int64(row[2])
				results["Innodb_pages_created"] = Int64(row[4])
				results["Innodb_pages_written"] = Int64(row[6])
			}

			// ROW OPERATIONS
		} else if strings.HasPrefix(line, "Number of rows inserted") {
			// Number of rows inserted 50678311, updated 66425915, deleted
			// 20605903, read 454561562
			results["Innodb_rows_inserted"] = Int64(row[4])
			results["Innodb_rows_updated"] = Int64(row[6])
			results["Innodb_rows_deleted"] = Int64(row[8])
			results["Innodb_rows_read"] = Int64(row[10])
		} else if strings.Index(line, " queries inside InnoDB, ") > 0 {
			// 0 queries inside InnoDB, 0 queries in queue
			results["Innodb_queries_inside"] = Int64(row[0])
			results["Innodb_queries_queued"] = Int64(row[4])
		}

		prev_line = line
	} // end for

	if a, ok := results["Innodb_lsn_current"]; ok {
		if b, ok := results["Innodb_lsn_last_checkpoint"]; ok {
			results["Innodb_checkpoint_age"] = a - b
		}
	}

	out := map[string]string{}
	for k, v := range results {
		out[k] = strconv.FormatInt(v, 10)
	}

	return out, nil
}

var (
	innodb_keys = []string{
		"Innodb_page_size",
		"Innodb_buffer_pool_pages_data",
		"Innodb_buffer_pool_pages_dirty",
		"Innodb_buffer_pool_pages_total",
		"Innodb_buffer_pool_pages_free",
	}
)

func (p *InnoDBMetrics) process_innodb_stats(results map[string]interface{}, options *Options, metrics map[string]interface{}) error {
	values := map[string]float64{}

	for _, k := range innodb_keys {
		if v, ok := results[k]; !ok {
			return fmt.Errorf("%s is inavailable, unable to process innodb stats", k)
		} else {
			values[k] = Float64(v)
		}
	}

	innodb_page_size := values["Innodb_page_size"]
	innodb_buffer_pool_pages_used := values["Innodb_buffer_pool_pages_total"] - values["Innodb_buffer_pool_pages_free"]

	if _, ok := results["Innodb_buffer_pool_bytes_data"]; !ok {
		results["Innodb_buffer_pool_bytes_data"] = values["Innodb_buffer_pool_pages_data"] * innodb_page_size
	}

	if _, ok := results["Innodb_buffer_pool_bytes_dirty"]; !ok {
		results["Innodb_buffer_pool_bytes_dirty"] = values["Innodb_buffer_pool_pages_dirty"] * innodb_page_size
	}

	if _, ok := results["Innodb_buffer_pool_bytes_free"]; !ok {
		results["Innodb_buffer_pool_bytes_free"] = values["Innodb_buffer_pool_pages_free"] * innodb_page_size
	}

	if _, ok := results["Innodb_buffer_pool_bytes_total"]; !ok {
		results["Innodb_buffer_pool_bytes_total"] = values["Innodb_buffer_pool_pages_total"] * innodb_page_size
	}

	if _, ok := results["Innodb_buffer_pool_pages_utilization"]; !ok {
		results["Innodb_buffer_pool_pages_utilization"] = innodb_buffer_pool_pages_used / values["Innodb_buffer_pool_pages_total"]
	}

	if _, ok := results["Innodb_buffer_pool_bytes_used"]; !ok {
		results["Innodb_buffer_pool_bytes_used"] = innodb_buffer_pool_pages_used * innodb_page_size
	}

	if options.ExtraInnodbMetrics {
		klog.V(6).Infof("Collecting Extra Innodb Metrics")
		// TODO:
		//metrics.update(OPTIONAL_INNODB_VARS)
	}
	return nil
}

func Int64(a interface{}) int64 {
	if s, ok := a.(string); ok {
		i, _ := strconv.ParseInt(s, 10, 0)
		return i
	}
	return int64(a.(float64))
}

func Float64(a interface{}) float64 {
	if s, ok := a.(string); ok {
		f, _ := strconv.ParseFloat(s, 0)
		return f
	}
	return a.(float64)
}

func _are_values_numeric(in []string) bool {
	for _, v := range in {
		for _, c := range []byte(v) {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
