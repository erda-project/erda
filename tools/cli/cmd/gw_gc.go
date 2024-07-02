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

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
)

var GwGc = command.Command{
	ParentName: "Gw",
	Name:       "gc",
	ShortHelp:  "Erda Gateway delete invalid endpoints",
	LongHelp:   "Erda Gateway delete invalid endpoints",
	Example:    "erda-cli gw gc -i erda.cloud.invalid-endpoints.json --kong-admin https://your-admin-host --cluster your-cluster",
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "i",
			Name:         "input",
			Doc:          "[Required] the input file which is generated from 'erda-cli gw ls'",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "o",
			Name:         "output",
			Doc:          "[Optional] the output file for the result",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "C",
			Name:         "cluster",
			Doc:          "[Required] cluster specified",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "",
			Name:         "kong-admin",
			Doc:          "[Required] replace the kong-admin host from the input file",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "y",
			Name:         "yes",
			Doc:          "[Optional] to be sure to delete",
			DefaultValue: false,
		},
		command.BoolFlag{
			Short:        "",
			Name:         "delete-from-erda",
			Doc:          "[Optional] delete endpoints from erda if specified this, or else only delete from Kong",
			DefaultValue: false,
		},
	},
	Run: RunGwGc,
}

type InvalidEndpoints struct {
	Success bool                 `json:"success"`
	Data    InvalidEndpointsData `json:"data"`
}

type InvalidEndpointsData struct {
	Total                   int                    `json:"total"`
	TotalProjectIsInvalid   int                    `json:"totalProjectIsInvalid"`
	TotalRuntimeIsInvalid   int                    `json:"totalRuntimeIsInvalid"`
	TotalInnerAddrIsInvalid int                    `json:"totalInnerAddrIsInvalid"`
	List                    []InvalidEndpointsItem `json:"list"`
}

type InvalidEndpointsItem struct {
	InvalidReason   string `json:"invalidReason"`
	Type            string `json:"type"`
	ProjectID       string `json:"projectID"`
	PackageID       string `json:"packageID"`
	PackageApiID    string `json:"packageApiID"`
	RuntimeID       string `json:"runtimeID"`
	InnerHostname   string `json:"innerHostname"`
	KongRouteID     string `json:"kongRouteID"`
	KongServiceID   string `json:"kongServiceID"`
	ClusterName     string `json:"clusterName"`
	RouteDeleting   string `json:"routeDeleting"`
	ServiceDeleting string `json:"serviceDeleting"`
}

func RunGwGc(context *command.Context, input, output, cluster, kongAdmin string, yes, deleteFromErda bool) error {
	var ctx = *context

	if err := GwDelPreCheck(input, &output, cluster, kongAdmin, deleteFromErda); err != nil {
		return err
	}

	out, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	// read from file
	in, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	var endpoints InvalidEndpoints
	if err = json.Unmarshal(in, &endpoints); err != nil {
		return err
	}
	if !endpoints.Success {
		return errors.Errorf("invalid endpoints file: %s", "not success")
	}
	if endpoints.Data.Total == 0 || len(endpoints.Data.List) == 0 {
		ctx.Warn("exit: %s", "endpoints total is 0")
		return nil
	}
	for _, item := range endpoints.Data.List {
		if item.ClusterName != cluster {
			return errors.Errorf("the item's cluster %s dose not equals with %s that you specified. api: %s/%s", item.ClusterName, cluster, item.PackageID, item.PackageApiID)
		}
	}

	// do deleting
	for _, item := range endpoints.Data.List {
		if err := deleteItem(ctx, out, item, kongAdmin, !yes); err != nil {
			return err
		}
	}

	return nil
}

func GwDelPreCheck(input string, output *string, cluster, kongAdmin string, deleteFromErda bool) error {
	// check --input
	if input == "" {
		return errors.New("the endpoints input file must be specified")
	}
	// check --output
	if *output == "" {
		*output = "result." + input
	}
	// check the cluster
	if cluster == "" {
		return errors.New("--cluster must be specified")
	}
	// check --kong-admin
	kongAdmin = strings.TrimSuffix(kongAdmin, "/")
	if _, err := url.Parse(kongAdmin); err != nil {
		return err
	}
	// check --delete-from
	if deleteFromErda {
		return errors.New("the flag --delete-from-erda is not supported yet")
	}
	return nil
}

func deleteItem(ctx command.Context, w io.Writer, item InvalidEndpointsItem, kongAdmin string, needSure bool) error {
	ctx.Info("%s/%s ", item.PackageID, item.PackageApiID)
	_, _ = fmt.Fprintf(w, "%s/%s: ", item.PackageID, item.PackageApiID)
	defer fmt.Fprintf(w, "\n")

	client := http.DefaultClient
	client.Timeout = time.Second * 100

	const (
		Route   = "Route"
		Service = "Service"
	)
	var deleting = func(typ, id string) error {
		if err := sure(ctx, needSure, typ, id); err != nil {
			return err
		}
		var uri string
		switch typ {
		case Route:
			uri = kongAdmin + "/routes/" + id
		case Service:
			uri = kongAdmin + "/services/" + id
		default:
			return errors.New("invalid Kong Object")
		}
		request, err := http.NewRequest(http.MethodDelete, uri, nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, "Delete Kong %s %v ", typ, response.StatusCode)
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return nil
		}
		ctx.Warn("failed to delete")
		if data, err := io.ReadAll(response.Body); err == nil {
			defer response.Body.Close()
			_, _ = fmt.Fprint(w, string(data))
		}
		return nil
	}

	if item.KongRouteID == "" {
		ctx.Warn("No KongRouteID in the Item. %s/%s", item.PackageID, item.PackageApiID)
		_, _ = fmt.Fprintf(w, "No Kong Route. ")
	} else if err := deleting(Route, item.KongRouteID); err != nil {
		return err
	}
	if item.KongServiceID == "" {
		ctx.Warn("No KongServiceID in the Item. %s/%s", item.PackageID, item.PackageApiID)
		_, _ = fmt.Fprintf(w, "No Kong Service. ")
	} else if err := deleting(Service, item.KongServiceID); err != nil {
		return err
	}

	return nil
}

func sure(ctx command.Context, needSure bool, tye, id string) error {
	if !needSure {
		return nil
	}
	ctx.Info("please press 'y' to be sure you are going to delete the %s %s", tye, id)
	var sure string
	if _, err := fmt.Scanln(&sure); err != nil {
		return err
	}
	if strings.EqualFold(sure, "y") ||
		strings.EqualFold(sure, "yes") ||
		strings.EqualFold(sure, "y\n") ||
		strings.EqualFold(sure, "yes\n") {
		return nil
	}
	return errors.New("not sure")
}
