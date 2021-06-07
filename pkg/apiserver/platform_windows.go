// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package apiserver

import (
	"fmt"
	"io/ioutil"
	"syscall"

	acl "github.com/hectane/go-acl"
	"k8s.io/klog/v2"
	"golang.org/x/sys/windows"
)

var (
	wellKnownSidStrings = map[string]string{
		"Administrators": "S-1-5-32-544",
		"System":         "S-1-5-18",
		"Users":          "S-1-5-32-545",
	}
	wellKnownSids = make(map[string]*windows.SID)
)

func init() {
	for key, val := range wellKnownSidStrings {
		sid, err := windows.StringToSid(val)
		if err == nil {
			wellKnownSids[key] = sid
		}
	}
}

// lookupUsernameAndDomain obtains the username and domain for usid.
func lookupUsernameAndDomain(usid *syscall.SID) (username, domain string, e error) {
	username, domain, t, e := usid.LookupAccount("")
	if e != nil {
		return "", "", e
	}
	if t != syscall.SidTypeUser {
		return "", "", fmt.Errorf("user: should be user account type, not %d", t)
	}
	return username, domain, nil
}

// writes auth token(s) to a file with the same permissions as datadog.yaml
func saveAuthToken(token, tokenPath string) error {
	// get the current user
	var sidString string
	klog.Infof("Getting sidstring from user")
	tok, e := syscall.OpenCurrentProcessToken()
	if e != nil {
		klog.Warningf("Couldn't get process token %v", e)
		return e
	}
	defer tok.Close()
	user, e := tok.GetTokenUser()
	if e != nil {
		klog.Warningf("Couldn't get  token user %v", e)
		return e
	}
	sidString, e = user.User.Sid.String()
	if e != nil {
		klog.Warningf("Couldn't get  user sid string %v", e)
		return e
	}
	klog.Infof("Getting sidstring from current user")
	currUserSid, err := windows.StringToSid(sidString)
	if err != nil {
		klog.Warningf("Unable to get current user sid %v", err)
		return err
	}
	err = ioutil.WriteFile(tokenPath, []byte(token), 0755)
	if err == nil {
		err = acl.Apply(
			tokenPath,
			true,  // replace the file permissions
			false, // don't inherit
			acl.GrantSid(windows.GENERIC_ALL, wellKnownSids["Administrators"]),
			acl.GrantSid(windows.GENERIC_ALL, wellKnownSids["System"]),
			acl.GrantSid(windows.GENERIC_ALL, currUserSid))
		klog.Infof("Wrote auth token acl %v", err)
	}
	return err
}
