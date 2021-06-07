package options

import (
	"io"
	//"github.com/yubo/apiserver/pkg/authentication/authenticator"
	//"github.com/yubo/apiserver/pkg/authorization/authorizer"
	//genericoptions "github.com/yubo/apiserver/modules/apiserver/options"
	//"github.com/yubo/apiserver/modules/apiserver/server"
)

type Client interface {
	GetId() string
	GetSecret() string
	GetRedirectUri() string
}

type HttpServer interface {
	// http
	//Handle(pattern string, handler http.Handler)
	//HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))

	//// restful.Container
	//Add(service *restful.WebService) *restful.Container
	//Filter(filter restful.FilterFunction)
}

type GenericServer interface {
	//Handle(string, http.Handler)
	//HandleFunc(string, func(http.ResponseWriter, *http.Request))
	//UnlistedHandle(string, http.Handler)
	//UnlistedHandleFunc(string, func(http.ResponseWriter, *http.Request))
	//Add(*restful.WebService) *restful.Container
	//Filter(restful.FilterFunction)
	//Server() *server.GenericAPIServer
}

type Executer interface {
	Execute(wr io.Writer, data interface{}) error
}

type Core interface {
	Hostname() string
	ConfigFile() string
}

type Authn interface {
	//APIAudiences() authenticator.Audiences
	//Authenticator() authenticator.Request
}

type Authz interface {
	//Authorizer() authorizer.Authorizer
	//RuleResolver() authorizer.RuleResolver
}
