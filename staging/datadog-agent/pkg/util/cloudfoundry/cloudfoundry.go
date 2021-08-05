// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package cloudfoundry

import (
	"os"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	"k8s.io/klog/v2"
)

// Define alias in order to mock in the tests
var getFqdn = util.Fqdn

// GetHostAliases returns the host aliases from Cloud Foundry
func GetHostAliases() ([]string, error) {
	if !config.C.CloudFoundry {
		klog.V(5).Infof("cloud_foundry is not enabled in the conf: no cloudfoudry host alias")
		return nil, nil
	}

	aliases := []string{}

	// Always send the bosh_id if specified
	boshID := config.C.BoshID
	if boshID != "" {
		aliases = append(aliases, boshID)
	}

	hostname, _ := os.Hostname()
	fqdn := getFqdn(hostname)

	if config.C.CfOSHostnameAliasing {
		// If set, send os hostname and fqdn as additional aliases
		aliases = append(aliases, hostname)
		if fqdn != hostname {
			aliases = append(aliases, fqdn)
		}
	} else if boshID == "" {
		// If no hostname aliasing, only alias with fqdn if no bosh id
		aliases = append(aliases, fqdn)
	}

	return aliases, nil
}