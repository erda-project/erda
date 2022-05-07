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

package gallery

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	_ "github.com/erda-project/erda-infra/providers/mysql/v2"
	"github.com/erda-project/erda-proto-go/admin/gallery/pb"
	commonPb "github.com/erda-project/erda-proto-go/common/pb"
	extensionPb "github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	releasepb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/admin/services/cache"
	"github.com/erda-project/erda/modules/admin/services/gallery/apierr"
	"github.com/erda-project/erda/modules/admin/services/gallery/model"
	"github.com/erda-project/erda/pkg/common/apis"
)

var (
	name = "erda.admin.gallery"
	spec = servicehub.Spec{
		Define:               nil,
		Services:             pb.ServiceNames(),
		Dependencies:         nil,
		OptionalDependencies: []string{"service-register"},
		DependenciesFunc:     nil,
		Summary:              "gallery service",
		Description:          "gallery service",
		ConfigFunc: func() interface{} {
			return new(struct{})
		},
		Types: pb.Types(),
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

func init() {
	servicehub.Register(name, &spec)
}

// +provider
type provider struct {
	R transport.Register `autowired:"service-register" required:"true"`

	// providers clients
	C *cache.Cache `autowired:"easy-memory-cache-client"`
	D *gorm.DB     `autowired:"mysql-gorm.v2-client"`

	// gRPC clients
	extensionCli extensionPb.ExtensionServiceServer `autowired:"erda.core.dicehub.extension.ExtensionService"`
	releaseCli   releasepb.ReleaseServiceServer     `autowired:"erda.core.dicehub.release.ReleaseGetDiceService"`

	l *logrus.Entry
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.l = logrus.WithField("provider", name)
	p.l.Infoln("Init")
	if p.R == nil {
		p.l.Infoln("register self")
		pb.RegisterGalleryImp(p.R, p, apis.Options())
	}
	return nil
}

func (p *provider) ListOpusTypes(ctx context.Context, _ *commonPb.VoidRequest) (*pb.ListOpusTypesRespData, error) {
	// todo: i18n
	var data pb.ListOpusTypesRespData
	for type_, name := range apistructs.OpusTypes {
		data.List = append(data.List, &pb.OpusType{Type: type_.String(), Name: name})
	}
	data.Total = uint32(len(data.List))
	return &data, nil
}

func (p *provider) ListOpus(ctx context.Context, req *pb.ListOpusReq) (*pb.ListOpusResp, error) {
	var l = p.l.WithField("func", "ListOpus")

	orgID := apis.GetOrgID(ctx)
	if orgID == "" {
		return nil, apierr.ListOpus.InvalidParameter("invalid orgID")
	}
	var pageNo, pageSize = 1, 10
	if req.GetPageNo() >= 1 {
		pageNo = int(req.GetPageNo())
	}
	if req.GetPageSize() >= 10 && req.GetPageSize() <= 1000 {
		pageSize = int(req.GetPageSize())
	}
	where := p.D.Where("org_id = ? OR level = ?", orgID, apistructs.OpusLevelSystem)
	if req.GetType() != "" {
		where = where.Where("type = ?", req.GetType())
	}
	if req.GetName() != "" {
		where = where.Where("name = ?", req.GetName())
	}

	// 查出满足 where 条件的 opuses
	var (
		opuses    []*model.Opus
		opusesIDs []string
	)
	if err := where.
		Find(&opuses).
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find opuses")
		return nil, apierr.ListOpus.InternalError(err)
	}
	if len(opuses) == 0 {
		l.Warnln("not found")
		return new(pb.ListOpusResp), nil
	}
	for _, opus := range opuses {
		opusesIDs = append(opusesIDs, opus.ID.String)
	}

	if req.GetKeyword() == "" {
		return p.listOpusByIDs(ctx, pageSize, pageNo, opusesIDs)
	}
	return p.listOpusWithKeyword(ctx, pageSize, pageNo, req.GetKeyword(), opusesIDs)
}

func (p *provider) ListOpusVersions(ctx context.Context, req *pb.ListOpusVersionsReq) (*pb.ListOpusVersionsResp, error) {
	var l = p.l.WithField("func", "listOpusVersions").WithField("opusID", req.GetOpusID())

	orgID := apis.GetOrgID(ctx)
	if orgID == "" {
		return nil, apierr.ListOpus.InvalidParameter("invalid orgID")
	}
	lang := apis.GetLang(ctx)
	lang = strings.ToLower(lang)
	lang = strings.ReplaceAll(lang, "-", "_")

	// todo: 鉴权

	var (
		opus          model.Opus
		versions      []*model.OpusVersion
		presentations []*model.OpusPresentation
		readmes       []*model.OpusReadme
		installations []*model.OpusInstallation
	)
	if err := p.D.Where("id = ?", req.GetOpusID()).First(&opus).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierr.ListOpusVersions.NotFound()
		}
		l.WithError(err).Errorln("failed to First opus")
		return nil, apierr.ListOpusVersions.InternalError(err)
	}
	if err := p.D.Where("opus_id = ?", req.GetOpusID()).Find(&versions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.WithError(err).Errorln("opus's version not found")
			return nil, apierr.ListOpusVersions.NotFound()
		}
		l.WithError(err).Errorln("failed to Find versions")
		return nil, apierr.ListOpusVersions.InternalError(err)
	}
	if err := p.D.Where("opus_id = ?", req.GetOpusID()).Find(&presentations).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find presentations")
		return nil, apierr.ListOpusVersions.InternalError(err)
	}
	if err := p.D.Where("opus_id = ?", req.GetOpusID()).Find(&readmes).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find readmes")
		return nil, apierr.ListOpusVersions.InternalError(err)
	}
	// todo: not used
	if err := p.D.Where("opus_id = ?", req.GetOpusID()).Find(&installations).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find installation")
		return nil, apierr.ListOpusVersions.InternalError(err)
	}

	var (
		presentationMap = make(map[string]*model.OpusPresentation)
		readmesMap      = make(map[string]map[string]*model.OpusReadme)
		installationMap = make(map[string]*model.OpusInstallation)
	)
	for _, item := range presentations {
		presentationMap[item.VersionID] = item
	}
	for _, item := range readmes {
		m, ok := readmesMap[item.VersionID]
		if !ok {
			m = make(map[string]*model.OpusReadme)
		}
		m[item.Lang] = item
		readmesMap[item.VersionID] = m
	}
	for _, item := range installations {
		installationMap[item.VersionID] = item
	}

	var resp = pb.ListOpusVersionsResp{Data: &pb.ListOpusVersionsRespData{
		Id:               opus.ID.String,
		CreatedAt:        timestamppb.New(opus.CreatedAt),
		UpdatedAt:        timestamppb.New(opus.UpdatedAt),
		OrgID:            orgID,
		OrgName:          opus.OrgName,
		CreatorID:        opus.CreatorID,
		UpdaterID:        opus.UpdaterID,
		Level:            opus.Level,
		Type:             opus.Type,
		Name:             apistructs.OpusTypes[apistructs.OpusType(opus.Type)],
		DisplayName:      opus.DisplayName,
		Catalog:          opus.Catalog,
		DefaultVersionID: opus.DefaultVersionID,
		LatestVersionID:  opus.LatestVersionID,
	}}
	for _, version := range versions {
		item := pb.ListOpusVersionRespDataVersion{
			Id:        version.ID.String,
			CreatedAt: timestamppb.New(version.CreatedAt),
			UpdatedAt: timestamppb.New(version.UpdatedAt),
			CreatorID: version.CreatorID,
			UpdaterID: version.UpdaterID,
			Version:   version.Version,
			Summary:   version.Summary,
			Labels:    nil,
			LogoURL:   version.LogoURL,
			IsValid:   version.IsValid,
		}
		if err := json.Unmarshal([]byte(version.Labels), &item.Labels); err != nil {
			l.WithError(err).Errorf("failed to Unmarshal version.Labels, labels: %s", version.Labels)
		}
		if pre, ok := presentationMap[version.ID.String]; ok {
			item.Ref = pre.Ref
			item.Desc = pre.Desc
			item.ContactName = pre.ContactName
			item.ContactURL = pre.ContactURL
			item.ContactEmail = pre.ContactEmail
			item.IsOpenSourced = pre.IsOpenSourced
			item.OpensourceURL = pre.OpensourceURL
			item.LicenceName = pre.LicenceName
			item.LicenceURL = pre.LicenceURL
			item.HomepageName = pre.HomepageName
			item.HomepageURL = pre.HomepageURL
			item.HomepageLogoURL = pre.HomepageLogoURL
			item.IsDownloadable = pre.IsDownloadable
			item.DownloadURL = pre.DownloadURL
		}
		if langs, ok := readmesMap[version.ID.String]; ok {
			for k := range langs {
				item.ReadmeLang = langs[k].Lang
				item.ReadmeLangName = langs[k].LangName
				item.Readme = langs[k].Text
				if k == lang {
					break
				}
			}
		}

		resp.Data.Versions = append(resp.Data.Versions, &item)
		resp.UserIDs = append(resp.UserIDs, version.CreatorID)
	}

	return &resp, nil
}

func (p *provider) PutOnArtifacts(ctx context.Context, req *pb.PutOnArtifactsReq) (*commonPb.VoidResponse, error) {
	l := p.l.WithField("func", "PutOnArtifacts").
		WithField("name", req.GetName()).
		WithField("version", req.GetVersion())

	// check parameters and get org info and user info
	var orgID = apis.GetOrgID(ctx)
	if orgID == "" {
		orgID = req.GetOrgID()
	}
	if orgID == "" {
		return nil, apierr.PutOnArtifacts.InvalidParameter("missing orgID")
	}
	orgIDInt, err := strconv.ParseUint(orgID, 10, 32)
	if err != nil {
		return nil, apierr.PutOnArtifacts.InvalidParameter("invalid orgID: " + orgID)
	}
	var userID = apis.GetUserID(ctx)
	if userID == "" {
		userID = req.GetUserID()
	}
	if userID == "" {
		return nil, apierr.PutOnArtifacts.InvalidParameter("missing userID")
	}
	orgDto, ok := p.C.GetOrgByOrgID(orgID)
	if !ok {
		return nil, apierr.PutOnArtifacts.InvalidParameter("invalid orgID: " + orgID)
	}
	if req.GetName() == "" {
		return nil, apierr.PutOnArtifacts.InvalidParameter("missing Opus name")
	}
	if req.GetVersion() == "" {
		return nil, apierr.PutOnArtifacts.InvalidParameter("missing Opus version")
	}

	// todo: 鉴权

	// Check if artifacts already exist
	var (
		opus     model.Opus
		versions []*model.OpusVersion
		common   = model.Common{
			OrgID:     uint32(orgIDInt),
			OrgName:   orgDto.Name,
			CreatorID: userID,
			UpdaterID: userID,
		}
	)
	tx := p.D.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	switch err = p.D.Where(map[string]interface{}{
		"org_id": orgID,
		"type":   apistructs.OpusTypeArtifactsProject,
		"name":   req.GetName(),
	}).First(&opus).Error; {
	case err == nil:
		err = p.D.Where(map[string]interface{}{
			"opus_id": opus.ID,
			"version": req.GetVersion(),
		}).Find(&versions).Error
		if err == nil {
			l.Warnln("already exists")
			return nil, apierr.PutOnArtifacts.AlreadyExists()
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			l.WithError(err).Errorln("failed to Find versions")
			return nil, apierr.PutOnArtifacts.InternalError(err)
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		opus = model.Opus{
			Common:      common,
			Level:       string(apistructs.OpusLevelOrg),
			Type:        string(apistructs.OpusTypeArtifactsProject),
			Name:        req.GetName(),
			DisplayName: req.GetDisplayName(),
			Catalog:     req.GetCatalog(),
		}
		if err = tx.Create(&opus).Error; err != nil {
			l.WithError(err).Errorln("failed to Create opus")
			return nil, apierr.PutOnArtifacts.InternalError(err)
		}
	default:
		l.WithError(err).Errorln("failed to First opus")
		return nil, apierr.PutOnArtifacts.InternalError(err)
	}

	// create version
	labels, err := json.Marshal(req.GetLabels())
	if err != nil {
		l.WithError(err).Warnf("failed to json.Marshal labels, labels: %v", req.GetLabels())
	}
	var version = model.OpusVersion{
		Model:   model.Model{},
		Common:  common,
		OpusID:  opus.ID.String,
		Version: req.GetVersion(),
		Summary: req.GetSummary(),
		Labels:  string(labels),
		LogoURL: req.GetLogoURL(),
		IsValid: true,
	}
	if err := tx.Create(&version).Error; err != nil {
		l.WithError(err).Errorln("failed to Create version")
		return nil, apierr.PutOnArtifacts.InternalError(err)
	}

	// create presentation
	var presentation = model.OpusPresentation{
		Model:           model.Model{},
		Common:          common,
		OpusID:          opus.ID.String,
		VersionID:       version.ID.String,
		Desc:            req.GetDesc(),
		ContactName:     req.GetContactName(),
		ContactURL:      req.GetContactURL(),
		ContactEmail:    req.GetContactEmail(),
		IsOpenSourced:   req.GetIsOpenSourced(),
		OpensourceURL:   req.GetOpensourceURL(),
		LicenceName:     req.GetLicenseName(),
		LicenceURL:      req.GetLicenseURL(),
		HomepageName:    req.GetHomepageName(),
		HomepageURL:     req.GetHomepageURL(),
		HomepageLogoURL: req.GetHomepageLogoURL(),
		IsDownloadable:  req.GetIsDownloadable(),
		DownloadURL:     req.GetDownloadURL(),
	}
	if err := tx.Create(&presentation).Error; err != nil {
		l.WithError(err).Errorln("failed to Create presentation")
		return nil, apierr.PutOnArtifacts.InternalError(err)
	}

	// create readmes
	var readmes []*model.OpusReadme
	for _, item := range req.GetReadme() {
		readme := &model.OpusReadme{
			Model:     model.Model{},
			Common:    common,
			OpusID:    opus.ID.String,
			VersionID: version.ID.String,
			Lang:      item.Lang,
			LangName:  item.LangName,
			Text:      item.Text,
		}
		readmes = append(readmes, readme)
	}
	if err = tx.CreateInBatches(readmes, len(readmes)).Error; err != nil {
		l.WithError(err).Errorln("failed to CreateInBatches readmes")
		return nil, apierr.PutOnArtifacts.InternalError(err)
	}

	// create installation
	spec, err := json.Marshal(req.GetInstallation())
	if err != nil {
		return nil, apierr.PutOnArtifacts.InternalError(err)
	}
	var installation = model.OpusInstallation{
		Model:     model.Model{},
		Common:    common,
		OpusID:    opus.ID.String,
		VersionID: version.ID.String,
		Installer: string(apistructs.OpusTypeArtifactsProject),
		Spec:      string(spec),
	}
	if err = tx.Create(&installation).Error; err != nil {
		l.WithError(err).Errorln("failed to Create installation")
		return nil, apierr.PutOnArtifacts.InternalError(err)
	}

	// update opus
	if err = tx.Model(&opus).Where(map[string]interface{}{"id": opus.ID}).
		Updates(map[string]interface{}{
			"updater_id":         userID,
			"default_version_id": version.ID,
			"latest_version_id":  version.ID,
		}).Error; err != nil {
		l.WithError(err).Errorln("failed to Updates opus")
		return nil, apierr.PutOnArtifacts.InternalError(err)
	}

	return new(commonPb.VoidResponse), nil
}

func (p *provider) PutOffArtifacts(ctx context.Context, req *pb.PutOffArtifactsReq) (*commonPb.VoidResponse, error) {
	l := p.l.WithField("func", "PutOffArtifacts").
		WithField("opus_id", req.GetOpusID()).
		WithField("version_id", req.GetVersionID())

	// check parameters and get org info and user info
	var orgID = apis.GetOrgID(ctx)
	if orgID == "" {
		orgID = req.GetOrgID()
	}
	if orgID == "" {
		return nil, apierr.PutOffArtifacts.InvalidParameter("missing orgID")
	}
	var userID = apis.GetUserID(ctx)
	if userID == "" {
		userID = req.GetUserID()
	}
	if userID == "" {
		return nil, apierr.PutOffArtifacts.InvalidParameter("missing userID")
	}

	// todo: 鉴权

	// query version
	var version model.OpusVersion
	if err := p.D.Where(map[string]interface{}{
		"id": req.GetVersionID(),
	}).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.WithError(err).Warnln("delete version not found")
			return new(commonPb.VoidResponse), nil
		}
		l.WithError(err).Errorln("failed to First version")
		return nil, apierr.PutOffArtifacts.InternalError(err)
	}

	if req.GetOpusID() != version.OpusID {
		return nil, apierr.PutOffArtifacts.InvalidParameter("invalid opusID and versionID")
	}

	tx := p.D.Begin()
	var err error
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	if err = tx.Delete(model.OpusVersion{}, map[string]interface{}{"id": req.GetVersionID()}).Error; err != nil {
		l.WithError(err).Errorln("failed to Delete version")
		return nil, apierr.PutOffArtifacts.InternalError(err)
	}
	if err = tx.Delete(model.OpusPresentation{}, map[string]interface{}{"version_id": req.GetVersionID()}).Error; err != nil {
		l.WithError(err).Errorln("failed to Delete presentation")
		return nil, apierr.PutOffArtifacts.InternalError(err)
	}
	if err = tx.Delete(model.OpusReadme{}, map[string]interface{}{"version_id": req.GetVersionID()}).Error; err != nil {
		l.WithError(err).Errorln("failed to Delete readme")
		return nil, apierr.PutOffArtifacts.InternalError(err)
	}
	if err = tx.Delete(model.OpusInstallation{}, map[string]interface{}{"version_id": req.GetVersionID()}).Error; err != nil {
		l.WithError(err).Errorln("failed to Delete installation")
		return nil, apierr.PutOffArtifacts.InternalError(err)
	}
	if err2 := p.D.Where(map[string]interface{}{"opus_id": req.GetOpusID()}).Find([]*model.OpusVersion{}).Error; err2 != nil && errors.Is(err2, gorm.ErrRecordNotFound) {
		if err = tx.Delete(model.Opus{}).Where(map[string]interface{}{"id": req.GetOpusID()}).Error; err != nil {
			l.WithError(err).Errorln("failed to Delete opuses")
			return nil, apierr.PutOffArtifacts.InternalError(err)
		}
	}

	return new(commonPb.VoidResponse), nil
}

func (p *provider) PutOnExtensions(ctx context.Context, req *pb.PubOnExtensionsReq) (*commonPb.VoidResponse, error) {
	l := p.l.WithField("func", "PutOnExtensions").
		WithFields(map[string]interface{}{
			"orgID":   req.GetOrgID(),
			"type":    req.GetType(),
			"name":    req.GetName(),
			"version": req.GetVersion(),
			"level":   req.GetLevel(),
			"mode":    req.GetMode(),
		})

	// check parameters and get org info and user info
	var orgID = apis.GetOrgID(ctx)
	if orgID == "" {
		orgID = req.GetOrgID()
	}
	if orgID == "" && apistructs.OpusLevelOrg.Equal(req.GetLevel()) {
		return nil, apierr.PutOnExtension.InvalidState("missing orgID")
	}
	var userID = apis.GetUserID(ctx)
	if userID == "" {
		userID = req.GetUserID()
	}
	if userID == "" && (apistructs.OpusLevelOrg.Equal(req.GetLevel())) {
		return nil, apierr.PutOnExtension.InvalidParameter("missing userID")
	}
	if apistructs.OpusTypeExtensionAddon.Equal(req.GetType()) && apistructs.OpusTypeExtensionAction.Equal(req.GetType()) {
		return nil, apierr.PutOnExtension.InvalidParameter("invalid type: " + req.GetType())
	}
	if req.GetName() == "" {
		return nil, apierr.PutOnExtension.InvalidParameter("missing name")
	}
	if req.GetVersion() == "" {
		return nil, apierr.PutOnExtension.InvalidParameter("missing version")
	}
	if apistructs.OpusLevelSystem.Equal(req.GetLevel()) && apistructs.OpusLevelOrg.Equal(req.GetLevel()) {
		return nil, apierr.PutOnExtension.InvalidParameter("invalid level: " + req.GetLevel())
	}

	// todo: 鉴权

	// Check if extension already exist
	var (
		where = map[string]interface{}{
			"type":  req.GetType(),
			"name":  req.GetName(),
			"level": req.GetLevel(),
		}
		opus    model.Opus
		version model.OpusVersion
	)
	if apistructs.OpusLevelOrg.Equal(req.GetLevel()) {
		where["org_id"] = orgID
	}
	switch err := p.D.Where(where).First(&opus).Error; {
	case err == nil:
		where = map[string]interface{}{
			"opus_id": opus.ID,
			"version": req.GetVersion(),
		}
		switch err := p.D.Where(where).First(&version).Error; {
		case err == nil:
			if apistructs.PutOnOpusModeAppend.Equal(req.GetMode()) {
				l.Warnln("failed to put on extension, the version already exists, can not append")
				return nil, apierr.PutOnExtension.AlreadyExists()
			}
			return p.updateExtension(ctx, l, userID, &opus, &version, req)
		case errors.Is(err, gorm.ErrRecordNotFound):
			return p.createExtensions(ctx, l, userID, &opus, req)
		default:
			l.WithError(err).WithFields(where).Errorln("failed to First version")
			return nil, apierr.PutOnExtension.InternalError(err)
		}

	case errors.Is(err, gorm.ErrRecordNotFound):
		return p.createExtensions(ctx, l, userID, nil, req)
	default:
		l.WithError(err).WithFields(where).Errorln("failed to First opus")
		return nil, apierr.PutOnExtension.InternalError(err)
	}
}

func (p *provider) listOpusByIDs(_ context.Context, pageSize, pageNo int, opusesIDs []string) (*pb.ListOpusResp, error) {
	var (
		l        = p.l.WithField("func", "listOpusByIDs")
		opuses   []*model.Opus
		versions []*model.OpusVersion
		total    int64
	)

	if err := p.D.Limit(pageSize).Offset((pageNo-1)*pageSize).
		Where("id IN (?)", opusesIDs).
		Find(opuses).
		Count(&total).
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find opuses")
		return nil, apierr.ListOpus.InternalError(err)
	}
	if err := p.D.Where("opus_id IN (?)", opusesIDs).
		Find(&versions).
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find versions")
		return nil, apierr.ListOpus.InternalError(err)
	}

	var versionsMap = make(map[string]map[string]*model.OpusVersion)
	for _, version := range versions {
		m, ok := versionsMap[version.OpusID]
		if !ok {
			m = make(map[string]*model.OpusVersion)
		}
		m[version.ID.String] = version
		versionsMap[version.OpusID] = m
	}

	var result = pb.ListOpusResp{Data: &pb.ListOpusRespData{Total: int32(total)}}
	for _, opus := range opuses {
		item := pb.ListOpusRespDataItem{
			Id:          opus.ID.String,
			CreatedAt:   timestamppb.New(opus.CreatedAt),
			UpdatedAt:   timestamppb.New(opus.UpdatedAt),
			OrgID:       strconv.FormatUint(uint64(opus.OrgID), 10),
			OrgName:     opus.OrgName,
			CreatorID:   opus.CreatorID,
			UpdaterID:   opus.UpdaterID,
			Type:        opus.Type,
			TypeName:    apistructs.OpusTypes[apistructs.OpusType(opus.Type)],
			Name:        opus.Name,
			DisplayName: opus.DisplayName,
			Summary:     "",
			Catalog:     opus.Catalog,
			LogoURL:     "",
		}
		if m, ok := versionsMap[opus.ID.String]; ok {
			for k := range m {
				item.Summary = m[k].Summary
				item.LogoURL = m[k].LogoURL
				if k == opus.DefaultVersionID {
					break
				}
			}
		}
		result.Data.List = append(result.Data.List, &item)
	}
	return &result, nil
}

func (p *provider) listOpusWithKeyword(ctx context.Context, pageSize, pageNo int, keyword string, opusesIDs []string) (*pb.ListOpusResp, error) {
	var (
		l        = p.l.WithField("func", "listOpusWithKeyword")
		opuses   []*model.Opus
		versions []*model.OpusVersion
	)

	keyword = "%" + keyword + "%"
	if err := p.D.Where("opus_id IN (?)", opusesIDs).
		Where("summary LIKE ?", keyword).
		Find(&versions).
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find versions")
		return nil, apierr.ListOpus.InternalError(err)
	}
	if err := p.D.Where("id IN (?)", opusesIDs).
		Where("name LIKE ? OR display_name LIKE ?", keyword, keyword).
		Find(&opuses).
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.WithError(err).Errorln("failed to Find opuses")
		return nil, apierr.ListOpus.InternalError(err)
	}

	var opusesIDsMap = make(map[string]struct{})
	for _, version := range versions {
		opusesIDsMap[version.OpusID] = struct{}{}
	}
	for _, opus := range opuses {
		opusesIDsMap[opus.ID.String] = struct{}{}
	}

	opusesIDs = nil
	for k := range opusesIDsMap {
		opusesIDs = append(opusesIDs, k)
	}

	return p.listOpusByIDs(ctx, pageSize, pageNo, opusesIDs)
}

func (p *provider) updateExtension(ctx context.Context, l *logrus.Entry, userID string, opus *model.Opus, version *model.OpusVersion, req *pb.PubOnExtensionsReq) (*commonPb.VoidResponse, error) {
	l = l.WithField("func", "updateExtension")

	var err error
	tx := p.D.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// update version
	labels, _ := json.Marshal(req.GetLabels())
	if err = tx.Model(&version).
		Where(map[string]interface{}{"id": version.ID}).
		Updates(map[string]interface{}{
			"summary":    req.GetSummary(),
			"labels":     string(labels),
			"logo_url":   req.GetLogoURL(),
			"updater_id": userID,
		}).Error; err != nil {
		l.WithError(err).Errorln("failed to Updates")
		return nil, apierr.PutOnExtension.InternalError(err)
	}

	// update presentation
	if err = tx.Model(new(model.OpusPresentation)).
		Where(map[string]interface{}{"version_id": version.ID}).
		Updates(map[string]interface{}{
			"desc":              req.GetDesc(),
			"contact_name":      req.GetContactName(),
			"contact_url":       req.GetContactURL(),
			"contact_email":     req.GetContactEmail(),
			"is_open_sourced":   req.GetIsOpenSourced(),
			"opensource_url":    req.GetOpensourceURL(),
			"licence_name":      req.GetLicenseName(),
			"license_url":       req.GetLicenseURL(),
			"homepage_name":     req.GetHomepageName(),
			"homepage_url":      req.GetHomepageURL(),
			"homepage_logo_url": req.GetHomepageLogoURL(),
			"is_downloadable":   req.GetIsDownloadable(),
			"download_url":      req.GetDownloadURL(),
			"updater_id":        userID,
		}).Error; err != nil {
		l.WithError(err).Errorln("failed to Updates presentation")
		return nil, apierr.PutOnExtension.InternalError(err)
	}

	// create or update readmes
	for _, item := range req.GetReadme() {
		if item.GetLang() == "" {
			continue
		}
		var (
			readme model.OpusReadme
			where  = map[string]interface{}{
				"version_id": version.ID,
				"lang":       item.GetLang(),
			}
		)
		switch err2 := p.D.Where(where).First(&readme).Error; {
		case err2 == nil:
			if err = tx.Model(new(model.OpusReadme)).
				Where(where).
				Updates(map[string]interface{}{
					"lang_name": apistructs.LangTypes[apistructs.Lang(item.GetLang())],
					"text":      item.GetText(),
				}).Error; err != nil {
				l.WithError(err).WithField("lang", item.GetLang()).Errorln("failed to Updates readme")
				return nil, apierr.PutOnExtension.InternalError(err)
			}
		case errors.Is(err2, gorm.ErrRecordNotFound):
			if err = tx.Create(model.OpusReadme{
				Common: model.Common{
					OrgID:     opus.OrgID,
					OrgName:   opus.OrgName,
					CreatorID: userID,
					UpdaterID: userID,
				},
				OpusID:    version.OpusID,
				VersionID: version.ID.String,
				Lang:      item.GetLang(),
				LangName:  apistructs.LangTypes[apistructs.Lang(item.GetLang())],
				Text:      item.GetText(),
			}).Error; err != nil {
				l.WithError(err).WithField("lang", item.GetLang()).Errorln("failed to Create readme")
				return nil, apierr.PutOnExtension.InternalError(err)
			}
		default:
			l.WithError(err2).WithFields(where).Errorln("failed to First readme")
			return nil, apierr.PutOnExtension.InternalError(err2)
		}
	}

	// update opus
	updates := map[string]interface{}{
		"updater_id": userID,
	}
	if req.GetIsDefault() {
		updates["default_version_id"] = version.ID
	}
	if err = tx.Model(opus).Where(map[string]interface{}{"id": opus.ID}).
		Updates(updates).Error; err != nil {
		l.WithError(err).Errorln("failed to Updates opus")
		return nil, apierr.PutOnExtension.InternalError(err)
	}

	return new(commonPb.VoidResponse), nil
}

func (p *provider) createExtensions(ctx context.Context, l *logrus.Entry, userID string, opus *model.Opus, req *pb.PubOnExtensionsReq) (*commonPb.VoidResponse, error) {
	l = l.WithField("func", "createExtensions")

	var err error
	tx := p.D.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// create version
	common := model.Common{
		OrgID:     opus.OrgID,
		OrgName:   opus.OrgName,
		CreatorID: userID,
		UpdaterID: userID,
	}

	if opus == nil {
		opus = &model.Opus{
			Model:       model.Model{},
			Common:      common,
			Level:       req.GetLevel(),
			Type:        req.GetType(),
			Name:        req.GetName(),
			DisplayName: req.GetDisplayName(),
			Catalog:     req.GetCatalog(),
		}
		if err = tx.Create(opus).Error; err != nil {
			l.WithError(err).Errorln("failed to Create opus")
			return nil, apierr.PutOnExtension.InternalError(err)
		}
	}
	labels, _ := json.Marshal(req.GetLabels())
	var version = model.OpusVersion{
		Common:  common,
		OpusID:  opus.ID.String,
		Version: req.GetVersion(),
		Summary: req.GetSummary(),
		Labels:  string(labels),
		LogoURL: req.GetLogoURL(),
		IsValid: true,
	}
	if err = tx.Create(&version).Error; err != nil {
		l.WithError(err).Errorln("failed to Create version")
		return nil, apierr.PutOnExtension.InternalError(err)
	}

	// create presentation
	var presentation = model.OpusPresentation{
		Common:          common,
		OpusID:          opus.ID.String,
		VersionID:       version.ID.String,
		Ref:             "",
		Desc:            req.GetDesc(),
		ContactName:     req.GetContactName(),
		ContactURL:      req.GetContactURL(),
		ContactEmail:    req.GetContactEmail(),
		IsOpenSourced:   req.GetIsOpenSourced(),
		OpensourceURL:   req.GetOpensourceURL(),
		LicenceName:     req.GetLicenseName(),
		LicenceURL:      req.GetLicenseURL(),
		HomepageName:    req.GetHomepageName(),
		HomepageURL:     req.GetHomepageURL(),
		HomepageLogoURL: req.GetHomepageLogoURL(),
		IsDownloadable:  req.GetIsDownloadable(),
		DownloadURL:     req.GetDownloadURL(),
	}
	if err = tx.Create(&presentation).Error; err != nil {
		l.WithError(err).Errorln("failed to Create presentation")
		return nil, apierr.PutOnExtension.InternalError(err)
	}

	// create readmes
	var readmes []*model.OpusReadme
	for _, item := range req.GetReadme() {
		if item.GetLang() == "" {
			continue
		}
		readme := model.OpusReadme{
			Common:    common,
			OpusID:    opus.ID.String,
			VersionID: version.ID.String,
			Lang:      item.GetLang(),
			LangName:  apistructs.LangTypes[apistructs.Lang(item.GetLang())],
			Text:      item.GetText(),
		}
		readmes = append(readmes, &readme)
	}
	if err = tx.CreateInBatches(readmes, len(readmes)).Error; err != nil {
		l.WithError(err).Errorln("failed to CreateInBatches readmes")
		return nil, apierr.PutOnExtension.InternalError(err)
	}

	// update opus
	updates := map[string]interface{}{
		"updater_id": userID,
	}
	if req.GetIsDefault() {
		updates["default_version_id"] = version.ID
	}
	if err = tx.Model(opus).Where(map[string]interface{}{"id": opus.ID}).Updates(updates).Error; err != nil {
		l.WithError(err).Errorln("failed to Updates opus")
		return nil, apierr.PutOnExtension.InternalError(err)
	}

	return new(commonPb.VoidResponse), nil
}
