package util

import (
	"fmt"
	"io/ioutil"
	"syscall"

	acl "github.com/hectane/go-acl"
	"golang.org/x/sys/windows"
	"k8s.io/klog/v2"
)

// SetupCoreDump enables core dumps and sets the core dump size limit based on configuration
func SetupCoreDump() error {
	return fmt.Errorf("Not supported on Windows")
}

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
