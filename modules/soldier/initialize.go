// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package soldier

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pkg/colonyutil"
	"github.com/erda-project/erda/modules/soldier/autoop"
	"github.com/erda-project/erda/modules/soldier/command"
	"github.com/erda-project/erda/modules/soldier/conf"
	"github.com/erda-project/erda/modules/soldier/health"
	"github.com/erda-project/erda/modules/soldier/mysql"
	"github.com/erda-project/erda/modules/soldier/proxy"
	"github.com/erda-project/erda/modules/soldier/registry"
	"github.com/erda-project/erda/modules/soldier/settings"
)

// Initialize Application-related initialization operations
func Initialize() error {
	defer settings.Wait()
	conf.Load()

	go autoop.StartCron()

	router := mux.NewRouter()
	router.Use(colonyutil.LoggingMiddleware)
	router.HandleFunc("/_health_check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	})

	router.HandleFunc(health.ErdaHealthPath, health.GetSoldierHealth)

	registryRouter := router.PathPrefix("/registry").Subrouter()
	registryRouter.Methods("POST").PathPrefix("/remove/manifests").HandlerFunc(registry.RemoveManifests)
	//registryRouter.Methods("POST").PathPrefix("/remove/layers").HandlerFunc(registry.RemoveLayers)
	registryRouter.Methods("GET").PathPrefix("/readonly").HandlerFunc(registry.Readonly)

	autoopRouter := router.PathPrefix("/autoop").Subrouter()
	autoopRouter.Methods("POST").PathPrefix("/run/{name}").HandlerFunc(autoop.RunAction)
	autoopRouter.Methods("POST").PathPrefix("/cancel/{name}").HandlerFunc(autoop.CancelAction)
	autoopRouter.Methods("POST").PathPrefix("/cron").HandlerFunc(autoop.CronActions)

	mysqlRouter := router.PathPrefix("/mysql").Subrouter()
	mysqlRouter.Methods("POST").Path("/init").HandlerFunc(mysql.Init)
	mysqlRouter.Methods("POST").Path("/check").HandlerFunc(mysql.Check)
	mysqlRouter.Methods("POST").Path("/exec").HandlerFunc(mysql.Exec)
	mysqlRouter.Methods("POST").Path("/exec_file").HandlerFunc(mysql.ExecFile)

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.PathPrefix("/proxy/{service}").HandlerFunc(proxy.ProxyService)
	apiRouter.HandleFunc("/command", command.Command)
	apiRouter.HandleFunc("/terminal", command.Terminal)
	//apiRouter.Methods("POST").Path("/cluster/download").HandlerFunc(command.DiceDownload)
	//apiRouter.Methods("GET").Path("/cluster/download").HandlerFunc(command.ReadDownloadResult)
	apiRouter.PathPrefix("/nodes").HandlerFunc(proxy.ProxyFPS)

	//router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)
	//router.PathPrefix("/debug/pprof/cmdline").HandlerFunc(pprof.Cmdline)
	//router.PathPrefix("/debug/pprof/profile").HandlerFunc(pprof.Profile)
	//router.PathPrefix("/debug/pprof/symbol").HandlerFunc(pprof.Symbol)
	//router.PathPrefix("/debug/pprof/trace").HandlerFunc(pprof.Trace)

	go func() {
		logrus.Infoln("http serve", settings.HTTPAddr)
		err := http.ListenAndServe(settings.HTTPAddr, router)
		if err != nil {
			logrus.Fatalln(err)
		}
	}()
	return nil
}
