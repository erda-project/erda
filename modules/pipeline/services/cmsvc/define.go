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

package cmsvc

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/cms"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/encryption"
)

type CMSvc struct {
	bdl *bundle.Bundle
	cm  cms.ConfigManager

	rsaCrypt *encryption.RsaCrypt
}

func New(bdl *bundle.Bundle, dbClient *dbclient.Client, ops ...Option) *CMSvc {
	s := CMSvc{}
	s.bdl = bdl
	s.rsaCrypt = encryption.NewRSAScrypt(encryption.RSASecret{
		PublicKey:          conf.CmsBase64EncodedRsaPublicKey(),
		PublicKeyDataType:  encryption.Base64,
		PrivateKey:         conf.CmsBase64EncodedRsaPrivateKey(),
		PrivateKeyDataType: encryption.Base64,
		PrivateKeyType:     encryption.PKCS1,
	})

	for _, op := range ops {
		op(&s)
	}

	s.cm = cms.NewPipelineCms(dbClient, s.rsaCrypt)
	return &s
}

type Option func(*CMSvc)

func WithRsaCrypt(rsaCrypt *encryption.RsaCrypt) Option {
	return func(s *CMSvc) {
		s.rsaCrypt = rsaCrypt
	}
}
