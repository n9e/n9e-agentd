package mocker

import (
	"encoding/json"
	"net/http"

	"github.com/n9e/agent-payload/gogen"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/yubo/apiserver/pkg/rest"
	"k8s.io/klog/v2"
)

func (p *mocker) routesV1N9e() []rest.WsRoute {
	return []rest.WsRoute{{
		Method:  "POST",
		SubPath: "/series",
		Handle:  p.n9eSeries,
	}, {
		Method:  "POST",
		SubPath: "/collect-rules-belong-to-ident",
		Handle:  p.getCollectRules,
		Output:  api.CollectRulesWrap{},
	}, {
		Method:  "POST",
		SubPath: "/collect-rules-summary",
		Handle:  p.getCollectRulesSummary,
		Output:  api.CollectRulesSummaryWrap{},
	}}

}

func (p *mocker) installN9eWs(http rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("n9e", "N9E API - n9e api")

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/apis/n9e.didiyun.com/v1",
		GoRestfulContainer: http,
		Tags:               []string{"api groups"},
		Routes:             p.routesV1N9e(),
		RespWrite:          N9eRespWrite,
	})

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/v1/n9e",
		GoRestfulContainer: http,
		Tags:               []string{"n9e"},
		Routes:             p.routesV1N9e(),
		RespWrite:          N9eRespWrite,
	})

}

func (p *mocker) n9eSeries(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *gogen.N9EMetricsPayload) (string, error) {
	buf, _ := json.Marshal(data)
	klog.InfoS("recv n9e series", "len(samples)", len(data.Samples), "buf", string(buf))
	return "", nil
}

func (p *mocker) getCollectRules(w http.ResponseWriter, _ *http.Request) ([]api.CollectRule, error) {
	return p.rules.GetRules(), nil
}

func (p *mocker) getCollectRulesSummary(w http.ResponseWriter, _ *http.Request) (*api.CollectRulesSummary, error) {
	return p.rules.GetSummary(), nil
}
