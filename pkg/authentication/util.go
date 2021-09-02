package authentication

import (
	"os"

	"github.com/n9e/n9e-agentd/pkg/util"
)

const (
	authTokenMinimalLen = 32
)

// GetClusterAgentAuthToken gets the session token
//func GetClusterAgentAuthToken() (string, error) {
//	return _auth.Token, nil
//}

// GetAuthToken gets the session token
func GetAuthToken() string {
	return _auth.Token
}

// GetAuthTokenFilepath returns the path to the auth_token file.
func GetAuthTokenFilepath() string {
	return _auth.AuthTokenFile
}

// FetchAuthToken gets the authentication token from the auth token file & creates one if it doesn't exist
// Requires that the config has been set up before calling
func FetchAuthToken() (string, error) {
	return _auth.fetchAuthToken(false)
}

// CreateOrFetchToken gets the authentication token from the auth token file & creates one if it doesn't exist
// Requires that the config has been set up before calling
func CreateOrFetchToken() (string, error) {
	return _auth.fetchAuthToken(true)
}

// DeleteAuthToken removes auth_token file (test clean up)
func DeleteAuthToken() error {
	return os.Remove(_auth.AuthTokenFile)
}

func (p *Config) fetchAuthToken(tokenCreationAllowed bool) (string, error) {
	if p.Token != "" {
		return p.Token, nil
	}
	return util.FetchAuthToken(p.AuthTokenFile, tokenCreationAllowed)
}
