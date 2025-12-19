package policy_group

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
)

// RoutingModelInstance represents a routable model endpoint with labels.
type RoutingModelInstance struct {
	ModelWithProvider *cachehelpers.ModelWithProvider
	Labels            map[string]string
}

// RequestMeta carries extracted request keys (e.g. sticky key).
type RequestMeta struct {
	Keys map[string]string
}

func (m RequestMeta) Get(inputKey string) (string, bool) {
	if inputKey == "" {
		return "", false
	}
	for k, v := range m.Keys {
		if strings.EqualFold(k, inputKey) {
			return v, true
		}
	}
	return "", false
}

// RouteRequest is the input for routing to a model instance.
type RouteRequest struct {
	ClientID  string
	Group     *pb.PolicyGroup
	Instances []*RoutingModelInstance // Instances is the pool of instances under the client.
	Meta      RequestMeta
}

func (r RouteRequest) GetStickyKey() string {
	return r.Group.StickyKey
}

func (r RouteRequest) GetStickyValue() string {
	if r.GetStickyKey() == "" {
		return ""
	}
	stickyKey := r.GetStickyKey()
	stickyValue, ok := r.Meta.Get(stickyKey)
	if !ok || stickyValue == "" {
		const headerPrefix = "req.header."
		if strings.HasPrefix(strings.ToLower(stickyKey), headerPrefix) {
			stickyValue, _ = r.Meta.Get(stickyKey[len(headerPrefix):])
		}
	}
	return stickyValue
}

func (r RouteRequest) getRoutingKeyByNamespace(namespace string) string {
	stickyValue := r.GetStickyValue()
	if stickyValue == "" {
		return ""
	}
	return fmt.Sprintf("client:%s|group:%s|namespace:%s|sticky_value:%s", r.ClientID, r.Group.Name, namespace, stickyValue)
}

func (r RouteRequest) GetRoutingKeyForBranch(branchName string) string {
	return r.getRoutingKeyByNamespace("branch:" + branchName)
}

type RouteTraceGroup struct {
	Source string `json:"source"`
	Name   string `json:"name"`
	Mode   string `json:"mode"`
	Desc   string `json:"desc"`
}

type RouteTraceSticky struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Success bool   `json:"success"`
	// FallbackFromSticky indicates sticky routing was requested but didn't match,
	// and the final decision was made by normal routing.
	FallbackFromSticky bool `json:"fallbackFromSticky"`
}

type RouteTraceBranch struct {
	Name     string `json:"name"`
	Weight   uint64 `json:"weight"`
	Priority uint64 `json:"priority"`
	Strategy string `json:"strategy"`
}

type RouteTraceInstance struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RouteTrace contains the decision trail for observability.
type RouteTrace struct {
	Group  RouteTraceGroup   `json:"group"`
	Sticky *RouteTraceSticky `json:"sticky"`
	Branch RouteTraceBranch  `json:"branch"`
}
