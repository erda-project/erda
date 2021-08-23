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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/scheduler/endpoints"
	notifierapi "github.com/erda-project/erda/modules/scheduler/events"
	"github.com/erda-project/erda/modules/scheduler/impl/cap"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/impl/instanceinfo"
	"github.com/erda-project/erda/modules/scheduler/impl/job"
	"github.com/erda-project/erda/modules/scheduler/impl/labelmanager"
	"github.com/erda-project/erda/modules/scheduler/impl/resourceinfo"
	"github.com/erda-project/erda/modules/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/modules/scheduler/impl/volume"
	"github.com/erda-project/erda/modules/scheduler/impl/volume/driver"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/strutil"
)

type Responser = httpserver.Responser

type HTTPResponse = httpserver.HTTPResponse

type Server struct {
	router        *mux.Router
	listenAddr    string
	sched         *task.Sched
	store         jsonstore.JsonStore
	runtimeMap    map[string]*apistructs.ServiceGroup
	l             sync.Mutex
	notifier      notifierapi.Notifier
	httpendpoints *endpoints.HTTPEndpoints
}

// NewServer creates a server instance, it be used to handle http request and offer
// a api gateway role.
func NewServer(addr string) *Server {
	option := jsonstore.UseCacheEtcdStore(context.Background(), "/dice", 500)
	store, err := jsonstore.New(option)
	if err != nil {
		panic(err)
	}
	sched, err := task.NewSched()
	if err != nil {
		panic(err)
	}
	notifier, err := notifierapi.New("dice-scheduler", nil)
	if err != nil {
		panic(err)
	}

	volumeStore, err := jsonstore.New(jsonstore.UseMemEtcdStore(context.Background(), volume.ETCDVolumeMetadataDir, nil, nil))
	if err != nil {
		panic(err)
	}
	drivers := map[apistructs.VolumeType]volume.Volume{
		apistructs.LocalVolume: driver.NewLocalVolumeDriver(volumeStore),
		apistructs.NasVolume:   driver.NewNasVolumeDriver(volumeStore),
	}
	js, err := jsonstore.New()
	volumeImpl := volume.NewVolumeImpl(drivers)
	//metricImpl := metric.New()
	clusterImpl := cluster.NewClusterImpl(store)
	clusterinfoImpl := clusterinfo.NewClusterInfoImpl(js)
	servicegroupImpl := servicegroup.NewServiceGroupImpl(store, sched, clusterinfoImpl)
	jobImpl := job.NewJobImpl(store, sched)
	labelManagerImpl := labelmanager.NewLabelManager()
	instanceinfoImpl := instanceinfo.NewInstanceInfoImpl()
	componentImpl := instanceinfo.NewComponentInfoImpl()
	resourceinfoImpl := resourceinfo.NewResourceInfoImpl()
	capImpl := cap.NewCapImpl()

	if err != nil {
		panic(err)
	}

	httpendpoints := endpoints.NewHTTPEndpoints(volumeImpl, servicegroupImpl, clusterImpl, jobImpl, labelManagerImpl, instanceinfoImpl, clusterinfoImpl, componentImpl, resourceinfoImpl, capImpl)

	server := &Server{
		router:        mux.NewRouter(),
		listenAddr:    addr,
		sched:         sched,
		store:         store,
		runtimeMap:    make(map[string]*apistructs.ServiceGroup),
		l:             sync.Mutex{},
		notifier:      notifier,
		httpendpoints: httpendpoints,
	}

	go server.clearMap()

	// Register cluster event hook
	if err = registerClusterHook(); err != nil {
		panic(err)
	}

	return server
}

func (s *Server) clearMap() {
	logrus.Infof("enter clearMap")
	for {
		select {
		case <-time.After(30 * time.Second):
			ks := make([]string, 0)
			for k := range s.runtimeMap {
				//delete(s.runtimeMap, k)
				ks = append(ks, k)
			}

			for _, k := range ks {
				s.l.Lock()
				delete(s.runtimeMap, k)
				s.l.Unlock()
			}
			//s.runtimeMap = make(map[string]*spec.ServiceGroup)
		}
	}
}

// ListendAndServe boot the server to lisen and accept requests.
func (s *Server) ListenAndServe() error {
	s.initEndpoints()
	go getDCOSTokenAuthPeriodically()

	srv := &http.Server{
		Addr:              s.listenAddr,
		Handler:           s.router,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		logrus.Errorf("failed to listen and serve: %s", err)
	}
	return err
}

// initEndpoints adds the all routes of endpoint.
func (s *Server) initEndpoints() {
	var endpoints = []struct {
		Path    string
		Method  string
		Handler func(context.Context, *http.Request, map[string]string) (Responser, error)
	}{

		// Deprecated, use for forward compatibility
		// system endpoints
		{"/info", http.MethodGet, s.epInfo},

		// Deprecated, use for forward compatibility
		// job endpoints
		{"/v1/job/create", http.MethodPut, s.httpendpoints.JobCreate},
		{"/v1/job/{namespace}/{name}/start", http.MethodPost, s.httpendpoints.JobStart},
		{"/v1/job/{namespace}/{name}/stop", http.MethodPost, s.httpendpoints.JobStop},
		{"/v1/job/{namespace}/{name}/delete", http.MethodDelete, s.httpendpoints.JobDelete},
		{"/v1/jobs", http.MethodDelete, s.httpendpoints.DeleteJobs},
		{"/v1/job/{namespace}/{name}", http.MethodGet, s.httpendpoints.JobInspect},
		{"/v1/jobs/{namespace}", http.MethodGet, s.httpendpoints.JobList},
		{"/api/jobvolume", http.MethodPost, s.httpendpoints.JobVolumeCreate},

		// Deprecated, use for forward compatibility
		// service endpoints
		{"/v1/runtime/create", http.MethodPost, s.epCreateRuntime},
		{"/v1/runtime/{namespace}/{name}/update", http.MethodPut, s.epUpdateRuntime},
		{"/v1/runtime/{namespace}/{name}/service/{serviceName}/update", http.MethodPut, s.epUpdateService},
		{"/v1/runtime/{namespace}/{name}/restart", http.MethodPost, s.epRestartRuntime},
		{"/v1/runtime/{namespace}/{name}/cancel", http.MethodPost, s.epCancelAction},
		{"/v1/runtime/{namespace}/{name}/delete", http.MethodDelete, s.epDeleteRuntime},
		{"/v1/runtime/{namespace}/{name}", http.MethodGet, s.epGetRuntime},
		{"/v1/runtime/{namespace}/{name}/nocache", http.MethodGet, s.epGetRuntimeNoCache},

		// Deprecated, use for forward compatibility
		// runtimeinfo endpoints for display
		{"/v1/runtimeinfo/{prefix}", http.MethodGet, s.epPrefixGetRuntime},
		{"/v1/runtimeinfo/{namespace}/{prefix}", http.MethodGet, s.epPrefixGetRuntimeWithNamespace},
		{"/v1/runtimeinfo/status/{namespace}/{name}", http.MethodGet, s.epGetRuntimeStatus},

		// Deprecated
		// for notifying endpoints
		//{"/v1/notify/job/{namespace}/{name}", http.MethodGet, s.epGetJobStatusForNotify},

		// platform endpoints
		{"/platform", http.MethodPost, s.epCreatePlatform},
		{"/platform/{name}", http.MethodDelete, s.epDeletePlatform},
		{"/platform/{name}", http.MethodGet, s.epGetPlatform},
		{"/platform/{name}", http.MethodPut, s.epUpdatePlatform},
		{"/platform/{name}/actions/restart", http.MethodPost, s.epRestartPlatform},

		// volume endpoints
		{"/api/volumes", http.MethodPost, s.httpendpoints.VolumeCreate},
		{"/api/volumes/{id}", http.MethodDelete, s.httpendpoints.VolumeDelete},
		{"/api/volumes/{id}", http.MethodGet, s.httpendpoints.VolumeInfo},

		{"/api/servicegroup", http.MethodPost, s.httpendpoints.ServiceGroupCreate},
		{"/api/servicegroup", http.MethodPut, s.httpendpoints.ServiceGroupUpdate},
		{"/api/servicegroup", http.MethodDelete, s.httpendpoints.ServiceGroupDelete},
		{"/api/servicegroup", http.MethodGet, i18nPrinter(s.httpendpoints.ServiceGroupInfo)},
		{"/api/servicegroup/actions/restart", http.MethodPost, s.httpendpoints.ServiceGroupRestart},
		{"/api/servicegroup/actions/cancel", http.MethodPost, s.httpendpoints.ServiceGroupCancel},
		{"/api/servicegroup/actions/precheck", http.MethodPost, s.httpendpoints.ServiceGroupPrecheck},
		{"/api/servicegroup/actions/config", http.MethodPut, s.httpendpoints.ServiceGroupConfigUpdate},
		{"/api/servicegroup/actions/killpod", http.MethodPost, s.httpendpoints.ServiceGroupKillPod},
		{"/api/servicegroup/actions/scale", http.MethodPut, s.httpendpoints.ServiceScaling},

		// creating cluster by hooking colony-soldier's event
		{"/clusterhook", http.MethodPost, s.httpendpoints.ClusterHook},
		// DEPRECATED
		{"/clusters", http.MethodPost, s.httpendpoints.ClusterCreate},

		{"/api/nodelabels", http.MethodGet, s.httpendpoints.LabelList},
		{"/api/nodelabels", http.MethodPost, s.httpendpoints.SetNodeLabels},

		{"/api/podinfo", http.MethodGet, s.httpendpoints.PodInfo},
		{"/api/instanceinfo", http.MethodGet, s.httpendpoints.InstanceInfo},
		{"/api/serviceinfo", http.MethodGet, s.httpendpoints.ServiceInfo},
		{"/api/componentinfo", http.MethodGet, s.httpendpoints.ComponentInfo},
		// {"/api/servicegroupinfo", http.MethodGet, s.httpendpoints.ServiceGroupStatusInfo},

		{"/api/clusterinfo/{clusterName}", http.MethodGet, s.httpendpoints.ClusterInfo},
		{"/api/resourceinfo/{clusterName}", http.MethodGet, s.httpendpoints.ResourceInfo},
		{"/api/capacity", http.MethodGet, s.httpendpoints.CapacityInfo},
	}

	for _, ep := range endpoints {
		s.router.Path(ep.Path).Methods(ep.Method).HandlerFunc(s.internal(ep.Handler))
		logrus.Infof("added endpoint: %s %s", ep.Method, ep.Path)
	}

	s.router.Path("/api/terminal").HandlerFunc(s.httpendpoints.Terminal)

	subroute := s.router.PathPrefix("/debug/").Subrouter()
	subroute.HandleFunc("/pprof/profile", pprof.Profile)
	subroute.HandleFunc("/pprof/trace", pprof.Trace)
	subroute.HandleFunc("/pprof/block", pprof.Handler("block").ServeHTTP)
	subroute.HandleFunc("/pprof/heap", pprof.Handler("heap").ServeHTTP)
	subroute.HandleFunc("/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	subroute.HandleFunc("/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}

func (s *Server) internal(handler func(context.Context, *http.Request, map[string]string) (Responser, error)) http.HandlerFunc {
	pctx := context.Background()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(pctx)
		defer cancel()
		defer r.Body.Close()

		//start := time.Now()
		//logrus.Infof("start %s %s", r.Method, r.URL.String())
		//defer func() {
		//	logrus.Infof("end %s %s (took %v)", r.Method, r.URL.String(), time.Since(start))
		//}()

		response, err := handler(ctx, r, mux.Vars(r))
		if err != nil {
			logrus.Errorf("failed to handle request: %s (%v)", r.URL.String(), err)

			if response != nil {
				w.WriteHeader(response.GetStatus())
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			io.WriteString(w, err.Error())
			return
		}
		if response == nil || response.GetContent() == nil && response.GetStatus() != http.StatusNotFound {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.GetStatus())

		if err := json.NewEncoder(w).Encode(response.GetContent()); err != nil {
			logrus.Errorf("failed to send response: %s (%v)", r.URL.String(), err)
			return
		}
	}
}

// TODO
func (s *Server) epInfo(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	return HTTPResponse{
		Status:  http.StatusOK,
		Content: version.String(),
	}, nil
}

type endpoint func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error)

func i18nPrinter(f endpoint) endpoint {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
		lang := r.Header.Get("Lang")
		p := message.NewPrinter(language.English)
		if strutil.Contains(lang, "zh") {
			p = message.NewPrinter(language.SimplifiedChinese)
		}
		ctx2 := context.WithValue(ctx, "i18nPrinter", p)
		return f(ctx2, r, vars)
	}
}

// registerClusterHook register cluster webhook in eventBox
func registerClusterHook() error {
	bdl := bundle.New(bundle.WithEventBox())

	ev := apistructs.CreateHookRequest{
		Name:   "scheduler-clusterhook",
		Events: []string{"cluster"},
		URL:    fmt.Sprintf("http://%s/clusterhook", discover.Scheduler()),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}

	if err := bdl.CreateWebhook(ev); err != nil {
		logrus.Warnf("failed to register cluster event, (%v)", err)
		return err
	}

	logrus.Infof("register cluster event success")
	return nil
}
