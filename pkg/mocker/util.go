package mocker

import (
	"encoding/json"
	"net/http"
)

func non(w http.ResponseWriter, req *http.Request) {}

func writeRawJSON(object interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
