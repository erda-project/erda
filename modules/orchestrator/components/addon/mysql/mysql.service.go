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

package mysql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
	"github.com/erda-project/erda/pkg/mysqlhelper"
	"github.com/erda-project/erda/pkg/strutil"
)

type mysqlService struct {
	logger logs.Logger

	bdl *bundle.Bundle
	db  *dbclient.DBClient
}

// ListMySQLAccount returns a list of MySQL accounts
func (s *mysqlService) ListMySQLAccount(ctx context.Context, req *pb.ListMySQLAccountRequest) (*pb.ListMySQLAccountResponse, error) {
	accounts, err := s.db.GetMySQLAccountListByRoutingInstanceID(req.InstanceId)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.MySQLAccount, len(accounts))
	for i, acc := range accounts {
		res[i] = s.ToDTO(&acc, true)
	}
	return &pb.ListMySQLAccountResponse{Accounts: res}, nil
}

func (s *mysqlService) ToDTO(acc *dbclient.MySQLAccount, decrypt bool) *pb.MySQLAccount {
	pass := "******"
	if decrypt {
		p, err := s.decrypt(acc)
		if err != nil {
			s.logger.Errorf("pass decrypt failed, mySQLAccountID: %s, err: %+v", acc.ID, err)
		}
		pass = p
	}
	return &pb.MySQLAccount{
		Id:         acc.ID,
		InstanceId: acc.RoutingInstanceID,
		Creator:    acc.Creator,
		CreateAt:   timestamppb.New(acc.CreatedAt),
		Username:   acc.Username,
		Password:   pass,
	}
}

func (s *mysqlService) execSql(ins *dbclient.AddonInstance, sql ...string) error {
	if len(sql) == 0 {
		return nil
	}
	var mysqlConfig struct {
		Host string `json:"MYSQL_HOST"`
		Port string `json:"MYSQL_PORT"`
		User string `json:"MYSQL_USERNAME"`
		Pass string `json:"MYSQL_PASSWORD"`
	}
	if err := json.Unmarshal([]byte(ins.Config), &mysqlConfig); err != nil {
		return err
	}
	if mysqlConfig.Host == "" || mysqlConfig.Port == "" || mysqlConfig.User == "" || mysqlConfig.Pass == "" {
		return fmt.Errorf("missing key MYSQL_HOST | MYSQL_PASSWORD | MYSQL_PORT | MYSQL_USERNAME")
	}
	decryptData, err := s.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
		DecryptRequest: kmstypes.DecryptRequest{
			KeyID:            ins.KmsKey,
			CiphertextBase64: mysqlConfig.Pass,
		},
	})
	if err != nil {
		return err
	}
	b, err := base64.StdEncoding.DecodeString(decryptData.PlaintextBase64)
	if err != nil {
		return err
	}
	pass := string(b)

	var req mysqlhelper.Request
	req.Url = mysqlConfig.Host + ":" + mysqlConfig.Port
	req.User = mysqlConfig.User
	req.Password = pass
	req.Sqls = sql
	req.ClusterKey = ins.Cluster
	return req.Exec()
}

func (s *mysqlService) GenerateMySQLAccount(ctx context.Context, req *pb.GenerateMySQLAccountRequest) (*pb.GenerateMySQLAccountResponse, error) {
	//userID := apis.GetUserID(ctx)
	// TODO: apis.GetUserID(ctx) not working
	userID := req.UserID
	if userID == "" {
		return nil, errors.Errorf("user not login")
	}
	// TODO: check permission
	// (userID) has permission to operate on (instanceID) in (orgID)

	routingInstance, err := s.db.GetInstanceRouting(req.InstanceId)
	if err != nil {
		return nil, err
	}
	ins, err := s.db.GetAddonInstance(routingInstance.RealInstance)
	if err != nil {
		return nil, err
	}

	user := "mysql-" + strutil.RandStr(6)
	pass := strutil.RandStr(12)

	sql := []string{
		fmt.Sprintf(`CREATE USER '%s'@'%%' IDENTIFIED by '%s';`, user, pass),
		fmt.Sprintf(`GRANT ALL ON *.* TO '%s'@'%%';`, user),
		"flush privileges;",
	}

	if err := s.execSql(ins, sql...); err != nil {
		return nil, err
	}

	kr, err := s.bdl.KMSCreateKey(apistructs.KMSCreateKeyRequest{
		CreateKeyRequest: kmstypes.CreateKeyRequest{
			PluginKind: kmstypes.PluginKind_DICE_KMS,
		},
	})
	if err != nil {
		return nil, err
	}
	encryptData, err := s.bdl.KMSEncrypt(apistructs.KMSEncryptRequest{
		EncryptRequest: kmstypes.EncryptRequest{
			KeyID:           kr.KeyMetadata.KeyID,
			PlaintextBase64: base64.StdEncoding.EncodeToString([]byte(pass)),
		},
	})
	if err != nil {
		return nil, err
	}

	account := &dbclient.MySQLAccount{
		Username:          user,
		Password:          encryptData.CiphertextBase64,
		KMSKey:            kr.KeyMetadata.KeyID,
		InstanceID:        routingInstance.RealInstance,
		RoutingInstanceID: req.InstanceId,
		Creator:           userID,
	}
	if err := s.db.CreateMySQLAccount(account); err != nil {
		return nil, err
	}
	dto := s.ToDTO(account, false)
	dto.Password = pass
	return &pb.GenerateMySQLAccountResponse{
		Account: dto,
	}, nil
}

func (s *mysqlService) DeleteMySQLAccount(ctx context.Context, req *pb.DeleteMySQLAccountRequest) (*pb.DeleteMySQLAccountResponse, error) {
	// TODO: do real delete mysql account
	account, err := s.db.GetMySQLAccountByID(req.Id)
	if err != nil {
		return nil, err
	}
	account.IsDeleted = true
	if err := s.db.UpdateMySQLAccount(account); err != nil {
		return nil, err
	}
	return &pb.DeleteMySQLAccountResponse{}, nil
}

func (s *mysqlService) ListAttachment(ctx context.Context, req *pb.ListAttachmentRequest) (*pb.ListAttachmentResponse, error) {
	attachments, err := s.db.GetAttachmentsByRoutingInstanceID(req.InstanceId)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.Attachment, len(*attachments))
	for i, att := range *attachments {
		cf, err := s.getConfig(&att)
		if err != nil {
			s.logger.Errorf("get config of instance %s failed, err: %+v", att.InstanceID, err)
		}
		res[i] = &pb.Attachment{
			Id:           att.ID,
			InstanceId:   att.RoutingInstanceID,
			AppId:        att.ApplicationID,
			RuntimeId:    att.RuntimeID,
			RuntimeName:  att.RuntimeName,
			AccountId:    att.MySQLAccountID,
			PreAccountId: att.PreviousMySQLAccountID,
			AccountState: att.MySQLAccountState,
			Configs:      cf,
		}
	}
	return &pb.ListAttachmentResponse{Attachments: res}, nil
}

func (s *mysqlService) UpdateAttachmentAccount(ctx context.Context, req *pb.UpdateAttachmentAccountRequest) (*pb.UpdateAttachmentAccountResponse, error) {
	att, err := s.db.GetAttachmentByID(req.Id)
	if err != nil {
		return nil, err
	}
	if req.AccountId == att.MySQLAccountID {
		// no need to update
		return &pb.UpdateAttachmentAccountResponse{}, nil
	}
	if att.MySQLAccountState == "PRE" {
		// in switching
		if req.AccountId == att.PreviousMySQLAccountID {
			// switch back
			att.MySQLAccountState = "CUR"
			att.MySQLAccountID = req.AccountId
			att.PreviousMySQLAccountID = ""
		} else {
			// switch to new account
			att.MySQLAccountID = req.AccountId
		}
	} else {
		// normal state
		att.MySQLAccountState = "PRE"
		att.PreviousMySQLAccountID = att.MySQLAccountID
		att.MySQLAccountID = req.AccountId
	}
	if err := s.db.UpdateAttachment(att); err != nil {
		return nil, err
	}
	return &pb.UpdateAttachmentAccountResponse{}, nil
}

func (s *mysqlService) decrypt(acc *dbclient.MySQLAccount) (string, error) {
	pass := "***fail***"
	r, err := s.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
		DecryptRequest: kmstypes.DecryptRequest{
			KeyID:            acc.KMSKey,
			CiphertextBase64: acc.Password,
		},
	})
	if err != nil {
		return pass, err
	}
	decodePasswordStr, err := base64.StdEncoding.DecodeString(r.PlaintextBase64)
	if err != nil {
		return pass, err
	} else {
		pass = string(decodePasswordStr)
	}
	return pass, nil
}

func (s *mysqlService) getConfig(att *dbclient.AddonAttachment) (map[string]string, error) {
	ins, err := s.db.GetAddonInstance(att.InstanceID)
	if err != nil {
		return nil, err
	}
	var configMap map[string]string
	if err := json.Unmarshal([]byte(ins.Config), &configMap); err != nil {
		return nil, err
	}
	if att.MySQLAccountID != "" {
		acc, err := s.db.GetMySQLAccountByID(att.MySQLAccountID)
		if err != nil {
			return nil, err
		}
		pass, err := s.decrypt(acc)
		configMap["MYSQL_USERNAME"] = acc.Username
		configMap["MYSQL_PASSWORD"] = pass
	} else {
		configMap["MYSQL_PASSWORD"] = "******"
	}
	return configMap, nil
}
