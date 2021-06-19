package util

import (
	"context"
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
