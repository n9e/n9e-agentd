package processor

var (
	valueFuncs = map[string]func(float64) float64{
		"b_to_kb": func(v float64) float64 {
			return v / 1024
		},
		"kb_to_b": func(v float64) float64 {
			return v * 1024
		},
		"idle_to_util": func(v float64) float64 {
			return 100 - v
		},
		"fraction_to_percent": func(v float64) float64 {
			return v * 100
		},
		"idle_fraction_to_util_percent": func(v float64) float64 {
			return 100 - (v * 100)
		},
	}

	_metricPolicies = map[string]*MetricPolicy{
		"ntp_offset":                              &MetricPolicy{NewName: "system_ntp_offset"},
		"system_cpu_context_switches":             &MetricPolicy{NewName: "system_cpu_switches"},
		"system_cpu_stolen":                       &MetricPolicy{NewName: "system_cpu_steal"},
		"system_disk_in_use":                      &MetricPolicy{NewName: "system_disk_used_percent", ValueFunc: "fraction_to_percent"},
		"system_disk_read_time_pct":               &MetricPolicy{NewName: "system_disk_read_time_percent"},
		"system_disk_write_time_pct":              &MetricPolicy{NewName: "system_disk_write_time_percent"},
		"system_fs_file_handles_allocated":        &MetricPolicy{NewName: "system_files_allocated"},
		"system_fs_file_handles_allocated_unused": &MetricPolicy{NewName: "system_files_left"},
		"system_fs_file_handles_max":              &MetricPolicy{NewName: "system_files_max"},
		"system_fs_file_handles_in_use":           &MetricPolicy{NewName: "system_files_used_percent", ValueFunc: "fraction_to_percent"},
		"system_fs_file_handles_used":             &MetricPolicy{NewName: "system_files_used"},
		"system_fs_inodes_in_use":                 &MetricPolicy{NewName: "system_disk_inodes_used_percent", ValueFunc: "fraction_to_percent"},
		"system_fs_inodes_free":                   &MetricPolicy{NewName: "system_disk_inodes_free"},
		"system_fs_inodes_total":                  &MetricPolicy{NewName: "system_disk_inodes_total"},
		"system_fs_inodes_used":                   &MetricPolicy{NewName: "system_disk_inodes_used"},
		"system_io_avg_q_sz":                      &MetricPolicy{NewName: "system_io_avgqu_sz"},
		"system_io_avg_rq_sz":                     &MetricPolicy{NewName: "system_io_avgrq_sz"},
		"system_io_r_s":                           &MetricPolicy{NewName: "system_io_read_request"},
		"system_io_rkb_s":                         &MetricPolicy{NewName: "system_io_read_bytes", ValueFunc: "kb_to_b"},
		"system_io_w_s":                           &MetricPolicy{NewName: "system_io_write_request"},
		"system_io_wkb_s":                         &MetricPolicy{NewName: "system_io_write_bytes", ValueFunc: "kb_to_b"},
		"system_mem_pct_usable":                   &MetricPolicy{NewName: "system_mem_used_percent", ValueFunc: "idle_fraction_to_util_percent"},
		"system_swap_pct_free":                    &MetricPolicy{NewName: "system_swap_used_percent", ValueFunc: "idle_fraction_to_util_percent"},
		"system_mem_pagefile_pct_free":            &MetricPolicy{NewName: "system_mem_pagefile_percent", ValueFunc: "idle_fraction_to_util_percent"},
		"system_disk_free":                        &MetricPolicy{NewName: "system_disk_bytes_free", ValueFunc: "kb_to_b"},
		"system_disk_used":                        &MetricPolicy{NewName: "system_disk_bytes_used", ValueFunc: "kb_to_b"},
		"system_disk_total":                       &MetricPolicy{NewName: "system_disk_bytes_total", ValueFunc: "kb_to_b"},
		"proc_cpu_total":                          &MetricPolicy{NewName: "proc_cpu_util"},
		"system_cpu_idle":                         &MetricPolicy{NewName: "system_cpu_util", ValueFunc: "idle_to_util"},
	}
)
