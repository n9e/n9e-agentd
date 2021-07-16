package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/n9e/n9e-agentd/pkg/util"
	"k8s.io/klog/v2"
)

func init() {
	defaultTransformer = NewTransformer()
	defaultTransformer.SetMetric(transformMetricMap)
}

var (
	defaultTransformer *Transformer
)

type Transformer struct {
	metrics map[string]string
}

func NewTransformer() *Transformer { return &Transformer{metrics: make(map[string]string)} }

func (t *Transformer) SetMetric(config map[string]string) {
	for k, v := range config {
		t.metrics[util.SanitizeMetric(k)] = v
	}
}

func (t *Transformer) Metric(name string) string {
	name = util.SanitizeMetric(name)
	if v, ok := t.metrics[name]; ok {
		return v
	}

	return name
}

func (t *Transformer) SetMetricFromFile(file string) error {
	if file == "" {
		return nil
	}

	klog.V(6).Infof("load transformer metric from %s", file)

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	config := map[string]string{}
	if err := json.Unmarshal(b, &config); err != nil {
		return nil
	}

	t.SetMetric(config)

	return nil
}

func TransformMetric(metric string) string             { return defaultTransformer.Metric(metric) }
func TransformMapTags(tags []string) map[string]string { return util.SanitizeMapTags(tags) }
func TransformTags(tags []string) []string             { return util.SanitizeTags(tags) }

var (
	transformMetricMap = map[string]string{
		"ntp_offset":                              "system_ntp_offset",
		"system_cpu_context_switches":             "system_cpu_switches",
		"system_cpu_stolen":                       "system_cpu_steal",
		"system_disk_in_use":                      "disk_bytes_used_percent",
		"system_disk_read_time_pct":               "system_disk_read_time_percent",
		"system_disk_write_time_pct":              "system_disk_write_time_percent",
		"system_fs_file_handles_allocated":        "system_files_allocated",
		"system_fs_file_handles_allocated_unused": "system_files_left",
		"system_fs_file_handles_max":              "system_files_max",
		"system_fs_file_handles_in_use":           "system_files_used_percent",
		"system_fs_file_handles_used":             "system_files_used",
		"system_fs_inodes_free":                   "system_disk_inodes_free",
		"system_fs_inodes_in_use":                 "system_disk_inodes_used_percent",
		"system_fs_inodes_total":                  "system_disk_inodes_total",
		"system_fs_inodes_used":                   "system_disk_inodes_used",
		"system_io_avg_q_sz":                      "system_io_avgqu_sz",
		"system_io_avg_rq_sz":                     "system_io_avgrq_sz",
		"system_io_r_s":                           "system_io_read_request",
		"system_io_rb_s":                          "system_io_read_bytes",
		"system_io_w_s":                           "system_io_write_request",
		"system_io_wb_s":                          "system_io_write_bytes",
		"system_mem_pct_usable":                   "system_mem_free_percent",
		"system_mem_pct_used":                     "system_mem_used_percent",
		"system_mem_usable":                       "system_mem_free",
		"system_swap_pct_free":                    "system_mem_free_percent",
	}
)
