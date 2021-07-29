package options

import (
	"context"
	// authentication "github.com/yubo/apiserver/modules/authentication/lib"
)

// The key type is unexported to prevent collisions
type key int

const (
	_ key = iota
	coreKey
	apiserverKey
)

// WithValue returns a copy of parent in which the value associated with key is val.
func WithValue(parent context.Context, key interface{}, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}
