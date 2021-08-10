package util

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

func AbsPath(path, rootPath, defPath string) string {
	if path == "" {
		path = defPath
	}

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(rootPath, path)
}
