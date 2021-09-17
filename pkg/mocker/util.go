package mocker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/yubo/apiserver/pkg/responsewriters"
	"k8s.io/klog/v2"
)

func payloads(name string) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, req *http.Request) error {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}
		klog.InfoS(name, "payload.len", len(b))
		return nil
	}
}

func writeRawJSON(object interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

// output: {"dat":"", "err":""}
func N9eRespWrite(resp *restful.Response, req *http.Request, data interface{}, err error) {
	var eMsg string
	status := responsewriters.ErrorToAPIStatus(err)

	if err != nil {
		eMsg = err.Error()

		if klog.V(3).Enabled() {
			klog.ErrorDepth(1, fmt.Sprintf("httpReturn %d %s", status.Code, eMsg))
		}
	}

	resp.WriteEntity(map[string]interface{}{
		"dat": data,
		"err": eMsg,
	})
}
