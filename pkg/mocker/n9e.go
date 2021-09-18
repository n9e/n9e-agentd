package mocker

import (
	"encoding/json"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/n9e/agent-payload/gogen"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/yubo/apiserver/pkg/rest"
	"k8s.io/klog/v2"
)

func (p *mocker) routesV1N9e() []rest.WsRoute {
	return []rest.WsRoute{{
		Method:  "POST",
		SubPath: "/series",
		Handle:  p.n9eSeries,
	}, {
		Method:  "GET",
		SubPath: "/collect-rules-belong-to-ident",
		Handle:  p.getCollectRules,
		Output:  api.CollectRulesWrap{},
	}, {
		Method:  "GET",
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
		RespWrite:          n9eRespWrite,
	})

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/v1/n9e",
		GoRestfulContainer: http,
		Tags:               []string{"n9e"},
		Routes:             p.routesV1N9e(),
		RespWrite:          n9eRespWrite,
	})

}

func (p *mocker) n9eSeries(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *gogen.N9EMetricsPayload) (string, error) {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "n9e.series")
	defer sp.Finish()
	sp.LogFields(log.Object("series", data))

	buf, _ := json.Marshal(data)
	klog.InfoS("recv n9e series", "len(samples)", len(data.Samples), "buf", string(buf))
	return "", nil
}

func (p *mocker) getCollectRules(w http.ResponseWriter, req *http.Request) ([]api.CollectRule, error) {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "n9e.get_collect_rules")
	defer sp.Finish()

	klog.Infof("%+v", p.rules)

	return p.rules.GetRules(), nil
}

func (p *mocker) getCollectRulesSummary(w http.ResponseWriter, req *http.Request) (*api.CollectRulesSummary, error) {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "n9e.get_collect_rules_summary")
	defer sp.Finish()

	klog.Infof("%+v", p.rules)

	return p.rules.GetSummary(), nil
}

// output: {"dat":"", "err":""}
func n9eRespWrite(resp *restful.Response, req *http.Request, data interface{}, err error) {
	v := map[string]interface{}{"dat": data}

	if err != nil {
		v["err"] = err.Error()
	}

	resp.WriteEntity(v)
}
