package cmdb

import (
	"context"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/erda-project/erda/modules/cmdb/conf"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/types"

	"bou.ke/monkey"
)

const testClusterName = "cmdb-test"

var (
	hostIps        = randomIP(10)
	header         = map[string]string{"Accept": "application/vnd.dice+json;version=1.0", "Previleged": "true"}
	diceContainers = ProduceServiceContainers(10)
)

// func init() {
// 	PatchParseConf()
// 	PatchRunConsumer()
// 	PatchNewServer()
// 	defer monkey.UnpatchAll()
// 	go func() {
// 		err := Initialize()
// 		if err != nil {
// 			panic("start cmdb server failed")
// 		}
// 	}()
// 	time.Sleep(5 * time.Second)
// }
//
// func initEndpoint(t *testing.T) *httpexpect.Expect {
// 	endpoint := httpexpect.New(t, "http://127.0.0.1:9093/api")
// 	return endpoint
// }

func randomIP(size int) []string {
	var result []string

	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= size; i++ {
		a := strconv.Itoa(rand.Intn(254))
		b := strconv.Itoa(rand.Intn(254))
		c := strconv.Itoa(rand.Intn(254))
		d := strconv.Itoa(rand.Intn(254))
		ip := a + "." + b + "." + c + "." + d
		result = append(result, ip)
	}
	return result
}

func RandStringBytes(n, size int) []string {
	var result []string
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= size; i++ {
		b := make([]byte, n)
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		result = append(result, string(b))
	}
	return result
}

func ProduceHosts() *[]types.CmHost {
	var hosts []types.CmHost
	for _, ip := range hostIps {
		host := &types.CmHost{
			Cluster:     testClusterName,
			PrivateAddr: ip,
			Cpus:        8,
			Memory:      34359738368,
			Disk:        107374182400,
		}
		hosts = append(hosts, *host)
	}
	return &hosts
}

func ProduceStandardContainers(number int) *[]types.CmContainer {
	var containers []types.CmContainer
	for _, ip := range hostIps {
		containerIps := randomIP(number)
		containerID := RandStringBytes(10, number)
		for i := 0; i < number; i++ {
			container := &types.CmContainer{
				ID:                containerID[i],
				HostPrivateIPAddr: ip,
				CPU:               0.04,
				Memory:            1073741824,
				Disk:              536870912,
				Cluster:           testClusterName,
				IPAddress:         containerIps[i],
			}
			containers = append(containers, *container)
		}
	}
	return &containers
}

func ProduceComponentContainers(number int) *[]types.CmContainer {
	var containers []types.CmContainer
	for _, ip := range hostIps {
		containerIps := randomIP(number)
		containerID := RandStringBytes(10, number)
		componentID := RandStringBytes(5, number)
		for i := 0; i < number; i++ {
			container := &types.CmContainer{
				ID:                containerID[i],
				HostPrivateIPAddr: ip,
				CPU:               0.04,
				Memory:            1073741824,
				Disk:              536870912,
				Cluster:           testClusterName,
				IPAddress:         containerIps[i],
				DiceComponent:     componentID[i],
			}
			containers = append(containers, *container)
		}
	}
	return &containers
}

func ProduceAddonContainer(number int) *[]types.CmContainer {
	projects := []string{"10", "11", "12"}
	workspaces := []string{"development", "test", "staging", "production"}
	sharedLevels := []string{"", "PROJECT", "ORG"}
	var addons []types.CmContainer
	for _, ip := range hostIps {
		addonContainerIps := randomIP(number)
		addonContainerID := RandStringBytes(10, number)
		addonsID := RandStringBytes(10, number)
		addonsName := RandStringBytes(5, number)
		for i := 0; i < number; i++ {
			rand.Seed(time.Now().UnixNano())
			add := &types.CmContainer{
				ID:                addonContainerID[i],
				HostPrivateIPAddr: ip,
				CPU:               0.04,
				Memory:            1073741824,
				Disk:              536870912,
				Cluster:           testClusterName,
				DiceAddon:         addonsID[i],
				DiceAddonName:     addonsName[i],
				DiceProject:       projects[rand.Intn(3)],
				DiceSharedLevel:   sharedLevels[rand.Intn(3)],
				IPAddress:         addonContainerIps[i],
				Deleted:           false,
				TimeStamp:         time.Now().UnixNano(),
			}
			if add.DiceSharedLevel == "PROJECT" {
				add.DiceWorkspace = workspaces[rand.Intn(3)]
			}
			addons = append(addons, *add)
		}
	}
	return &addons
}

func ProduceServiceContainers(number int) *[]types.CmContainer {
	var containers []types.CmContainer
	projectID := RandStringBytes(5, number)
	for _, ip := range hostIps {
		containerIps := randomIP(number)
		containerID := RandStringBytes(10, number)
		applicationID := RandStringBytes(5, number)
		runtimeID := RandStringBytes(5, number)
		serviceID := RandStringBytes(5, number)
		for i := 0; i < number; i++ {
			container := &types.CmContainer{
				ID:                containerID[i],
				HostPrivateIPAddr: ip,
				CPU:               0.04,
				Memory:            1073741824,
				Disk:              536870912,
				Cluster:           testClusterName,
				IPAddress:         containerIps[i],
				DiceProject:       projectID[i],
				DiceApplication:   applicationID[i],
				DiceRuntime:       runtimeID[i],
				DiceService:       serviceID[i],
			}
			containers = append(containers, *container)
		}
	}
	return &containers
}

func PatchParseConf() {
	monkey.Patch(conf.Load, func() error {
		return nil
	})
}

// func PatchNewServer() {
// 	monkey.Patch(NewServer, httpserver.New(":9093"))
// }

// func PatchRunConsumer() {
// 	var server *endpoints.Server
// 	monkey.PatchInstanceMethod(reflect.TypeOf(server), "RunConsumer",
// 		func(server *endpoints.Server) {})
// }

func PatchAllHostsByCluster() {
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllHostsByCluster",
		func(db *dao.DBClient, ctx context.Context, cluster string) (*[]types.CmHost, error) {
			return ProduceHosts(), nil
		})
}

func PatchAllContainersByHost() {
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllContainersByHost",
		func(db *dao.DBClient, ctx context.Context, cluster string, host []string) (*[]types.CmContainer, error) {
			return ProduceStandardContainers(10), nil
		})
}

func PatchOneContainerByHost() {
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllContainersByHost",
		func(db *dao.DBClient, ctx context.Context, cluster string, host []string) (*[]types.CmContainer, error) {
			return ProduceStandardContainers(1), nil
		})
}

func PatchAllComponentsByCluster() {
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllComponentsByCluster",
		func(db *dao.DBClient, ctx context.Context, cluster string) (*[]types.CmContainer, error) {
			return ProduceComponentContainers(1), nil
		})
}

func PatchAllContainersByAddon() {
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllAddonsByCluster",
		func(db *dao.DBClient, ctx context.Context, cluster string) (*[]types.CmContainer, error) {
			return ProduceAddonContainer(10), nil
		})
}

func PatchAllProjectsContainersByCluster() {
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllProjectsContainersByCluster",
		func(db *dao.DBClient, ctx context.Context, cluster string) (*[]types.CmContainer, error) {
			return diceContainers, nil
		})
}

func PatchAllContainersByProject(project string) {
	var projectContainers []types.CmContainer
	for _, container := range *diceContainers {
		if container.DiceProject == project {
			projectContainers = append(projectContainers, container)
		}
	}
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllContainersByProject",
		func(db *dao.DBClient, ctx context.Context, project []string) (*[]types.CmContainer, error) {
			return &projectContainers, nil
		})
}

func PatchAllContainersByApplication(app string) {
	var appContainers []types.CmContainer
	for _, container := range *diceContainers {
		if container.DiceApplication == app {
			appContainers = append(appContainers, container)
		}
	}
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllContainersByApplication",
		func(db *dao.DBClient, ctx context.Context, app []string) (*[]types.CmContainer, error) {
			return &appContainers, nil
		})
}

func PatchAllContainersByRuntime(runtime string) {
	var serviceContainers []types.CmContainer
	for _, container := range *diceContainers {
		if container.DiceRuntime == runtime {
			serviceContainers = append(serviceContainers, container)
		}
	}
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllContainersByRuntime",
		func(db *dao.DBClient, ctx context.Context, runtime []string) (*[]types.CmContainer, error) {
			return &serviceContainers, nil
		})
}

func PatchAllContainersByService(service string) {
	var serviceContainers []types.CmContainer
	for _, container := range *diceContainers {
		if container.DiceRuntime == service {
			serviceContainers = append(serviceContainers, container)
		}
	}
	var db *dao.DBClient
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "AllContainersByService",
		func(db *dao.DBClient, ctx context.Context, runtime string, service []string) (*[]types.CmContainer, error) {
			return &serviceContainers, nil
		})
}

// func TestClusterUsage(t *testing.T) {
// 	PatchAllContainersByHost()
// 	PatchAllHostsByCluster()
// 	defer monkey.UnpatchAll()
// 	repos := initEndpoint(t).GET("/clusters/cmdb-test/usage").
// 		WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 	repos.Object().Value("success").Boolean().Equal(true)
// 	result := repos.Object().Value("data").Raw()
// 	var usage *apistructs.GetClusterUsageResponseData
// 	bytes, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	err = json.Unmarshal(bytes, &usage)
// 	assert.Nil(t, err)
// 	// check total resource
// 	assert.Equal(t, usage.TotalCPU, float64(80))
// 	assert.Equal(t, usage.TotalMemory, float64(320))
// 	assert.Equal(t, usage.TotalDisk, float64(1000))
// 	// check used resource
// 	assert.Equal(t, usage.UsedCPU, float64(4))
// 	assert.Equal(t, usage.UsedMemory, float64(100))
// 	assert.Equal(t, usage.UsedDisk, float64(50))
// }

// func TestHostsUsage(t *testing.T) {
// 	PatchAllHostsByCluster()
// 	PatchOneContainerByHost()
// 	defer monkey.UnpatchAll()
// 	repos := initEndpoint(t).GET("/clusters/cmdb-test/hosts-usage").
// 		WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 	repos.Object().Value("success").Boolean().Equal(true)
// 	result := repos.Object().Value("data").Raw()
// 	var usages []*apistructs.GetHostUsageResponseData
// 	bytes, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	err = json.Unmarshal(bytes, &usages)
// 	assert.Nil(t, err)
// 	// check amount
// 	assert.Equal(t, len(usages), 10)
// 	// check resource
// 	for _, usage := range usages {
// 		assert.Equal(t, usage.TotalCPU, float64(8))
// 		assert.Equal(t, usage.TotalMemory, float64(32))
// 		assert.Equal(t, usage.TotalDisk, float64(100))
// 		assert.Equal(t, usage.UsedCPU, float64(0.4))
// 		assert.Equal(t, usage.UsedMemory, float64(10))
// 		assert.Equal(t, usage.UsedDisk, float64(5))
// 	}
// }

// func TestComponentUsage(t *testing.T) {
// 	PatchAllComponentsByCluster()
// 	defer monkey.UnpatchAll()
// 	repos := initEndpoint(t).GET("/clusters/cmdb-test/instances-usage").
// 		WithQuery("type", "component").
// 		WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 	repos.Object().Value("success").Boolean().Equal(true)
// 	result := repos.Object().Value("data").Raw()
// 	var usages []*apistructs.GetComponentUsageResponseData
// 	bytes, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	err = json.Unmarshal(bytes, &usages)
// 	assert.Nil(t, err)
// 	// check total number
// 	var amount int
// 	for _, usage := range usages {
// 		amount += usage.Instance
// 		assert.Equal(t, usage.Memory, float64(1024))
// 		assert.Equal(t, usage.CPU, float64(0.04))
// 		assert.Equal(t, usage.Disk, float64(512))
// 	}
// 	assert.Equal(t, amount, 10)
// }

// func TestAddonUsage(t *testing.T) {
// 	PatchAllContainersByAddon()
// 	defer monkey.UnpatchAll()
// 	repos := initEndpoint(t).GET("/clusters/cmdb-test/instances-usage").
// 		WithQuery("type", "addon").
// 		WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 	repos.Object().Value("success").Boolean().Equal(true)
// 	result := repos.Object().Value("data").Raw()
// 	var usages []*apistructs.GetAddOnUsageResponseData
// 	bytes, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	err = json.Unmarshal(bytes, &usages)
// 	assert.Nil(t, err)
// 	// check total number
// 	var amount int
// 	for _, usage := range usages {
// 		amount += usage.Instance
// 		assert.Equal(t, usage.Memory, float64(1024))
// 		assert.Equal(t, usage.CPU, float64(0.04))
// 		assert.Equal(t, usage.Disk, float64(512))
// 	}
// 	assert.Equal(t, amount, 100)
// }

// func TestProjectUsage(t *testing.T) {
// 	projectID := (*diceContainers)[0].DiceProject
// 	PatchAllContainersByProject(projectID)
//
// 	applicationID := (*diceContainers)[0].DiceApplication
// 	PatchAllContainersByApplication(applicationID)
//
// 	runtimeID := (*diceContainers)[0].DiceRuntime
// 	PatchAllContainersByRuntime(runtimeID)
//
// 	serviceName := (*diceContainers)[0].DiceService
// 	PatchAllContainersByService(serviceName)
//
// 	PatchAllProjectsContainersByCluster()
// 	defer monkey.UnpatchAll()
// 	repos := initEndpoint(t).GET("/clusters/cmdb-test/instances-usage").
// 		WithQuery("type", "project").
// 		WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 	repos.Object().Value("success").Boolean().Equal(true)
// 	result := repos.Object().Value("data").Raw()
// 	var usages []*apistructs.GetProjectUsageResponseData
// 	bytes, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	err = json.Unmarshal(bytes, &usages)
// 	assert.Nil(t, err)
// 	// check total number
// 	var amount int
// 	for _, usage := range usages {
// 		amount += usage.Instance
// 		assert.Equal(t, usage.Memory, float64(10240))
// 		assert.Equal(t, usage.CPU, float64(0.4))
// 		assert.Equal(t, usage.Disk, float64(5120))
// 	}
// 	assert.Equal(t, amount, 100)
//
// 	{
// 		repos := initEndpoint(t).GET("/clusters/cmdb-test/instances-usage").
// 			WithQuery("type", "application").
// 			WithQuery("project", projectID).
// 			WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 		repos.Object().Value("success").Boolean().Equal(true)
// 		result := repos.Object().Value("data").Raw()
// 		var usages []*apistructs.GetApplicationUsageResponseData
// 		bytes, err := json.Marshal(result)
// 		assert.Nil(t, err)
// 		err = json.Unmarshal(bytes, &usages)
// 		assert.Nil(t, err)
// 		// check total number
// 		var amount int
// 		for _, usage := range usages {
// 			amount += usage.Instance
// 			assert.Equal(t, usage.Memory, float64(1024))
// 			assert.Equal(t, usage.CPU, float64(0.04))
// 			assert.Equal(t, usage.Disk, float64(512))
// 		}
// 		assert.Equal(t, amount, 10)
// 	}
//
// 	{
// 		repos := initEndpoint(t).GET("/clusters/cmdb-test/instances-usage").
// 			WithQuery("type", "runtime").
// 			WithQuery("project", projectID).
// 			WithQuery("application", applicationID).
// 			WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 		repos.Object().Value("success").Boolean().Equal(true)
// 		result := repos.Object().Value("data").Raw()
// 		var usages []*apistructs.GetRuntimeUsageResponseData
// 		bytes, err := json.Marshal(result)
// 		assert.Nil(t, err)
// 		err = json.Unmarshal(bytes, &usages)
// 		assert.Nil(t, err)
// 		// check total number
// 		var amount int
// 		for _, usage := range usages {
// 			amount += usage.Instance
// 			assert.Equal(t, usage.Memory, float64(1024))
// 			assert.Equal(t, usage.CPU, float64(0.04))
// 			assert.Equal(t, usage.Disk, float64(512))
// 		}
// 		assert.Equal(t, amount, 1)
// 	}
//
// 	{
// 		repos := initEndpoint(t).GET("/clusters/cmdb-test/instances-usage").
// 			WithQuery("type", "service").
// 			WithQuery("project", projectID).
// 			WithQuery("application", applicationID).
// 			WithQuery("runtime", runtimeID).
// 			WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 		repos.Object().Value("success").Boolean().Equal(true)
// 		result := repos.Object().Value("data").Raw()
// 		var usages []*apistructs.GetServiceUsageResponseData
// 		bytes, err := json.Marshal(result)
// 		assert.Nil(t, err)
// 		err = json.Unmarshal(bytes, &usages)
// 		assert.Nil(t, err)
// 		// check total number
// 		var amount int
// 		for _, usage := range usages {
// 			amount += usage.Instance
// 			assert.Equal(t, usage.Memory, float64(1024))
// 			assert.Equal(t, usage.CPU, float64(0.04))
// 			assert.Equal(t, usage.Disk, float64(512))
// 		}
// 		assert.Equal(t, amount, 1)
// 	}
//
// 	{
// 		repos := initEndpoint(t).GET("/clusters/cmdb-test/instances").
// 			WithQuery("type", "service").
// 			WithQuery("project", projectID).
// 			WithQuery("application", applicationID).
// 			WithQuery("runtime", runtimeID).
// 			WithQuery("service", serviceName).
// 			WithHeaders(header).Expect().Status(http.StatusOK).JSON()
// 		repos.Object().Value("success").Boolean().Equal(true)
// 		result := repos.Object().Value("data").Raw()
// 		var containers []*types.CmContainer
// 		bytes, err := json.Marshal(result)
// 		assert.Nil(t, err)
// 		err = json.Unmarshal(bytes, &containers)
// 		assert.Nil(t, err)
// 		// check total number
// 		for _, container := range containers {
// 			assert.Equal(t, container.Memory, float64(1024))
// 			assert.Equal(t, container.CPU, float64(0.04))
// 			assert.Equal(t, container.Disk, float64(512))
// 			assert.Equal(t, container.DiceProject, projectID)
// 			assert.Equal(t, container.DiceApplication, applicationID)
// 			assert.Equal(t, container.DiceRuntime, runtimeID)
// 			assert.Equal(t, container.DiceService, serviceName)
// 		}
// 	}
// }
