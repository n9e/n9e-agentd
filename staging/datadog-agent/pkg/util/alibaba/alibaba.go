// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package alibaba

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	httputils "github.com/DataDog/datadog-agent/pkg/util/http"
	"github.com/n9e/n9e-agentd/pkg/config"
)

// declare these as vars not const to ease testing
var (
	metadataURL = "http://100.100.100.200"
	timeout     = 300 * time.Millisecond

	// CloudProviderName contains the inventory name of for EC2
	CloudProviderName = "Alibaba"
)

// IsRunningOn returns true if the agent is running on Alibaba
func IsRunningOn(ctx context.Context) bool {
	if _, err := GetHostAlias(ctx); err == nil {
		return true
	}
	return false
}

// GetHostAlias returns the VM ID from the Alibaba Metadata api
func GetHostAlias(ctx context.Context) (string, error) {
	if !config.IsCloudProviderEnabled(CloudProviderName) {
		return "", fmt.Errorf("cloud provider is disabled by configuration")
	}
	res, err := getResponseWithMaxLength(ctx, metadataURL+"/latest/meta-data/instance-id",
		config.C.MetadataEndpointsMaxHostnameSize)
	if err != nil {
		return "", fmt.Errorf("Alibaba HostAliases: unable to query metadata endpoint: %s", err)
	}
	return res, err
}

// GetNTPHosts returns the NTP hosts for Alibaba if it is detected as the cloud provider, otherwise an empty array.
// These are their public NTP servers, as Alibaba uses two different types of private/internal networks for their cloud
// machines and we can't be sure those servers are always accessible for every customer on every network type.
// Docs: https://www.alibabacloud.com/help/doc-detail/92704.htm
func GetNTPHosts(ctx context.Context) []string {
	if IsRunningOn(ctx) {
		return []string{
			"ntp.aliyun.com", "ntp1.aliyun.com", "ntp2.aliyun.com", "ntp3.aliyun.com",
			"ntp4.aliyun.com", "ntp5.aliyun.com", "ntp6.aliyun.com", "ntp7.aliyun.com",
		}
	}

	return nil
}

func getResponseWithMaxLength(ctx context.Context, endpoint string, maxLength int) (string, error) {
	result, err := getResponse(ctx, endpoint)
	if err != nil {
		return result, err
	}
	if len(result) > maxLength {
		return "", fmt.Errorf("%v gave a response with length > to %v", endpoint, maxLength)
	}
	return result, err
}

func getResponse(ctx context.Context, url string) (string, error) {
	client := http.Client{
		Transport: httputils.CreateHTTPTransport(),
		Timeout:   timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code %d trying to GET %s", res.StatusCode, url)
	}

	defer res.Body.Close()
	all, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading response from alibaba metadata endpoint: %s", err)
	}

	return string(all), nil
}
