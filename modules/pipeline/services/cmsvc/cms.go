package cmsvc

import (
	"context"
	"errors"
	"sort"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/cms"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

var (
	errEmptyNs = errors.New("ns is empty")
)

func (s *CMSvc) CreateNS(source apistructs.PipelineSource, ns string) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, cms.CtxKeyPipelineSource, source)

	// 参数校验
	if ns == "" {
		return apierrors.ErrCreatePipelineCmsNs.InvalidParameter(errEmptyNs)
	}

	err := s.cm.IdempotentCreateNS(ctx, ns)
	if err != nil {
		return apierrors.ErrCreatePipelineCmsNs.InternalError(err)
	}
	return nil
}

func (s *CMSvc) DeleteNS(source apistructs.PipelineSource, ns string) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, cms.CtxKeyPipelineSource, source)

	// 参数校验
	if ns == "" {
		return apierrors.ErrDeletePipelineCmsConfigs.InvalidParameter(errEmptyNs)
	}

	err := s.cm.IdempotentDeleteNS(ctx, ns)
	if err != nil {
		return apierrors.ErrDeletePipelineCmsNs.InternalError(err)
	}
	return nil
}

func (s *CMSvc) PrefixListNS(source apistructs.PipelineSource, nsPrefix string) ([]apistructs.PipelineCmsNs, error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, cms.CtxKeyPipelineSource, source)

	namespaces, err := s.cm.PrefixListNS(ctx, nsPrefix)
	if err != nil {
		return nil, apierrors.ErrListPipelineCmsNs
	}
	return namespaces, nil
}

func (s *CMSvc) UpdateConfigs(source apistructs.PipelineSource, ns string, kvs map[string]apistructs.PipelineCmsConfigValue) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, cms.CtxKeyPipelineSource, source)

	// 参数校验
	if ns == "" {
		return apierrors.ErrUpdatePipelineCmsConfigs.InvalidParameter(errEmptyNs)
	}

	err := s.cm.UpdateConfigs(ctx, ns, kvs)
	if err != nil {
		return apierrors.ErrUpdatePipelineCmsConfigs.InternalError(err)
	}
	return nil
}

func (s *CMSvc) DeleteConfigs(source apistructs.PipelineSource, ns string, deleteKeys []string, forceDel bool) error {

	ctx := context.Background()
	ctx = context.WithValue(ctx, cms.CtxKeyPipelineSource, source)
	ctx = context.WithValue(ctx, cms.CtxKeyForceDelete, forceDel)

	// 参数校验
	if ns == "" {
		return apierrors.ErrDeletePipelineCmsConfigs.InvalidParameter(errEmptyNs)
	}

	err := s.cm.DeleteConfigs(ctx, ns, deleteKeys...)
	if err != nil {
		return apierrors.ErrDeletePipelineCmsConfigs.InternalError(err)
	}
	return nil
}

func (s *CMSvc) GetConfigs(source apistructs.PipelineSource, ns string, globalDecrypt bool, keys ...apistructs.PipelineCmsConfigKey) (
	[]apistructs.PipelineCmsConfig, error) {

	ctx := context.Background()
	ctx = context.WithValue(ctx, cms.CtxKeyPipelineSource, source)

	// 参数校验
	if ns == "" {
		return nil, apierrors.ErrGetPipelineCmsConfigs.InvalidParameter(errEmptyNs)
	}

	configs, err := s.cm.GetConfigs(ctx, ns, globalDecrypt, keys...)
	if err != nil {
		return nil, apierrors.ErrGetPipelineCmsConfigs.InternalError(err)
	}

	results := make([]apistructs.PipelineCmsConfig, 0, len(configs))
	for key, value := range configs {
		results = append(results, apistructs.PipelineCmsConfig{
			Key:                    key,
			PipelineCmsConfigValue: value,
		})
	}

	// order by createdTime
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].TimeCreated == nil {
			return true
		}
		return results[i].TimeCreated.Before(*results[j].TimeCreated)
	})

	return results, nil
}
