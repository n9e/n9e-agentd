package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/klog/v2"
)

var (
	AuthTokenFile string // inited by configer
	AuthToken     string // inited/updated by fetchAuthToken()
)

const (
	authTokenMinimalLen = 32
)

// GetClusterAgentAuthToken gets the session token
func GetClusterAgentAuthToken() (string, error) {
	return AuthToken, nil
}

// GetAuthToken gets the session token
func GetAuthToken() string {
	return AuthToken
}

// GetAuthTokenFilepath returns the path to the auth_token file.
func GetAuthTokenFilepath() string {
	return AuthTokenFile
}

// FetchAuthToken gets the authentication token from the auth token file & creates one if it doesn't exist
// Requires that the config has been set up before calling
func FetchAuthToken() (string, error) {
	return fetchAuthToken(false)
}

// CreateOrFetchToken gets the authentication token from the auth token file & creates one if it doesn't exist
// Requires that the config has been set up before calling
func CreateOrFetchToken() (string, error) {
	return fetchAuthToken(true)
}

// DeleteAuthToken removes auth_token file (test clean up)
func DeleteAuthToken() error {
	return os.Remove(AuthTokenFile)
}

func fetchAuthToken(tokenCreationAllowed bool) (string, error) {
	authTokenFile := AuthTokenFile

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

	AuthToken = authToken

	return authToken, nil
}

func validateAuthToken(authToken string) error {
	if len(authToken) < authTokenMinimalLen {
		return fmt.Errorf("cluster agent authentication token length must be greater than %d, curently: %d", authTokenMinimalLen, len(authToken))
	}
	return nil
}
