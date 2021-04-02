package filesvc

import (
	"context"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

var initKmsOnce sync.Once
var kmsKey string

const etcdKmsFilesCMK = "/dice/cmdb/files/kms/key"

type FileService struct {
	db         *dao.DBClient
	bdl        *bundle.Bundle
	etcdClient *etcd.Store
}

// Option 定义 FileService 配置选项
type Option func(*FileService)

// New 新建 Issue 实例
func New(options ...Option) *FileService {
	svc := &FileService{}
	for _, op := range options {
		op(svc)
	}

	initKmsOnce.Do(func() {
		// 检查 etcd 是否有 key
		kv, err := svc.etcdClient.Get(context.Background(), etcdKmsFilesCMK)
		if err != nil {
			if err.Error() == "not found" {
				ApplyKmsCmk(svc)
				return
			}
			panic(err)
		}
		if len(kv.Value) == 0 {
			ApplyKmsCmk(svc)
			return
		}
		kmsKey = string(kv.Value)
		_, err = svc.bdl.KMSDescribeKey(apistructs.KMSDescribeKeyRequest{
			DescribeKeyRequest: kmstypes.DescribeKeyRequest{
				KeyID: kmsKey,
			},
		})
		if err != nil {
			logrus.Errorf("cmdb files kms cmk describe failed, keyID: %s", kmsKey)
			ApplyKmsCmk(svc)
			return
		}
		logrus.Infof("cmdb files kms cmk: %s", kmsKey)
		return
	})

	// clean expired files
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			_ = svc.CleanExpiredFiles()
		}
	}()

	return svc
}

// WithDBClient 配置 FileService 数据库选项
func WithDBClient(db *dao.DBClient) Option {
	return func(svc *FileService) {
		svc.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(svc *FileService) {
		svc.bdl = bdl
	}
}

func WithEtcdClient(etcdClient *etcd.Store) Option {
	return func(svc *FileService) {
		svc.etcdClient = etcdClient
	}
}

func GetKMSKey() string {
	return kmsKey
}

func ApplyKmsCmk(svc *FileService) {
	kmsResp, err := svc.bdl.KMSCreateKey(apistructs.KMSCreateKeyRequest{
		CreateKeyRequest: kmstypes.CreateKeyRequest{
			PluginKind:            kmstypes.PluginKind_DICE_KMS,
			CustomerMasterKeySpec: kmstypes.CustomerMasterKeySpec_SYMMETRIC_DEFAULT,
			KeyUsage:              kmstypes.KeyUsage_ENCRYPT_DECRYPT,
			Description:           "cmdb files kms key",
		},
	})
	if err != nil {
		panic(err)
	}
	keyID := kmsResp.KeyMetadata.KeyID
	defer func() {
		kmsKey = keyID
		logrus.Infof("cmdb files kms cmk: %s", kmsKey)
	}()
	_, err = svc.etcdClient.GetClient().Txn(context.Background()).
		If(clientv3.Compare(clientv3.Version(etcdKmsFilesCMK), "=", 0)).
		Then(clientv3.OpPut(etcdKmsFilesCMK, keyID)).
		Commit()
	if err != nil {
		panic(err)
	}
	return
}
