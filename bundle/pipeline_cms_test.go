package bundle

//import (
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestBundle_PipelineCms(t *testing.T) {
//	os.Setenv("PIPELINE_ADDR", "localhost:3081")
//	b := New(WithPipeline())
//
//	ns := "bundle-test-ns"
//	source := apistructs.PipelineSourceDefault
//
//	err := b.CreatePipelineCmsNs(apistructs.PipelineCmsCreateNsRequest{
//		PipelineSource: source,
//		NS:             ns,
//	})
//	assert.NoError(t, err)
//
//	err = b.CreateOrUpdatePipelineCmsNsConfigs(ns, apistructs.PipelineCmsUpdateConfigsRequest{
//		PipelineSource: source,
//		KVs: map[string]apistructs.PipelineCmsConfigValue{
//			"testName": {
//				Value:       "a",
//				EncryptInDB: true,
//			},
//		},
//	})
//	assert.NoError(t, err)
//
//	kvs, err := b.GetPipelineCmsNsConfigs(ns, apistructs.PipelineCmsGetConfigsRequest{
//		PipelineSource: source,
//		Keys: []apistructs.PipelineCmsConfigKey{
//			{
//				Key:     "testName",
//				Decrypt: true,
//			},
//		},
//	})
//	assert.NoError(t, err)
//	fmt.Printf("%+v\n", kvs)
//}
//
//func TestBundle_PipelineCmsGet(t *testing.T) {
//	os.Setenv("PIPELINE_ADDR", "localhost:3081")
//	b := New(WithPipeline())
//
//	ns := "pipeline-secrets-app-8-default"
//	source := apistructs.PipelineSourceDice
//
//	configs, err := b.GetPipelineCmsNsConfigs(ns, apistructs.PipelineCmsGetConfigsRequest{
//		PipelineSource: source,
//		Keys: []apistructs.PipelineCmsConfigKey{
//			{
//				Key:     "ACTION_VERSION",
//				Decrypt: false,
//			},
//			{
//				Key:     "IS_FOR_MOBIL",
//				Decrypt: true,
//			},
//		},
//		GlobalDecrypt: true,
//	})
//	assert.NoError(t, err)
//	for _, config := range configs {
//		fmt.Printf("key: %s, value: %s\n", config.Key, config.Value)
//	}
//}
//
//func TestBundle_PipelineCmsUpdateDiceFiles(t *testing.T) {
//	os.Setenv("PIPELINE_ADDR", "localhost:3081")
//	b := New(WithPipeline())
//
//	ns := "dice-files-ns"
//	source := apistructs.PipelineSourceDefault
//
//	err = b.CreateOrUpdatePipelineCmsNsConfigs(ns, apistructs.PipelineCmsUpdateConfigsRequest{
//		PipelineSource: source,
//		KVs: map[string]apistructs.PipelineCmsConfigValue{
//			"a.cert": { // pushed from cert manage
//				Value:       "uuid-1111",
//				EncryptInDB: false,
//				Type:        apistructs.PipelineCmsConfigTypeDiceFile,
//				Operations: &apistructs.PipelineCmsConfigOperations{
//					CanDownload: true,
//					CanEdit:     false,
//					CanDelete:   false,
//				},
//			},
//		},
//	})
//	assert.NoError(t, err)
//}
