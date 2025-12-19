package handler_policy_group

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type Handler struct {
	DAO dao.DAO
}

func (h *Handler) Create(ctx context.Context, req *pb.PolicyGroupCreateRequest) (*pb.PolicyGroup, error) {
	if req.Source == "" {
		req.Source = common_types.PolicyGroupSourceUserDefined.String()
	}
	resp, err := h.DAO.PolicyGroupClient().Create(ctx, req)
	if err == nil {
		ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypePolicyGroup)
	}
	return resp, err
}

func (h *Handler) Get(ctx context.Context, req *pb.PolicyGroupGetRequest) (*pb.PolicyGroup, error) {
	return h.DAO.PolicyGroupClient().Get(ctx, req)
}

func (h *Handler) Update(ctx context.Context, req *pb.PolicyGroupUpdateRequest) (*pb.PolicyGroup, error) {
	// check exist
	exist, err := h.Get(ctx, &pb.PolicyGroupGetRequest{Id: req.Id, ClientId: req.ClientId})
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, fmt.Errorf("failed to update policy group: %s not found", req.Id)
	}
	// do update
	resp, err := h.DAO.PolicyGroupClient().Update(ctx, req)
	if err == nil {
		ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypePolicyGroup)
	}
	return resp, err
}

func (h *Handler) Delete(ctx context.Context, req *pb.PolicyGroupDeleteRequest) (*commonpb.VoidResponse, error) {
	resp, err := h.DAO.PolicyGroupClient().Delete(ctx, req)
	if err == nil {
		ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypePolicyGroup)
	}
	return resp, err
}

func (h *Handler) Paging(ctx context.Context, req *pb.PolicyGroupPagingRequest) (*pb.PolicyGroupPagingResponse, error) {
	return h.DAO.PolicyGroupClient().Paging(ctx, req)
}
