package util

import "os"

// TODO: get hostname from command
func GetHostname(hostname string) (string, error) {
	if hostname != "" {
		return hostname, nil
	}
	return os.Hostname()
}
