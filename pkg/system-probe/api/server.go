package api

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-agent/pkg/process/net"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	gorilla "github.com/gorilla/mux"
	"github.com/n9e/n9e-agentd/pkg/system-probe/api/module"
	"github.com/n9e/n9e-agentd/pkg/system-probe/config"
	"github.com/n9e/n9e-agentd/pkg/system-probe/modules"
	"github.com/n9e/n9e-agentd/pkg/system-probe/utils"
)

// StartServer starts the HTTP server for the system-probe, which registers endpoints from all enabled modules.
func StartServer(cfg *config.Config) error {
	conn, err := net.NewListener(cfg.SysprobeSocket)
	if err != nil {
		return fmt.Errorf("error creating IPC socket: %s", err)
	}

	mux := gorilla.NewRouter()
	err = module.Register(cfg, mux, modules.All)
	if err != nil {
		return fmt.Errorf("failed to create system probe: %s", err)
	}

	// Register stats endpoint
	mux.HandleFunc("/debug/stats", func(w http.ResponseWriter, req *http.Request) {
		stats := module.GetStats()
		utils.WriteAsJSON(w, stats)
	})

	// TODO
	//setupConfigHandlers(mux)

	// Module-restart handler
	mux.HandleFunc("/module-restart/{module-name}", restartModuleHandler).Methods("POST")

	mux.Handle("/debug/vars", http.DefaultServeMux)

	go func() {
		err = http.Serve(conn.GetListener(), mux)
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("error creating HTTP server: %s", err)
		}
	}()

	return nil
}

func init() {
	expvar.Publish("modules", expvar.Func(func() interface{} {
		return module.GetStats()
	}))
}
