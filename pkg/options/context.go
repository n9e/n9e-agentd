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
	secureKey
	httpServerKey
	genericServerKey
	respKey
	authenticationKey
	authorizationKey
	//tracerKey
)

// WithValue returns a copy of parent in which the value associated with key is val.
func WithValue(parent context.Context, key interface{}, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}

func WithCore(parent context.Context, core Core) context.Context {
	return WithValue(parent, coreKey, core)
}

func CoreFrom(ctx context.Context) (Core, bool) {
	core, ok := ctx.Value(coreKey).(Core)
	return core, ok
}

func CoreMustFrom(ctx context.Context) Core {
	core, ok := ctx.Value(coreKey).(Core)
	if !ok {
		panic("unable get core module")
	}
	return core
}

// WithAuthn returns a copy of parent in which the authenticationInfo value is set
func WithAuthn(parent context.Context, authn Authn) context.Context {
	return WithValue(parent, authenticationKey, authn)
}

// AuthnFrom returns the value of the authenticationInfo key on the ctx
func AuthnFrom(ctx context.Context) (Authn, bool) {
	authn, ok := ctx.Value(authenticationKey).(Authn)
	return authn, ok
}

// WithAuthz returns a copy of parent in which the authorizationInfo value is set
func WithAuthz(parent context.Context, authz Authz) context.Context {
	return WithValue(parent, authorizationKey, authz)
}

// AuthzFrom returns the value of the authorizationInfo key on the ctx
func AuthzFrom(ctx context.Context) (Authz, bool) {
	authz, ok := ctx.Value(authorizationKey).(Authz)
	return authz, ok
}
