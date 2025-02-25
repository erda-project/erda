package runtime

import (
	"encoding/json"
	"github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/apistructs"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ConvertCreateRuntimePbToDTO(req *pb.RuntimeCreateRequest) *apistructs.RuntimeCreateRequest {
	return &apistructs.RuntimeCreateRequest{
		Name:        req.Name,
		ReleaseID:   req.ReleaseId,
		Operator:    req.Operator,
		ClusterName: req.ClusterName,
		Source:      apistructs.RuntimeSource(req.Source),
		Extra: apistructs.RuntimeCreateRequestExtra{
			OrgID:           req.Extra.OrgId,
			ProjectID:       req.Extra.ProjectId,
			ApplicationID:   req.Extra.ApplicationId,
			ApplicationName: req.Extra.ApplicationName,
			Workspace:       req.Extra.Workspace,
			BuildID:         req.Extra.BuildId,
			DeployType:      req.Extra.DeployType,
			InstanceID:      json.Number(req.Extra.InstanceId),
			ClusterId:       json.Number(req.Extra.ClusterId),
			AddonActions:    ConvertMapAnyStringToInterface(req.Extra.AddonActions),
		},
		SkipPushByOrch:    req.SkipPushByOrch,
		Param:             req.Param,
		DeploymentOrderId: req.DeploymentOrderId,
		ReleaseVersion:    req.ReleaseVersion,
		ExtraParams:       req.ExtraParams,
	}
}

func ConvertDeploymentCreateResponseDTOToPb(req *apistructs.DeploymentCreateResponseDTO) *pb.DeploymentCreateResponse {
	return &pb.DeploymentCreateResponse{
		DeploymentId:  req.DeploymentID,
		ApplicationId: req.ApplicationID,
		RuntimeId:     req.RuntimeID,
	}
}

func ConvertRuntimeReleaseCreateRequestToDTO(req *pb.RuntimeReleaseCreateRequest) *apistructs.RuntimeReleaseCreateRequest {
	return &apistructs.RuntimeReleaseCreateRequest{
		ReleaseID:     req.ReleaseId,
		Workspace:     req.Workspace,
		ProjectID:     req.ProjectId,
		ApplicationID: req.ApplicationId,
	}
}

func ConvertErrResponseToPb(respErr []apistructs.ErrorResponse) []*pb.ErrorResponse {
	errs := make([]*pb.ErrorResponse, 0, len(respErr))
	for _, e := range respErr {
		errs = append(errs, &pb.ErrorResponse{
			Code: e.Code,
			Msg:  e.Msg,
			Ctx:  ConvertInterfaceToAny(e.Ctx),
		})
	}
	return errs
}

func ConvertErrResponseToDTO(respErr []*pb.ErrorResponse) []apistructs.ErrorResponse {
	errs := make([]apistructs.ErrorResponse, 0, len(respErr))
	for _, e := range respErr {
		errs = append(errs, apistructs.ErrorResponse{
			Code: e.Code,
			Msg:  e.Msg,
			Ctx:  ConvertAnyToInterface(e.Ctx),
		})
	}
	return errs
}

func ConvertRuntimeServiceResourceDTOToPb(req *apistructs.RuntimeServiceResourceDTO) *pb.RuntimeServiceResource {
	return &pb.RuntimeServiceResource{
		Cpu:  req.CPU,
		Mem:  int64(req.Mem),
		Disk: int64(req.Disk),
	}
}

func ConvertRuntimeServiceDeploymentsDTOPb(req *apistructs.RuntimeServiceDeploymentsDTO) *pb.RuntimeServiceDeployments {
	return &pb.RuntimeServiceDeployments{Replicas: int64(req.Replicas)}
}

func ConvertRuntimeInspectServiceToDTO(req *pb.RuntimeInspectService) *apistructs.RuntimeInspectServiceDTO {
	return &apistructs.RuntimeInspectServiceDTO{
		Status:      req.Status,
		HPAEnabled:  req.HpaEnabled,
		VPAEnabled:  req.VpaEnabled,
		Type:        req.Type,
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{},
		Resources:   apistructs.RuntimeServiceResourceDTO{},
		Envs:        req.Envs,
		Addrs:       req.Addrs,
		Expose:      req.Expose,
		Errors:      ConvertErrResponseToDTO(req.Errors),
	}
}

func ConvertRuntimeInspectServiceDTOPb(req *apistructs.RuntimeInspectServiceDTO) *pb.RuntimeInspectService {
	return &pb.RuntimeInspectService{
		Status:      req.Status,
		HpaEnabled:  req.HPAEnabled,
		VpaEnabled:  req.VPAEnabled,
		Type:        req.Type,
		Deployments: ConvertRuntimeServiceDeploymentsDTOPb(&req.Deployments),
		Resources:   ConvertRuntimeServiceResourceDTOToPb(&req.Resources),
		Envs:        req.Envs,
		Addrs:       req.Addrs,
		Expose:      req.Expose,
		Errors:      ConvertErrResponseToPb(req.Errors),
	}
}

func MapRuntimeInspectServiceDTOToPb(req map[string]*apistructs.RuntimeInspectServiceDTO) map[string]*pb.RuntimeInspectService {
	data := map[string]*pb.RuntimeInspectService{}
	for k, v := range req {
		data[k] = ConvertRuntimeInspectServiceDTOPb(v)
	}
	return data
}

func ConvertMapStringMapToPb(req map[string]map[string]string) map[string]*pb.StringMap {
	data := map[string]*pb.StringMap{}
	for k, v := range req {
		data[k] = &pb.StringMap{Data: v}
	}
	return data
}

func ConvertRuntimeSummaryDTOToPb(req *apistructs.RuntimeSummaryDTO) *pb.RuntimeSummary {
	return &pb.RuntimeSummary{
		Id:                    req.ID,
		Name:                  req.Name,
		ServiceGroupName:      req.ServiceGroupName,
		ServiceGroupNamespace: req.ServiceGroupNamespace,
		Source:                string(req.Source),
		Status:                req.Status,
		DeployStatus:          string(req.DeployStatus),
		DeleteStatus:          req.DeleteStatus,
		ReleaseId:             req.ReleaseID,
		ClusterId:             req.ClusterID,
		ClusterName:           req.ClusterName,
		ClusterType:           req.ClusterType,
		Resources:             ConvertRuntimeServiceResourceDTOToPb(&req.Resources),
		Extra:                 ConvertMapInterfaceToAny(req.Extra),
		ProjectId:             req.ProjectID,
		Services:              MapRuntimeInspectServiceDTOToPb(req.Services),
		ModuleErrMsg:          ConvertMapStringMapToPb(req.ModuleErrMsg),
		TimeCreated:           timestamppb.New(req.TimeCreated),
		CreatedAt:             timestamppb.New(req.CreatedAt),
		UpdatedAt:             timestamppb.New(req.UpdatedAt),
		DeployAt:              timestamppb.New(req.DeployAt),
		Errors:                ConvertErrResponseToPb(req.Errors),
		Creator:               req.Creator,
		ApplicationId:         req.ApplicationID,
		ApplicationName:       req.ApplicationName,
		DeploymentOrderId:     req.DeploymentOrderId,
		DeploymentOrderName:   req.DeploymentOrderName,
		ReleaseVersion:        req.ReleaseVersion,
		RawStatus:             req.RawStatus,
		RawDeploymentStatus:   req.RawDeploymentStatus,
		LastOperator:          req.LastOperator,
		LastOperatorName:      req.LastOperatorName,
		LastOperatorAvatar:    req.LastOperatorAvatar,
		LastOperatorTime:      timestamppb.New(req.LastOperateTime),
		LastOperatorId:        req.LastOperatorId,
	}
}

func ConvertRuntimeSummaryToList(req []apistructs.RuntimeSummaryDTO) []*pb.RuntimeSummary {
	data := make([]*pb.RuntimeSummary, 0, len(req))
	for _, v := range req {
		summary := ConvertRuntimeSummaryDTOToPb(&v)
		data = append(data, summary)
	}
	return data
}

func ConvertEpBulkGetRuntimeStatusDetailToPb(req map[uint64]interface{}) *pb.EpBulkGetRuntimeStatusDetailResponse {
	data := map[uint64]*anypb.Any{}
	for k, v := range req {
		any := ConvertInterfaceToAny(v)
		data[k] = any
	}
	return &pb.EpBulkGetRuntimeStatusDetailResponse{
		Data: data,
	}

}

func ConvertAnyToInterface(req *anypb.Any) interface{} {
	return req
}

func ConvertInterfaceToAny(req interface{}) *anypb.Any {
	return req.(*anypb.Any)
}

func ConvertMapAnyStringToInterface(req map[string]*anypb.Any) map[string]interface{} {
	data := map[string]interface{}{}
	for k, v := range req {
		data[k] = ConvertInterfaceToAny(v)
	}
	return data
}

func ConvertMapInterfaceToAny(req map[string]interface{}) map[string]*anypb.Any {
	data := map[string]*anypb.Any{}
	for k, v := range req {
		data[k] = ConvertInterfaceToAny(v)
	}
	return data
}

func ConvertRuntimeInspectServiceMapToPb(req map[string]*pb.RuntimeInspectService) map[string]*apistructs.RuntimeInspectServiceDTO {
	data := map[string]*apistructs.RuntimeInspectServiceDTO{}
	for k, v := range req {
		data[k] = ConvertRuntimeInspectServiceToDTO(v)
	}
	return data
}

func ConvertPreDiceToDTO(req *pb.PreDiceDTO) apistructs.PreDiceDTO {
	return apistructs.PreDiceDTO{
		Name:     req.Name,
		Envs:     req.Envs,
		Services: ConvertRuntimeInspectServiceMapToPb(req.Services),
	}
}

func ConvertRuntimeScaleRecordToDTO(req *pb.RuntimeScaleRecord) apistructs.RuntimeScaleRecord {
	return apistructs.RuntimeScaleRecord{
		ApplicationId: req.ApplicationId,
		Workspace:     req.Workspace,
		Name:          req.Name,
		RuntimeID:     req.RuntimeId,
		PayLoad:       ConvertPreDiceToDTO(req.Payload),
		ErrMsg:        req.ErrMsg,
	}
}

func ConvertRuntimeScaleRecordsToDTO(req *pb.RuntimeScaleRecords) *apistructs.RuntimeScaleRecords {
	runtimes := make([]apistructs.RuntimeScaleRecord, len(req.Runtimes))
	for _, v := range req.Runtimes {
		runtimes = append(runtimes, ConvertRuntimeScaleRecordToDTO(v))
	}
	return &apistructs.RuntimeScaleRecords{
		Runtimes: runtimes,
		IDs:      req.Ids,
	}
}

func ConvertReferClusterResponseToPb(req bool) *pb.ReferClusterResponse {
	return &pb.ReferClusterResponse{
		Data: req,
	}
}

func ConvertRuntimeLogsRequestToDTO(req *pb.RuntimeLogsRequest) *apistructs.DashboardSpotLogRequest {
	return &apistructs.DashboardSpotLogRequest{
		ID:          req.Id,
		Source:      apistructs.DashboardSpotLogSource(req.Source),
		Stream:      apistructs.DashboardSpotLogStream(req.Stream),
		Count:       req.Count,
		Start:       time.Duration(req.Start),
		End:         time.Duration(req.End),
		Debug:       req.Debug,
		ClusterName: req.ClusterName,
		PipelineID:  req.PipelineID,
	}
}

func ConvertDashboardSpotLogLineToPb(req *apistructs.DashboardSpotLogLine) *pb.DashboardSpotLogLine {
	return &pb.DashboardSpotLogLine{
		Id:        req.ID,
		Source:    req.Source,
		Stream:    req.Stream,
		Timestamp: req.TimeStamp,
		Content:   req.Content,
		Offset:    req.Offset,
		Level:     req.Level,
		RequestId: req.RequestID,
	}
}

func ConvertDashboardSpotLogLineListToPb(req []apistructs.DashboardSpotLogLine) []*pb.DashboardSpotLogLine {
	data := make([]*pb.DashboardSpotLogLine, 0, len(req))
	for _, v := range req {
		data = append(data, ConvertDashboardSpotLogLineToPb(&v))
	}
	return data
}

func ConvertDashboardSpotLogDataToPb(req *apistructs.DashboardSpotLogData) *pb.DashboardSpotLogData {
	return &pb.DashboardSpotLogData{
		Lines:      ConvertDashboardSpotLogLineListToPb(req.Lines),
		IsFallback: req.IsFallBack,
	}
}

func ConvertCountPRByWorkspaceResponse(req map[string]uint64) *pb.CountPRByWorkspaceResponse {
	return &pb.CountPRByWorkspaceResponse{Data: req}
}

func ConvertBatchRuntimeServiceResponse(req map[uint64]*apistructs.RuntimeSummaryDTO) *pb.BatchRuntimeServiceResponse {
	var data *pb.BatchRuntimeServiceResponse
	for key, value := range req {
		data.Data[key] = ConvertRuntimeSummaryDTOToPb(value)
	}
	return data
}
