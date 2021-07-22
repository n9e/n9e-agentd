// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.
// +build !windows

package disk

import (
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
)

const (
	checkName   = "disk"
	diskMetric  = "system.disk.%s"
	inodeMetric = "system.fs.inodes.%s"
)

type diskConfig struct {
	useMount             bool
	excludedFilesystems  []string
	excludedDisks        []string
	excludedDiskRe       *regexp.Regexp
	tagByFilesystem      bool
	excludedMountpointRe *regexp.Regexp
	allPartitions        bool
	deviceTagRe          map[*regexp.Regexp][]string
}

func (c *Check) excludeDisk(mountpoint, device, fstype string) bool {

	// Hack for NFS secure mounts
	// Secure mounts might look like this: '/mypath (deleted)', we should
	// ignore all the bits not part of the mountpoint name. Take also into
	// account a space might be in the mountpoint.
	mountpoint = strings.Split(mountpoint, " ")[0]

	nameEmpty := device == "" || device == "none"

	// allow empty names if `all_partitions` is `yes` so we can evaluate mountpoints
	if nameEmpty {
		if !c.cfg.allPartitions {
			return true
		}
	} else {
		// I don't why I we do this only if the device name is not empty
		// This is useful only when `all_partitions` is true and `exclude_disk_re` matches empty strings or `excluded_devices` contains the device

		// device is listed in `excluded_disks`
		if stringSliceContain(c.cfg.excludedDisks, device) {
			return true
		}

		// device name matches `excluded_disk_re`
		if c.cfg.excludedDiskRe != nil && c.cfg.excludedDiskRe.MatchString(device) {
			return true
		}
	}

	// fs is listed in `excluded_filesystems`
	if stringSliceContain(c.cfg.excludedFilesystems, fstype) {
		return true
	}

	// device mountpoint matches `excluded_mountpoint_re`
	if c.cfg.excludedMountpointRe != nil && c.cfg.excludedMountpointRe.MatchString(mountpoint) {
		return true
	}

	// all good, don't exclude the disk
	return false
}

type diskInstanceConfig struct {
	UseMount             bool              `json:"use_mount"`
	ExcludedFilesystems  []string          `json:"excluded_filesystems"`
	ExcludedDisks        []string          `json:"excluded_disks"`
	ExcludedDiskRe       string            `json:"excluded_disk_re"`
	TagByFilesystem      bool              `json:"tag_by_filesystem"`
	ExcludedMountpointRe string            `json:"excluded_mountpoint_re"`
	AllPartitions        bool              `json:"all_partitions"`
	DeviceTagRe          map[string]string `json:"device_tag_re"`
}

func (c *Check) instanceConfigure(data integration.Data) error {
	conf := diskInstanceConfig{}
	c.cfg = &diskConfig{}
	err := yaml.Unmarshal(data, &conf)
	if err != nil {
		return err
	}

	c.cfg.useMount = conf.UseMount
	// Force exclusion of CDROM (iso9660) from disk check
	c.cfg.excludedFilesystems = append(conf.ExcludedFilesystems, "iso9660")
	c.cfg.excludedDisks = conf.ExcludedDisks
	if conf.ExcludedDiskRe != "" {
		c.cfg.excludedDiskRe, err = regexp.Compile(conf.ExcludedDiskRe)
		if err != nil {
			return err
		}
	}

	c.cfg.tagByFilesystem = conf.TagByFilesystem

	if conf.ExcludedMountpointRe != "" {

		c.cfg.excludedMountpointRe, err = regexp.Compile(conf.ExcludedMountpointRe)
		if err != nil {
			return err
		}
	}

	c.cfg.allPartitions = conf.AllPartitions

	c.cfg.deviceTagRe = make(map[*regexp.Regexp][]string)
	for reString, tags := range conf.DeviceTagRe {
		re, err := regexp.Compile(reString)
		if err != nil {
			return err
		}
		c.cfg.deviceTagRe[re] = strings.Split(tags, ",")
	}

	return nil
}

func stringSliceContain(slice []string, x string) bool {
	for _, e := range slice {
		if e == x {
			return true
		}
	}
	return false
}

func (c *Check) applyDeviceTags(device, mountpoint string, tags []string) []string {
	// apply device/mountpoint specific tags
	for re, deviceTags := range c.cfg.deviceTagRe {
		if re == nil {
			continue
		}
		if re.MatchString(device) || (mountpoint != "" && re.MatchString(mountpoint)) {
			for _, tag := range deviceTags {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

func diskFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func init() {
	core.RegisterCheck(checkName, diskFactory)
}
