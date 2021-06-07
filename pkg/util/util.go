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
