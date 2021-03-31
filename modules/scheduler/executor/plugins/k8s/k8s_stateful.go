package k8s

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/strutil"
)

func groupStatefulset(sg *apistructs.ServiceGroup) ([]*apistructs.ServiceGroup, error) {
	group, ok := getGroupNum(sg)
	if !ok || group == "1" {
		return []*apistructs.ServiceGroup{sg}, nil
	}

	num, err := strutil.Atoi64(group)
	if err != nil {
		return nil, errors.Errorf("failed to parse addon group number, (%v)", err)
	}

	logrus.Infof("parsed multi addon group, name: %s, number: %v", sg.ID, num)

	// store 分组标识与该标识下的服务的映射
	// e.g. redis : [redis-1, redis-2]
	store := make(map[string][]apistructs.Service)
	// rstore 是 store 的反向存储，服务与分组标识的映射
	rstore := make(map[string]string)
	// 分组标识与分组标识之间的依赖，key 依赖 value
	groupDep := make(map[string][]string)

	for i := range sg.Services {
		key, ok := getGroupID(&sg.Services[i])
		if !ok {
			return nil, errors.Errorf("failed to get GROUP_ID, name: %s", sg.Services[i].Name)
		}

		store[key] = append(store[key], sg.Services[i])
		rstore[sg.Services[i].Name] = key
	}

	if int(num) != len(store) {
		return nil, errors.Errorf("addons group num error, specified num: %v, group num: %v", num, len(store))
	}

	var groups []*apistructs.ServiceGroup
	for k, v := range store {
		serviceGroup := &apistructs.ServiceGroup{
			ClusterName:   sg.ClusterName,
			Extra:         sg.Extra,
			ScheduleInfo:  sg.ScheduleInfo,
			ScheduleInfo2: sg.ScheduleInfo2,
		}

		serviceGroup.Dice.ID = k
		serviceGroup.Dice.Type = sg.Dice.Type
		serviceGroup.Dice.Services = v
		// 记录一个group下所有service名字, e.g. redis 这个groupID 下有 [redis-1, redis-2]
		groupSrv := map[string]bool{}
		for i := range v {
			groupSrv[v[i].Name] = true
		}

		// 修正 service 的 depends 字段
		// 该 group 下任一服务的依赖只能是该 group 下的服务
		// 服务跨 group 的依赖转换成 group 之间的依赖，即statefulset之间的创建顺序
		for i := range v {
			idx := 0
			for _, depend := range v[i].Depends {
				if _, ok := groupSrv[depend]; ok {
					v[i].Depends[idx] = depend
					idx++
				} else if !existedInSlice(groupDep[k], rstore[depend]) {
					groupDep[k] = append(groupDep[k], rstore[depend])
				}
			}
			v[i].Depends = v[i].Depends[:idx]
			logrus.Infof("service and its depends, services: %s, depends: %v", v[i].Name, v[i].Depends)
		}
		groups = append(groups, serviceGroup)
	}

	// 构建virtualGroup, 根据 groupDep 调整 groups 的顺序, 保证被依赖的 statefulset 先被创建
	virtualGroup := &apistructs.ServiceGroup{}
	for i := range groups {
		virtualGroup.Services = append(virtualGroup.Services, apistructs.Service{
			Name: groups[i].ID,
		})
	}

	for k, v := range groupDep {
		for i := range virtualGroup.Services {
			if k == virtualGroup.Services[i].Name {
				virtualGroup.Services[i].Depends = v
				break
			}
		}
	}
	layers, err := util.ParseServiceDependency(virtualGroup)
	if err != nil {
		return nil, errors.Errorf("failed to adjust group sequence, (%v)", err)
	}
	// sortedGroup 按 virtualGroup 排好的依赖顺序安置group的id
	sortedGroup := []string{}
	for i := range layers {
		for j := range layers[i] {
			sortedGroup = append(sortedGroup, layers[i][j].Name)
		}
	}
	//
	for i := range groups {
		if groups[i].ID == sortedGroup[i] {
			continue
		}
		for j := i + 1; j < len(groups); j++ {
			if groups[j].ID != sortedGroup[i] {
				continue
			}
			groups[i], groups[j] = groups[j], groups[i]
			break
		}
	}
	return groups, nil
}

// 初始化annotations, annotations目的是记录该编号对应的原始服务名称
// annotations 不仅要记录一个 statefulset 内的编号(NO, N1...),
// 还要记录分组的组号, 因为有的环境变量会跨组依赖
// 比如 1主1从3哨兵的 redis 的 runtime, 1主1从是一组(一个 statefulset 内),
// 3个 sentinel 是一组, 而 sentinel 里却有主和从的环境变量的依赖,
// 即 sentinel 里有诸如 ${REDIS_MASTER_HOST}, ${REDIS_SLAVE_PORT} 环境变量

// 格式: G0_N0
// globalSeq 是全局分组号，代表的是服务隶属于第几个组
// N0 是在一组中的顺序号
func initAnnotations(layers [][]*apistructs.Service, globalSeq int) map[string]string {
	order := 0
	annotations := map[string]string{}
	for _, layer := range layers {
		for j := range layer {
			// 记录该编号对应的原始服务名称
			// e.g. annotations["G0_N1"]="redis-slave"
			key := strutil.Concat("G", strconv.Itoa(globalSeq), "_N", strconv.Itoa(order))
			annotations[key] = layer[j].Name

			// e.g. annotations["redis-slave"]="G0_N1"
			annotations[layer[j].Name] = key

			// e.g. annotations["G0_ID"]="redis"
			// annotations[strutil.Concat("G", strconv.Itoa(globalSeq), "_ID")] = layer[j].Labels[groupID]
			annotations[strutil.Concat("G", strconv.Itoa(globalSeq), "_ID")], _ = getGroupID(layer[j])

			// 记录各个服务的PORT
			// redis-slave -> redis_slave -> REDIS_SLAVE_PORT
			if len(layer[j].Ports) > 0 {
				name := strings.Replace(layer[j].Name, "-", "_", -1)
				annotations[strutil.Concat(strings.ToUpper(name), "_PORT")] = strconv.Itoa(layer[j].Ports[0].Port)
			}

			if j == len(layer)-1 {
				break
			}
			order++
		}
		order++
	}
	return annotations
}

// 1，确定各个服务的编号，从0开始
// 2，搜集各个服务的环境变量
func (k *Kubernetes) initGroupEnv(layers [][]*apistructs.Service, annotations map[string]string) map[string]string {
	order := 0
	allEnv := make(map[string]string)

	for _, layer := range layers {
		for j := range layer {
			for k, v := range layer[j].Env {
				globalKey := strutil.Concat("N", strconv.Itoa(order), "_", k)
				if str, ok := parseSpecificEnv(v, annotations); ok {
					v = str
				}
				allEnv[globalKey] = v
			}

			ciEnvs, err := k.ClusterInfo.Get()
			if err != nil {
				logrus.Error(err)
			} else {
				clusterName, ok := ciEnvs[clusterinfo.DiceClusterName]
				if ok {
					globalKey := strutil.Concat("N", strconv.Itoa(order), "_", clusterinfo.DiceClusterName)
					allEnv[globalKey] = clusterName
				}
			}

			if j == len(layer)-1 {
				break
			}
			order++
		}
		order++
	}
	return allEnv
}

// 将中间件中带变量的环境变量解析出来，如TERMINUS_ZOOKEEPER_1_HOST=${terminus-zookeeper-1}
func parseSpecificEnv(val string, annotations map[string]string) (string, bool) {
	results := envReg.FindAllString(val, -1)
	if len(results) == 0 {
		return "", false
	}
	replace := make(map[string]string)

	logrus.Infof("in parsing specific env: %s, annotations: %+v", val, annotations)

	for _, str := range results {
		if len(str) <= 3 {
			continue
		}
		// e.g. ${REDIS_HOST} -> REDIS_HOST
		key := str[2 : len(str)-1]
		// 目前只支持在变量中解析 _HOST, _PORT 类型的变量
		if strings.Contains(key, "_HOST") {
			pos := strings.LastIndex(key, "_")
			name := strings.TrimSuffix(key[:pos], "_HOST")
			name = toServiceName(name)

			// seq 值 "G0_N1" 代表第0分组(statefulset), 在该分组里的序号是1
			bigSeq, ok := annotations[name]
			if !ok {
				logrus.Errorf("failed to parse env as not found in annotations,"+
					" var: %s, name: %s, annotations: %+v", key, name, annotations)
				break
			}
			seqs := strings.Split(bigSeq, "_")
			if len(seqs) != 2 {
				logrus.Errorf("failed to parse env seq, key: %s, bigSeq: %s", key, bigSeq)
				break
			}
			// "N1" -> "1"
			seq := seqs[1][1:]
			// e.g. G0_ID, G1_ID, 组号标识
			id, ok := annotations[strutil.Concat(seqs[0], "_ID")]
			if !ok {
				logrus.Errorf("failed to get group id from annotations, key: %s, groupseq: %s", key, seqs[0])
			}

			bracedKey := strutil.Concat("${", key, "}")

			ns, ok := annotations["K8S_NAMESPACE"]
			if ok {
				replace[bracedKey] = strutil.Concat(id, "-", seq, ".", id, ".", ns, ".svc.cluster.local")
			} else {
				// e.g. 用户设置的 id 为 web, 在statefulset中的实例序列号为1，则该pod的短域名为web-1.web
				replace[bracedKey] = strutil.Concat(id, "-", seq, ".", id)
			}

		} else if strings.Contains(key, "_PORT") {
			port, ok := annotations[key]
			if !ok {
				logrus.Errorf("failed to parse env as not found in annotations, key: %s, annotations: %+v", key, annotations)
				break
			}
			bracedKey := strutil.Concat("${", key, "}")
			replace[bracedKey] = port
		}
	}

	if len(replace) == 0 {
		logrus.Infof("debug parseSpecificEnv empty replace")
		return "", false
	}

	before := val
	for k, v := range replace {
		val = strings.Replace(val, k, v, 1)
	}
	logrus.Infof("succeed to convert env, before: %s, after: %s, replace: %+v", before, val, replace)
	return val, true
}

// "TERMINUS_ZOOKEEPER_1" -> "terminus-zookeeper-1"
func toServiceName(origin string) string {
	return strings.Replace(strings.ToLower(origin), "_", "-", -1)
}

// 创建 statefulset 的 service, statefulset下的各个实例都有相应的 dns 域名
// 每个实例的域名规则：{podName}.{serviceName}.{namespace}.svc.cluster.local
// 暂不使用 headless service
func (k *Kubernetes) createStatefulService(sg *apistructs.ServiceGroup) error {
	if len(sg.Services[0].Ports) == 0 {
		return nil
	}
	// TODO: 和无状态 service 区分开
	// 构建一个 statefulset 的 service
	svc := sg.Services[0]

	newService(&svc)
	svc.Name = statefulsetName(sg)
	k8sSvc := newService(&svc)

	if err := k.service.Create(k8sSvc); err != nil {
		return err
	}
	v, ok := sg.Services[0].Labels["HAPROXY_0_VHOST"]
	// 无外部域名
	if !ok {
		return nil
	}
	// 将label中HAPROXY_0_VHOST对应的域名/vip集合都转发到该服务的第0个端口上
	publicHosts := strings.Split(v, ",")
	if len(publicHosts) == 0 {
		return nil
	}
	// 创建ingress
	rules := buildRules(publicHosts, svc.Name, sg.Services[0].Ports[0].Port)
	tls := buildTLS(publicHosts)
	ingress := &extensionsv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: rules,
			TLS:   tls,
		},
	}

	return k.ingress.Create(ingress)
}

// TODO: 状态要更精确
func (k *Kubernetes) GetStatefulStatus(sg *apistructs.ServiceGroup) (apistructs.StatusDesc, error) {
	var status apistructs.StatusDesc
	namespace := MakeNamespace(sg)
	statefulName := sg.Services[0].Name
	idx := strings.LastIndex(statefulName, "-")
	if idx > 0 {
		statefulName = statefulName[:idx]
	}

	logrus.Infof("in getStatefulStatus, name: %s, namespace: %s", statefulName, namespace)

	// 只有一个 statefulset
	if !strings.HasPrefix(namespace, "group-") {
		return k.getOneStatus(namespace, statefulName)
	}
	// 有多个 statefulset，需要组合状态
	groups, err := groupStatefulset(sg)
	if err != nil {
		return status, err
	}
	for i := range groups {
		//name := groups[i].Services[0].Name
		name := groups[i].ID
		oneStatus, err := k.getOneStatus(namespace, name)
		if err != nil {
			return status, err
		}
		if oneStatus.Status == apistructs.StatusProgressing {
			return oneStatus, nil
		}
	}
	status.Status = apistructs.StatusReady
	return status, nil
}

func (k *Kubernetes) getOneStatus(namespace, name string) (apistructs.StatusDesc, error) {
	logrus.Infof("in getOneStatus, name: %s, namespace: %s", name, namespace)
	var status apistructs.StatusDesc
	set, err := k.sts.Get(namespace, name)
	if err != nil {
		if err == k8serror.ErrNotFound {
			status.Status = apistructs.StatusProgressing
			status.LastMessage = "currently could not get the pod"
			return status, nil
		}
		return status, err
	}
	var replica int32 = 1
	if set.Spec.Replicas != nil {
		replica = *set.Spec.Replicas
	}
	if replica == set.Status.Replicas &&
		replica == set.Status.ReadyReplicas &&
		replica == set.Status.UpdatedReplicas {
		status.Status = apistructs.StatusReady
	} else {
		status.Status = apistructs.StatusProgressing
	}

	msgList, err := k.event.AnalyzePodEvents(namespace, name)
	if err != nil {
		logrus.Errorf("failed to analyze k8s events, namespace: %s, name: %s, (%v)",
			namespace, name, err)
	}
	if len(msgList) > 0 {
		status.LastMessage = msgList[len(msgList)-1].Comment
	}

	return status, nil
}

func (k *Kubernetes) inspectOne(g *apistructs.ServiceGroup, namespace, name string, groupNum int) (*OneGroupInfo, error) {
	set, err := k.sts.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	sg := &apistructs.ServiceGroup{
		ClusterName:  g.ClusterName,
		Executor:     g.Executor,
		ScheduleInfo: g.ScheduleInfo,
		Dice: apistructs.Dice{
			ID:   set.Annotations["RUNTIME_NAME"],
			Type: set.Annotations["RUNTIME_NAMESPACE"],
		},
	}
	replica := *(set.Spec.Replicas)
	container := &set.Spec.Template.Spec.Containers[0]

	envs := make(map[string]string)
	for _, env := range container.Env {
		envs[env.Name] = env.Value
	}

	for i := 0; i < int(replica); i++ {
		podName := strutil.Concat(container.Name, "-", strconv.Itoa(i))
		pod, err := k.pod.Get(namespace, podName)
		if err != nil && err != k8serror.ErrNotFound {
			return nil, err
		}
		// 从存储在statefulset中的anno
		key := strutil.Concat("G", strconv.Itoa(groupNum), "_N", strconv.Itoa(i))
		serviceName, ok := set.Annotations[key]
		if !ok {
			return nil, errors.Errorf("failed to get key from annotations, podname: %s, key: %s, annotations: %v",
				podName, key, set.Annotations)
		}

		podStatus := apistructs.StatusProgressing
		podIP := ""
		if err == nil {
			podStatus = convertStatus(pod.Status)
			podIP = pod.Status.PodIP
		}

		msgList, err := k.event.AnalyzePodEvents(namespace, podName)
		if err != nil {
			logrus.Errorf("failed to analyze job events, namespace: %s, name: %s, (%v)",
				namespace, name, err)
		}

		var lastMsg string
		if len(msgList) > 0 {
			lastMsg = msgList[len(msgList)-1].Comment
		}

		sg.Services = append(sg.Services, apistructs.Service{
			Name:     serviceName,
			Vip:      strutil.Concat(name, ".", namespace, ".svc.cluster.local"),
			ShortVIP: name,
			Env:      envs,
			StatusDesc: apistructs.StatusDesc{
				Status:      podStatus,
				LastMessage: lastMsg,
			},
			Image:         container.Image,
			InstanceInfos: []apistructs.InstanceInfo{{Ip: podIP}},
		})
	}

	sg.Status = apistructs.StatusReady
	for i := range sg.Services {
		if sg.Services[i].Status != apistructs.StatusReady {
			sg.Status = apistructs.StatusProgressing
			break
		}
	}
	groupInfo := &OneGroupInfo{
		sg:  sg,
		sts: set,
	}
	return groupInfo, nil
}

// inspect 多个 statefulset
func (k *Kubernetes) inspectGroup(g *apistructs.ServiceGroup, namespace, name string) (*apistructs.ServiceGroup, error) {
	mygroups, err := groupStatefulset(g)
	if err != nil {
		logrus.Errorf("failed to get groups sequence in inspectgroup, namespace: %s, name: %s", g.Type, g.ID)
		return nil, err
	}

	var groupsInfo []*OneGroupInfo
	for i, group := range mygroups {
		// 先找到groupNum
		//for k, v := range
		oneGroup, err := k.inspectOne(g, namespace, group.ID, i)
		if err != nil {
			return nil, err
		}
		groupsInfo = append(groupsInfo, oneGroup)
	}
	logrus.Infof("debug: inspectGroup, groupsInfo: %+v, Extra: %+v", groupsInfo, g.Extra)

	if len(groupsInfo) <= 1 {
		return nil, errors.Errorf("failed to parse multi group, namespace: %s, name: %s", namespace, name)
	}
	isReady := true
	for i := range g.Services {
		for j := range groupsInfo {
			for k := range groupsInfo[j].sg.Services {
				if groupsInfo[j].sg.Services[k].Name != g.Services[i].Name {
					continue
				}
				g.Services[i] = groupsInfo[j].sg.Services[k]
				if g.Services[i].Status != apistructs.StatusReady {
					isReady = false
				}
			}
		}
	}
	if isReady {
		g.Status = apistructs.StatusReady
	}

	return g, nil
}

func existedInSlice(array []string, elem string) bool {
	for _, x := range array {
		if x == elem {
			return true
		}
	}
	return false
}

// todo: 兼容老的标识
func getGroupNum(sg *apistructs.ServiceGroup) (string, bool) {
	if group, ok := sg.Labels[groupNum]; ok {
		return group, ok
	}
	group, ok := sg.Labels[groupNum2]
	return group, ok
}

func getGroupID(svc *apistructs.Service) (string, bool) {
	if id, ok := svc.Labels[groupID]; ok {
		return id, ok
	}
	id, ok := svc.Labels[groupID2]
	return id, ok
}
