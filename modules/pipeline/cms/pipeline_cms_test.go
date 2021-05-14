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

package cms_test

//import (
//	"context"
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/modules/pipeline/cms"
//	"github.com/erda-project/erda/modules/pipeline/dbclient"
//	"github.com/erda-project/erda/pkg/encryption"
//	"github.com/erda-project/erda/pkg/strutil"
//)
//
//var (
//	ctxWithDiceSource    = context.WithValue(context.Background(), cms.CtxKeyPipelineSource, apistructs.PipelineSourceDice)
//	ctxWithCDPDevSource  = context.WithValue(context.Background(), cms.CtxKeyPipelineSource, apistructs.PipelineSourceCDPDev)
//	ctxWithDefaultSource = context.WithValue(context.Background(), cms.CtxKeyPipelineSource, apistructs.PipelineSourceDefault)
//)
//
//var (
//	// 以下临时 publicKey/primaryKey 仅测试使用
//	rsaCrypt = encryption.NewRSAScrypt(encryption.RSASecret{
//		PublicKey:          "LS0tLS1CRUdJTiBwdWJsaWMga2V5LS0tLS0KTUlJQ0lqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FnOEFNSUlDQ2dLQ0FnRUEyT2VNdndqL3BLOGNpSjhPRWlvUgovSi9HTGpKV0phaHNQTW5mRmR0QmZmTHJmdEJxSzlyYW84clVuakZqWnc4a0RxYUp5dGRDSE5aN3VxMStGUWVxCmVTWGU2aTVvc1oyVms4TmJlM3RiTnlhVnpjS3V5L3dXYkFIeTNSVDh3YmJUQllZV2kzUnRpbTVETFhDbEdjV1UKUDlGL25CaDVCTG0xWXVDZlNuUU5kdE9hM0owRytrbjVEMGFDT2cwRFNMUnFpUDZDMklSNFMvTjNYWjBQQVFKeApCZ3ZSRldHN2ErMGkyWHR4RXZSUU5zaVV4MlV4Qk1nSE9BSHJHRW5tKzJsMkZaQVFUTU5kTU82cWVhYjJYMWNECmpra2xOTjhGTGlMazhPNURRby8xdGEzQ0pQU0pYbi9uY2E1a0Z1SktFU3kyUy9KbXlROG9PQzFtV3lFeVhRYmEKUEJNY2xIYnllYklmWnhMQysvQkhvY0h0L01LN2xlalRwZnBuSytCQXZDaFFLaEZsaDBQVTF6MnRrRVhjNE5VQgppTGI4dCtvUitJUHVYK0d2TjJ3bUJpUDV0U2lHNHRCbGx2NFlleEQxRmU2Y1JLeGRzSXlyKzAxM05hVHdjOVRjClZUbm5RY0tsZFRPS2t2MVR0NGRRQ25pdlRxNSt4UURrOXprK1VNY2RNeVdqQTd5YTVsWWFXR05hQ3I5MkJQcUMKUWR1R052eVMrc3FYdnZuaHZuRWN2OHpHbWsxMDg5dm1yU3BNUUQ5Rk1GWEwvK1YyTVVxRCtnNEZ4N2VIbmRyRwozdlZXVjJVbThhVEJ4WEFyd2RObUlCdEk2aDM0eTlFbE9uN0VVYmFqaDRlc2RoVFBMTzVnYjRJUVFwS2JhUmp1Cm53UW5xRXR4TFRQcVZxNTJBOG5PM2xFQ0F3RUFBUT09Ci0tLS0tRU5EIHB1YmxpYyBrZXktLS0tLQo=",
//		PublicKeyDataType:  encryption.Base64,
//		PrivateKey:         "LS0tLS1CRUdJTiBwcml2YXRlIGtleS0tLS0tCk1JSUpLUUlCQUFLQ0FnRUEyT2VNdndqL3BLOGNpSjhPRWlvUi9KL0dMakpXSmFoc1BNbmZGZHRCZmZMcmZ0QnEKSzlyYW84clVuakZqWnc4a0RxYUp5dGRDSE5aN3VxMStGUWVxZVNYZTZpNW9zWjJWazhOYmUzdGJOeWFWemNLdQp5L3dXYkFIeTNSVDh3YmJUQllZV2kzUnRpbTVETFhDbEdjV1VQOUYvbkJoNUJMbTFZdUNmU25RTmR0T2EzSjBHCitrbjVEMGFDT2cwRFNMUnFpUDZDMklSNFMvTjNYWjBQQVFKeEJndlJGV0c3YSswaTJYdHhFdlJRTnNpVXgyVXgKQk1nSE9BSHJHRW5tKzJsMkZaQVFUTU5kTU82cWVhYjJYMWNEamtrbE5OOEZMaUxrOE81RFFvLzF0YTNDSlBTSgpYbi9uY2E1a0Z1SktFU3kyUy9KbXlROG9PQzFtV3lFeVhRYmFQQk1jbEhieWViSWZaeExDKy9CSG9jSHQvTUs3CmxlalRwZnBuSytCQXZDaFFLaEZsaDBQVTF6MnRrRVhjNE5VQmlMYjh0K29SK0lQdVgrR3ZOMndtQmlQNXRTaUcKNHRCbGx2NFlleEQxRmU2Y1JLeGRzSXlyKzAxM05hVHdjOVRjVlRublFjS2xkVE9La3YxVHQ0ZFFDbml2VHE1Kwp4UURrOXprK1VNY2RNeVdqQTd5YTVsWWFXR05hQ3I5MkJQcUNRZHVHTnZ5UytzcVh2dm5odm5FY3Y4ekdtazEwCjg5dm1yU3BNUUQ5Rk1GWEwvK1YyTVVxRCtnNEZ4N2VIbmRyRzN2VldWMlVtOGFUQnhYQXJ3ZE5tSUJ0STZoMzQKeTlFbE9uN0VVYmFqaDRlc2RoVFBMTzVnYjRJUVFwS2JhUmp1bndRbnFFdHhMVFBxVnE1MkE4bk8zbEVDQXdFQQpBUUtDQWdBT3lacDYyNjR5R0E0bDhsSVBRdmIrOWhXWXlLMisyNENsbEUyMU84RjNTTHh0Wk9BWUpVK0tveVZqCnM1SkhVR3p3NHNHNkpuckhaSWdDN2hrT2JmdGRUd3VuZzRwM3NYcWxIRWg4WHFpVlZmZ1lreEUvcnV3SWFRbVoKc1BpYWJGQnVxL21WZ0ZhSGZZVHU4Q01SWXJyOHJ0ZTRXS0xIZzdHdUVBcE1GU1ZsMkg5U1V4Skt0Z2hZMWtIQwptMmlCNkdycTlBOFBtOWhudFMyS0lFOEpqcFVPQ0hnMHNQa0tIcHlsbnhqU1pmMmgvb0xHSlV2Mk8zemlnSjc2CmhPOU9iSjQwVWlJS1diZGN3cWkwcW9GWmRxRXpiaUV2UFpVbzFCQXZyTTdCRnZkMWIyY3hCY3JudW1pWkEzNm4KWUw0VDlheG4rUnF3MG11M2lNRFZyYW0xVmVaQzJnazVuOUVFRTV2eVdZd1ZIN08rbDltTUVRdmx3dFFQeHRocwpjMU1Telo2L2tISjlRemRCb0oya29IVVpqbjgzWGs4TWJ3aDJPbk1yWnVnQW1HWVlRNjhKTXN4QjRVcHJuZVlPClkvclBocFRqWGppd3Baa1ZnT0VadWZLK0ZadHo0UmsweWN1M0NSYXdRNkp0VVU5bHRtYUNkMG5NeVFmWEwyN2gKbWR3NmVneFRKYWhrMWVDdjBkMGhZVXRpOUNKczZZWS9FaXZYYWdhd29VNmlrU3pycHlnSHpSUzFaZ2o0Z205Sgp3MUtQT2Jxb28vMUFxdjk4aFdqSHpINDNxbS94a28yVnhzMUs5WTdLclNpbmNyUVZtY1NMVmJPbENNMWxucGUxClRoYzJ0eXpmbGFtRkdJc0JnWFZhWE5NSVhpZUY1b0t4MHVYMnkxWTI4anl4REZPMVVRS0NBUUVBK3RkODBMQXAKeVhaQVRXRUYxNk5hdGg0d0xJbzIycnpLYmZyRXN5M2d0UVdtRVZlR1VGS0RGZjRrbEw1bzd6eUYvdW5nQmg2UwpOZU9MU3VLTStOd2dDMlhSQ2xtcmlDMjlMaVBQT2FraU9KVkxYWk9Wd0c5YjBHNEpLYVl3Umd2emdvaEFveG5ICi9IQnUwNm1UbGRHTjUyOU5UWWFQeXcrUGF1N1prWE1raUM4NDFNN2hFOGJ6Vy81ZlRycnh2dnQxM3pNcDN2YjYKMUpiMWMxQ284cExGZVJGK2Rzd0tJeVFDWFF1Q2JEZGt5WWJ3cmdYbitXeDk4c25QZjI4MXN0TmhFdGJ5Y3lKTgozaFhLMEZGOEMxQ2FMMmxtV3pIaHkrZmVqVzlJN1hTN1lKc01BdkJEZEdNdFNLZUtqS2o2VEN4YUJaOGlXYllCCjBYci9ETjdPNHhVK1J3S0NBUUVBM1YxbnpmeHFzQXoyeVNPeTlSRjNxQnVxRDI4WEtmSUo2QllLWHRXSzkzRXUKR2hCdFFoY0ZqVUdSMkhDaFp6Y0p2SVRTQ3FHOHhTTjU4U0h6aVpnK0FGTDdLUmdYa05wS2phQVYxaytsdFZEagpsaEFTTDlWUDJUK1NFTEVxam44cU9NeHlaZUJ0ZEtkNG1WcFVVUVg5dVpMcnUzdUczZ1VRVEhrZVBrUzg1UWJjCmNHcnVxelNXS0U1cmxvQ2VZQjMveU01Um5qNmhiZm5DYTh0TjRuNDk3cWVsZytVTDhUdkNxODJ0M0c2WU1hSXoKdkFkSUtBOElVUVNxcm5mMzBWZHFmWG1sVXdSUFJjVHlZR2dhZ0pya1BWY1dGT0pDZGR3THVuc1Z2eEdsRGVHVgo3WmlmajRiaHRld05ZblU2dE80S2IzelF5KzErUm1xZ3R5ZWlaTHZTcHdLQ0FRQTA1VlFRdmRWU0FubTAxNHpmCjJEYTh5TWpuMjQyTnV0b0ZMeWhqa0gwZUx6N0IwVzhsYVFEemxsQW9mYTZySkZ3dFVTeEluaEcvQTJqUU5jMzgKZkk3VldIY29jNWhVY3pDOWxoZVExVFcrTU4xZnNrdVY5T1dyb2tpVVc5TTZNakw3aDdmNXJPb2JOYXBwUUEwNwpQcUZwK0hLWXNwT0lBcFAvdkxac2taZFdrSDZ2Z2FDOUJ1c3lyd1Z5R01INXdCVXZLQjdnUWJ6TEw3bzZ3dnVkCmk0M1E1ZnVCR2EzWmt6SmNaSnp3TFE0MzRSakgyYjc5UGYraFB5VmVmaGtZeUxKandxZ0YxMm9NTnhRNXNiVXkKdmFDRjl2ZjZxeDR1WFlyMDBFN1VwQlVQWGlLK1MrUXRtdXhsc2M3cHNvaDFuN1NzRXM2dmxFMzEycllHQk1Zago1TXJwQW9JQkFRQ1EzUGQ1alo0ajU5ZjRlU3c3eEZxUjRNakJvT2wvd2ExSi9HSjgvVEljREMwblVXaXV2M0lhCnByWlM5aUlwOFpLbGxDWUFYeWV4dXkycDU1WUFqV2pGdllndnRGeDNwdUx6RzdndXI2QzVyMTNBYm5QNGFaZi8KaStLQ21lNUhvbUIzR3hRaUoyUjUycjdKWEp1aENsS29oc1ZOdytEV21tbTRJZXJ3eFByNHhpeXNSTTQ3ckFZNApDbG5OL0Eyb3lQa0M2RUh4VlBzL2hScitmK1ZRTzEwOE9PblFEcXhxQ2JtenhMM3FhMVdkNVpBRmxKNWIyTHFRCmlvVkg5NnB0ak05Ym5hZmJWQTZza2Q0cnlQVFBCSTRvdGp6MUhieHdkTGdZS1VScDdab2VJMnFDT2tieEhrdkgKU2RyWWUrOFhTRS83OWFxT1NiVkJUN1l0SmZyWUFUSXhBb0lCQVFEdWhVQkJxWjR3ajNEd0o0aitiS0Z2UWNuSwpDSGgxNEc1N1RxcDdyUjIxdFRWd1U5bDZvMHUxNk9zTkUyaGtzVU9RaDZKbnVnVnE5dGUrelVFb1dLRGRsMmRzCnN6S3FGcWVNV2xrZnNsSWd4ZGsySVRVZ3lscnorbWRUYlJ2c2M0cUZCTlRIOHlqc053RU1PM043blpsbXk4aW4KYUV2bE5INXdjc3ljajczZ09PYjlpYTN3a1g4WkJuaG96T2FSUjVaQjhFa2VFK2FEM3hZWGIxZFpSQlYvRms0dQp2aEthWWgvNzBOZjZGeHFzQU8xdU9BQTE5V0QzdWM4ODNSWitCWFY4aU1tWWhpVjk0aGE4Y213TmJhcUIybzJDClJoR2VqWkdoQ0NKbXZiZHZUNEpuWEVYWk54OGVlWjF4UHExb3VtMVhZVTJLSEgydyszVmQ5ZXJIelhvYQotLS0tLUVORCBwcml2YXRlIGtleS0tLS0tCg==",
//		PrivateKeyDataType: encryption.Base64,
//		PrivateKeyType:     encryption.PKCS1,
//	})
//)
//
//func initPipelineCms() {
//	once.Do(func() {
//		fmt.Println("init pipeline cm")
//		_ = os.Setenv("MYSQL_HOST", "localhost")
//		_ = os.Setenv("MYSQL_PORT", "3306")
//		_ = os.Setenv("MYSQL_USERNAME", "root")
//		_ = os.Setenv("MYSQL_PASSWORD", "anywhere")
//		_ = os.Setenv("MYSQL_DATABASE", "dice")
//		_ = os.Setenv("MYSQL_SHOW_SQL", "true")
//		dbClient, err := dbclient.New()
//		if err != nil {
//			panic(err)
//		}
//		cm = cms.NewPipelineCms(dbClient, rsaCrypt)
//	})
//}
//
//func TestPipelineCm_IdempotentCreateNS(t *testing.T) {
//	initPipelineCms()
//
//	var err error
//
//	err = cm.IdempotentCreateNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//	err = cm.IdempotentCreateNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//
//	err = cm.IdempotentCreateNS(ctxWithDefaultSource, ns)
//	assert.NoError(t, err)
//}
//
//func TestPipelineCm_IdempotentDeleteNS(t *testing.T) {
//	initPipelineCms()
//
//	var err error
//
//	err = cm.IdempotentDeleteNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//	err = cm.IdempotentDeleteNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//	err = cm.IdempotentDeleteNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//
//	err = cm.IdempotentDeleteNS(ctxWithDefaultSource, ns)
//	assert.NoError(t, err)
//
//	err = cm.IdempotentDeleteNS(ctxWithCDPDevSource, ns)
//	assert.NoError(t, err)
//}
//
//func TestPipelineCm_UpdateConfigs(t *testing.T) {
//	initPipelineCms()
//
//	var err error
//
//	kvs := map[string]apistructs.PipelineCmsConfigValue{
//		"a":  {Value: "aa", EncryptInDB: true},
//		"bb": {Value: "bbc", EncryptInDB: false},
//	}
//
//	err = cm.UpdateConfigs(ctxWithDiceSource, ns, kvs)
//	assert.NoError(t, err)
//}
//
//func TestPipelineCm_DeleteConfigs(t *testing.T) {
//	initPipelineCms()
//
//	var err error
//
//	err = cm.DeleteConfigs(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//
//	err = cm.DeleteConfigs(ctxWithDiceSource, ns, []string{"a", "b", "bb", "cc"}...)
//	assert.NoError(t, err)
//
//	err = cm.DeleteConfigs(ctxWithDefaultSource, "not-exist", []string{"a", "b"}...)
//	assert.NoError(t, err)
//}
//
//func TestPipelineCm_GetConfigs(t *testing.T) {
//	initPipelineCms()
//	initPipelineCms()
//
//	var err error
//
//	// delete ns
//	err = cm.IdempotentDeleteNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//	err = cm.IdempotentDeleteNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//
//	// get from ns
//	configs, err := cm.GetConfigs(ctxWithDiceSource, ns, false)
//	assert.NoError(t, err)
//	assert.True(t, len(configs) == 0)
//
//	// create ns
//	err = cm.IdempotentCreateNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//	err = cm.IdempotentCreateNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//
//	// get from ns
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false)
//	assert.NoError(t, err)
//	assert.True(t, len(configs) == 0)
//
//	// update configs
//	newKVs := map[string]apistructs.PipelineCmsConfigValue{
//		"a": {Value: "aa", EncryptInDB: false},
//		"b": {Value: "bb", EncryptInDB: true},
//	}
//	err = cm.UpdateConfigs(ctxWithDiceSource, ns, newKVs)
//	assert.NoError(t, err)
//
//	// get configs
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false)
//	assert.NoError(t, err)
//	assert.True(t, len(configs) == 2)
//
//	// filter get configs
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key:     "a",
//		Decrypt: false,
//	})
//	assert.NoError(t, err)
//	assert.True(t, len(configs) == 1)
//	assert.True(t, configs["a"].Value == "aa")
//	assert.True(t, configs["a"].EncryptInDB == false)
//
//	// delete config
//	err = cm.DeleteConfigs(ctxWithDiceSource, ns, "b")
//	assert.NoError(t, err)
//
//	// filter get config
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key:     "a",
//		Decrypt: false,
//	})
//	assert.NoError(t, err)
//	assert.True(t, len(configs) == 1)
//
//	// filter get config
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key:     "b",
//		Decrypt: false,
//	})
//	assert.NoError(t, err)
//	assert.True(t, len(configs) == 0)
//
//	// delete ns
//	err = cm.IdempotentDeleteNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//	err = cm.IdempotentDeleteNS(ctxWithDiceSource, ns)
//	assert.NoError(t, err)
//}
//
//func TestPipelineCm_encrypt(t *testing.T) {
//	initPipelineCms()
//
//	var err error
//
//	newKVs := map[string]apistructs.PipelineCmsConfigValue{
//		"a": {Value: "aa", EncryptInDB: true},
//	}
//	err = cm.UpdateConfigs(ctxWithDiceSource, ns, newKVs)
//	assert.NoError(t, err)
//
//	aWithoutDecrypt, err := cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key:     "a",
//		Decrypt: false,
//	})
//	assert.NoError(t, err)
//	assert.True(t, len(aWithoutDecrypt) == 1)
//	assert.True(t, aWithoutDecrypt["a"].Value != "aa")
//	assert.True(t, aWithoutDecrypt["a"].EncryptInDB == true)
//
//	aWithDecrypt, err := cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key:     "a",
//		Decrypt: true,
//	})
//	assert.NoError(t, err)
//	assert.True(t, len(aWithDecrypt) == 1)
//	assert.True(t, aWithDecrypt["a"].Value == "aa")
//	assert.True(t, aWithDecrypt["a"].EncryptInDB == true)
//
//	// change a to plaintext
//	newKVs = map[string]apistructs.PipelineCmsConfigValue{
//		"a": {Value: "aa", EncryptInDB: false},
//	}
//	err = cm.UpdateConfigs(ctxWithDiceSource, ns, newKVs)
//	assert.NoError(t, err)
//
//	aWithoutDecrypt, err = cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key: "a",
//	})
//	assert.NoError(t, err)
//	assert.True(t, aWithoutDecrypt["a"].Value == "aa")
//	assert.True(t, aWithoutDecrypt["a"].EncryptInDB == false)
//}
//
//func TestPipelineCm_showEncryptedValue(t *testing.T) {
//	initPipelineCms()
//
//	ns := "local-test-0416"
//
//	var err error
//	var configs map[string]apistructs.PipelineCmsConfigValue
//
//	keyA := "secret.a"
//	valueA := "xxx"
//	keyB := "secret.b"
//	valueB := "yyy"
//
//	newKVs := map[string]apistructs.PipelineCmsConfigValue{
//		keyA: {Value: valueA, EncryptInDB: true, Comment: "一个简单的配置项 A"},
//		keyB: {Value: valueB, EncryptInDB: false, Comment: "一个简单的配置项 B"},
//	}
//	err = cm.UpdateConfigs(ctxWithDiceSource, ns, newKVs)
//	assert.NoError(t, err)
//
//	// 全局解密=false，取所有 key => value=""
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false)
//	assert.NoError(t, err)
//	assert.Empty(t, configs[keyA].Value)
//	assert.NotEmpty(t, configs[keyB].Value)
//
//	// 全局解密=false，取指定 key，配置项级别：decrypt=false, showEncryptedValue=false => value=""
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key:                keyA,
//		Decrypt:            false,
//		ShowEncryptedValue: false,
//	})
//	assert.NoError(t, err)
//	assert.Empty(t, configs[keyA].Value)
//
//	// 全局解密=false，取指定 key，配置项级别：decrypt=true, showEncryptedValue=false => value=xxx
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, false, apistructs.PipelineCmsConfigKey{
//		Key:                keyA,
//		Decrypt:            true,
//		ShowEncryptedValue: false,
//	})
//	assert.NoError(t, err)
//	assert.Equal(t, valueA, configs[keyA].Value)
//
//	// 全局解密=true key，配置项级别：decrypt=false, showEncryptedValue=false => value=""
//	configs, err = cm.GetConfigs(ctxWithDiceSource, ns, true, apistructs.PipelineCmsConfigKey{
//		Key:                keyA,
//		Decrypt:            false,
//		ShowEncryptedValue: false,
//	})
//	assert.NoError(t, err)
//	assert.Empty(t, configs[keyA].Value)
//}
//
//func TestPipelineCm_DiceFiles(t *testing.T) {
//	initPipelineCms()
//
//	var err error
//
//	newKVs := map[string]apistructs.PipelineCmsConfigValue{
//		"a.cert": {
//			Value:       "xxxxxxxxxxxxxxx",
//			EncryptInDB: true,
//			Type:        apistructs.PipelineCmsConfigTypeDiceFile,
//			From:        "syncFromCertsManagement",
//		},
//	}
//	err = cm.UpdateConfigs(ctxWithDiceSource, ns, newKVs)
//	assert.NoError(t, err)
//}
//
//func TestKeyValidator(t *testing.T) {
//	assert.NoError(t, strutil.Validate(`abc`, cms.KeyValidator))
//	assert.NoError(t, strutil.Validate(`abc-`, cms.KeyValidator))
//	assert.NoError(t, strutil.Validate(`abc_`, cms.KeyValidator))
//	assert.NoError(t, strutil.Validate(`abc_-`, cms.KeyValidator))
//	assert.NoError(t, strutil.Validate(`abc_-123`, cms.KeyValidator))
//	assert.NoError(t, strutil.Validate(`_123_-a`, cms.KeyValidator))
//	assert.NoError(t, strutil.Validate(`dice.id`, cms.KeyValidator))
//	assert.NoError(t, strutil.Validate(`dice..id.`, cms.KeyValidator))
//
//	assert.Error(t, strutil.Validate(`-`, cms.KeyValidator))
//	assert.Error(t, strutil.Validate(`-a`, cms.KeyValidator))
//	assert.Error(t, strutil.Validate(`-_1`, cms.KeyValidator))
//	assert.Error(t, strutil.Validate(`123_-a`, cms.KeyValidator))
//	assert.Error(t, strutil.Validate(`中文`, cms.KeyValidator))
//	assert.Error(t, strutil.Validate(`.dice.id`, cms.KeyValidator))
//	assert.Error(t, strutil.Validate(`dice\.id`, cms.KeyValidator))
//	assert.Error(t, strutil.Validate(`dice\\id`, cms.KeyValidator))
//}
//
//func TestEnvTransferValidator(t *testing.T) {
//	assert.Error(t, strutil.Validate(`1`, cms.EnvTransferValidator))
//
//	assert.NoError(t, strutil.Validate(`a_aa1-2`, cms.EnvTransferValidator))
//}
