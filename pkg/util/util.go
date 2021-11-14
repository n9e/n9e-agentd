package util

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/klog/v2"
)

// SleepContext sleeps until the context is closed or the duration is reached.
func SleepContext(ctx context.Context, duration time.Duration) error {
	if duration == 0 {
		return nil
	}

	t := time.NewTimer(duration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}

func SanitizeMetric(s string) string {
	b := make([]byte, len(s))
	copy(b, s)
	for i, c := range b {
		if !((c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9')) {
			b[i] = '_'
		}
	}
	return string(b)
}

func SanitizeTag(s string) string {
	b := make([]byte, len(s))
	copy(b, s)
	for i, c := range b {
		if c == ':' {
			break
		}
		if !((c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9')) {
			b[i] = '_'
		}
	}
	return string(b)
}

func SanitizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	out := make([]string, len(tags))

	for i, tag := range tags {
		out[i] = SanitizeTag(tag)
	}

	return out
}

func SanitizeMapTag(s string) (string, string) {
	b := make([]byte, len(s))
	copy(b, s)

	var i int
	var c byte
	for i, c = range b {
		if c == ':' {
			break
		}
		if !((c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9')) {
			b[i] = '_'
		}
	}

	if c == ':' && i+1 < len(b) {
		return string(b[:i]), string(b[i+1:])
	}

	return string(b), ""
}

func SanitizeMapTags(tags []string) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	out := make(map[string]string)
	for _, tag := range tags {
		k, v := SanitizeMapTag(tag)
		out[k] = v
	}

	return out
}

func WriteRawJSON(statusCode int, object interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(output)
}

func ResolveRootPath(root string) (string, error) {
	if root != "" {
		return filepath.Abs(root)
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	// strip bin/
	if filepath.Base(dir) == "bin" {
		return filepath.Dir(dir), nil
	}

	return dir, nil

}

const (
	authTokenMinimalLen = 32
)

func FetchAuthToken(file string, tokenCreationAllowed bool) (string, error) {
	authTokenFile := file

	// Create a new token if it doesn't exist and if permitted by calling func
	if _, e := os.Stat(authTokenFile); os.IsNotExist(e) && tokenCreationAllowed {
		key := make([]byte, authTokenMinimalLen)
		_, e = rand.Read(key)
		if e != nil {
			return "", fmt.Errorf("can't create agent authentication token value: %s", e)
		}

		// Write the auth token to the auth token file (platform-specific)
		e = saveAuthToken(hex.EncodeToString(key), authTokenFile)
		if e != nil {
			return "", fmt.Errorf("error writing authentication token file on fs: %s", e)
		}
		klog.Infof("Saved a new authentication token to %s", authTokenFile)
	}
	// Read the token
	authTokenRaw, e := ioutil.ReadFile(authTokenFile)
	if e != nil {
		return "", fmt.Errorf("unable to read authentication token file: " + e.Error())
	}

	// Do some basic validation
	authToken := string(authTokenRaw)
	if len(authToken) < authTokenMinimalLen {
		return "", fmt.Errorf("invalid authentication token: must be at least %d characters in length", authTokenMinimalLen)
	}

	return authToken, nil
}

func validateAuthToken(authToken string) error {
	if len(authToken) < authTokenMinimalLen {
		return fmt.Errorf("agent authentication token length must be greater than %d, curently: %d", authTokenMinimalLen, len(authToken))
	}
	return nil
}
