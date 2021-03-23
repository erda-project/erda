package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

var (
	s      *Sched
	action *spec.PipelineTask
)

func init() {
	s = &Sched{
		name:    "test",
		addr:    "scheduler.default.svc.cluster.local:9091",
		options: nil,
	}
	action = &spec.PipelineTask{
		Name:         "pipeline-local-test",
		Type:         "custom-script",
		ExecutorKind: "SCHEDULER",
		Status:       apistructs.PipelineStatusAnalyzed,
		Extra: spec.PipelineTaskExtra{
			Namespace:    "pipeline-test-namespace-1",
			ExecutorName: "scheduler",
			ClusterName:  "terminus-dev",
			PrivateEnvs: map[string]string{
				"AGENT_PRE_FETCHER_DEST_DIR": "/opt/emptydir",
				"WORKDIR":                    "/.pipeline/container/context/custom-script",
				"CONTEXTDIR":                 "/.pipeline/container/context",
				"METAFILE":                   "/.pipeline/container/metadata/custom-script/metadata",
				"DICE_OPENAPI_ADDR":          "openapi.default.svc.cluster.local:9529",
				"DICE_OPENAPI_TOKEN":         "c81037f8-b42f-498b-abd6-65fa867c1316",
			},
			Image:   "registry.cn-hangzhou.aliyuncs.com/dice/default-action-image:3.4.0-20190704-1765035",
			Cmd:     "/opt/emptydir/agent",
			CmdArgs: []string{"eyJjb21tYW5kcyI6WyJlY2hvIGhlbGxvIHBpcGVsaW5lISIsInNsZWVwIDFoIl0sImNvbnRleHQiOnsib3V0U3RvcmFnZXMiOlt7Im5hbWUiOiJjdXN0b20tc2NyaXB0IiwidmFsdWUiOiIvLnBpcGVsaW5lL2NvbnRleHQvY3VzdG9tLXNjcmlwdCIsInR5cGUiOiJkaWNlLW5mcy12b2x1bWUifV19LCJwaXBlbGluZUlEIjoxOTUsInBpcGVsaW5lVGFza0lEIjo3MzB9"},
			RuntimeResource: spec.RuntimeResource{
				CPU:    0.1,
				Memory: 32,
			},
			UUID: "uuid1",
			PreFetcher: &apistructs.PreFetcher{
				FileFromImage: "registry.cn-hangzhou.aliyuncs.com/dice/action-agent:3.7-20191022-fcc24a74bd",
				FileFromHost:  "/netdata/devops/ci/action-agent",
				ContainerPath: "/opt/emptydir",
			},
			Volumes: []apistructs.MetadataField{
				{
					Name:  "custom-script",
					Value: "/.pipeline/context/custom-script",
					Type:  "dice-nfs-volume",
				},
			},
		},
	}
}

func TestSched_Create(t *testing.T) {
	//job, err := s.Create(context.Background(), action)
	//require.NoError(t, err)
	//b, _ := json.MarshalIndent(job, "", "  ")
	//fmt.Println("created:\n" + string(b))

	job, err := s.Start(context.Background(), action)
	require.NoError(t, err)
	b, _ := json.MarshalIndent(job, "", "  ")
	fmt.Println("started:\n" + string(b))

	//data, err := s.Cancel(context.Background(), action)
	//require.NoError(t, err)
	//spew.Dump(data)
}

func TestSched_DeleteNamespace(t *testing.T) {
	fmt.Println(s.DeleteNamespace(context.Background(), action))
}

func TestSched_Status(t *testing.T) {
	fmt.Println(s.Status(context.Background(), action))
}
