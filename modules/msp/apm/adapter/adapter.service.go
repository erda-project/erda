// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package adapter

import (
	context "context"
	"encoding/json"

	pb "github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type adapterService struct {
	p *provider
}

var (
	JAVAAdapterStrategies = []*pb.AdapterStrategy{
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_JAVA_AGENT), Strategy: pb.AdapterStrategyKey_JAVA_AGENT.String(), Enable: true},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_APACHE_SKYWALKING), Strategy: pb.AdapterStrategyKey_APACHE_SKYWALKING.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_JAEGER), Strategy: pb.AdapterStrategyKey_JAEGER.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_OPEN_TELEMETRY), Strategy: pb.AdapterStrategyKey_OPEN_TELEMETRY.String(), Enable: false},
	}

	GOAdapterStrategies = []*pb.AdapterStrategy{
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_APACHE_SKYWALKING), Strategy: pb.AdapterStrategyKey_APACHE_SKYWALKING.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_JAEGER), Strategy: pb.AdapterStrategyKey_JAEGER.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_OPEN_TELEMETRY), Strategy: pb.AdapterStrategyKey_OPEN_TELEMETRY.String(), Enable: false},
	}

	PHPAdapterStrategies = []*pb.AdapterStrategy{
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_APACHE_SKYWALKING), Strategy: pb.AdapterStrategyKey_APACHE_SKYWALKING.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_JAEGER), Strategy: pb.AdapterStrategyKey_JAEGER.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_OPEN_TELEMETRY), Strategy: pb.AdapterStrategyKey_OPEN_TELEMETRY.String(), Enable: false},
	}

	DOTNETAdapterStrategies = []*pb.AdapterStrategy{
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_APACHE_SKYWALKING), Strategy: pb.AdapterStrategyKey_APACHE_SKYWALKING.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_JAEGER), Strategy: pb.AdapterStrategyKey_JAEGER.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_OPEN_TELEMETRY), Strategy: pb.AdapterStrategyKey_OPEN_TELEMETRY.String(), Enable: false},
	}

	NODEJSAdapterStrategies = []*pb.AdapterStrategy{
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_NODEJS_AGENT), Strategy: pb.AdapterStrategyKey_NODEJS_AGENT.String(), Enable: true},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_APACHE_SKYWALKING), Strategy: pb.AdapterStrategyKey_APACHE_SKYWALKING.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_JAEGER), Strategy: pb.AdapterStrategyKey_JAEGER.String(), Enable: false},
		{DisplayName: getStrategyDisplayName(pb.AdapterStrategyKey_OPEN_TELEMETRY), Strategy: pb.AdapterStrategyKey_OPEN_TELEMETRY.String(), Enable: false},
	}
)

var (
	JAVAAdapterList   = pb.Adapters{Language: pb.AdapterLanguage_JAVA.String(), DisplayName: getLanguageDisplayName(pb.AdapterLanguage_JAVA), Strategies: JAVAAdapterStrategies}
	GOAdapterList     = pb.Adapters{Language: pb.AdapterLanguage_GO.String(), DisplayName: getLanguageDisplayName(pb.AdapterLanguage_GO), Strategies: GOAdapterStrategies}
	PHPAdapterList    = pb.Adapters{Language: pb.AdapterLanguage_PHP.String(), DisplayName: getLanguageDisplayName(pb.AdapterLanguage_PHP), Strategies: PHPAdapterStrategies}
	DOTNETAdapterList = pb.Adapters{Language: pb.AdapterLanguage_DOT_NET.String(), DisplayName: getLanguageDisplayName(pb.AdapterLanguage_DOT_NET), Strategies: DOTNETAdapterStrategies}
	NODEJSAdapterList = pb.Adapters{Language: pb.AdapterLanguage_NODEJS.String(), DisplayName: getLanguageDisplayName(pb.AdapterLanguage_NODEJS), Strategies: NODEJSAdapterStrategies}
)

var Adapters = []*pb.Adapters{
	&JAVAAdapterList,
	&GOAdapterList,
	&PHPAdapterList,
	&DOTNETAdapterList,
	&NODEJSAdapterList,
}

func getLanguageDisplayName(language pb.AdapterLanguage) string {
	switch language {
	case pb.AdapterLanguage_JAVA:
		return "Java"
	case pb.AdapterLanguage_GO:
		return "Golang"
	case pb.AdapterLanguage_PHP:
		return "PHP"
	case pb.AdapterLanguage_DOT_NET:
		return ".Net Core"
	case pb.AdapterLanguage_NODEJS:
		return "Node.js"
	default:
		return ""
	}
}

func getStrategyDisplayName(strategyKey pb.AdapterStrategyKey) string {
	switch strategyKey {
	case pb.AdapterStrategyKey_JAVA_AGENT:
		return "Java Agent"
	case pb.AdapterStrategyKey_APACHE_SKYWALKING:
		return "Apache Skywalking"
	case pb.AdapterStrategyKey_JAEGER:
		return "Jaeger"
	case pb.AdapterStrategyKey_OPEN_TELEMETRY:
		return "Open Telemetry"
	case pb.AdapterStrategyKey_NODEJS_AGENT:
		return "Node.js Agent"
	default:
		return ""
	}
}

func (s *adapterService) GetAdapters(ctx context.Context, req *pb.GetAdaptersRequest) (*pb.GetAdaptersResponse, error) {
	var adapterList []*pb.Adapters
	buffer, _ := json.Marshal(&Adapters)
	err := json.Unmarshal(buffer, &adapterList)
	if err != nil {
		return nil, err
	}
	return &pb.GetAdaptersResponse{Data: adapterList}, nil
}

func (s *adapterService) GetAdapterDocs(ctx context.Context, request *pb.GetAdapterDocsRequest) (*pb.GetAdapterDocsResponse, error) {
	return nil, errors.NewUnimplementedError("GetAdapterDocs")
}
