package apiserver

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/api/security"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/api/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type contextKey int

const (
	contextKeyTokenInfoID contextKey = iota
	contextKeyConn
)

// validateToken - validates token for legacy API
func validateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := util.Validate(w, r); err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// parseToken parses the token and validate it for our gRPC API, it returns an empty
// struct and an error or nil
func (p *module) parseToken(token string) (struct{}, error) {
	if token != p.token {
		return struct{}{}, errors.New("Invalid session token")
	}

	// Currently this empty struct doesn't add any information
	// to the context, but we could potentially add some custom
	// type.
	return struct{}{}, nil
}

//grpcAuth is a middleware (interceptor) that extracts and verifies token from header
func (p *module) grpcAuth(ctx context.Context) (context.Context, error) {

	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, err
	}

	tokenInfo, err := p.parseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	// do we need this at all?
	newCtx := context.WithValue(ctx, contextKeyTokenInfoID, tokenInfo)

	return newCtx, nil
}

func (p *module) buildSelfSignedKeyPair() ([]byte, []byte) {

	hosts := []string{"127.0.0.1", "localhost", "::1"}
	if p.config.IpcAddress != "" {
		hosts = append(hosts, p.config.IpcAddress)
	}
	_, rootCertPEM, rootKey, err := security.GenerateRootCert(hosts, 2048)
	if err != nil {
		return nil, nil
	}

	// PEM encode the private key
	rootKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey),
	})

	// Create and return TLS private cert and key
	return rootCertPEM, rootKeyPEM
}

func (p *module) initializeTLS() {

	cert, key := p.buildSelfSignedKeyPair()
	if cert == nil {
		panic("unable to generate certificate")
	}
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}
	p.tlsKeyPair = &pair
	p.tlsCertPool = x509.NewCertPool()
	ok := p.tlsCertPool.AppendCertsFromPEM(cert)
	if !ok {
		panic("bad certs")
	}

	p.tlsAddr = p.config.IpcAddress
}

// FetchAuthToken gets the authentication token from the auth token file & creates one if it doesn't exist
// Requires that the config has been set up before calling
func (p *module) fetchAuthToken() (string, error) {
	return fetchAuthToken(p, false)
}

// CreateOrFetchToken gets the authentication token from the auth token file & creates one if it doesn't exist
// Requires that the config has been set up before calling
func (p *module) createOrFetchAuthToken() (string, error) {
	return fetchAuthToken(p, true)
}

func fetchAuthToken(p *module, tokenCreationAllowed bool) (string, error) {
	authTokenFile := p.tokenFile

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

// DeleteAuthToken removes auth_token file (test clean up)
func (p *module) deleteAuthToken() error {
	return os.Remove(p.tokenFile)
}
