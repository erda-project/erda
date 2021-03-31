package terminal

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/dbengine"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return strings.Contains(r.Header.Get("Origin"), os.Getenv("DICE_ROOT_DOMAIN"))
	},
}
var instanceinfoClient = instanceinfo.New(dbengine.MustOpen())

type ContainerInfo struct {
	Env  []string        `json:"env"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type ContainerInfoArg struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Container string `json:"container"`
}

func Terminal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("upgrade: %v", err)
		return
	}
	defer conn.Close()

	// 1. 首先获取前端发送过来所要连接的容器的信息
	t, message, err := conn.ReadMessage()
	if err != nil {
		logrus.Infof("failed to ReadMessage: %v", err)
		return
	}
	if t != websocket.TextMessage {
		return
	}
	containerinfo := ContainerInfo{}
	if err := json.Unmarshal(message, &containerinfo); err != nil {
		logrus.Errorf("failed to unmarshal containerinfo: %v, content: %s", err, string(message))
		return
	}
	if containerinfo.Name != "docker" {
		// 不是容器控制台, 作为 proxy 走 soldier
		SoldierTerminal(r, message, conn)
		return
	}
	var args ContainerInfoArg
	if err := json.Unmarshal(containerinfo.Args, &args); err != nil {
		logrus.Errorf("failed to unmarshal containerinfoArgs: %v", err)
		return
	}

	// 2. 在实例列表中查询该 containerid
	instances, err := instanceinfoClient.InstanceReader().ByContainerID(args.Container).Do()
	if err != nil {
		logrus.Errorf("failed to get instance by containerid: %v", err)
		return
	}

	if len(instances) == 0 {
		logrus.Errorf("no instances found: containerid: %v", args.Container)
		return
	}
	if len(instances) > 1 {
		logrus.Errorf("more than one instance found: containerid: %v", args.Container)
		return
	}
	instance := instances[0]

	// 3. 检查权限
	access := false
	if instance.OrgID != "" {
		orgid, err := strconv.ParseUint(instance.OrgID, 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse orgid for instance: %v, %v", instance.ContainerID, err)
			return
		}
		p, err := bundle.New(bundle.WithCMDB()).CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   r.Header.Get("User-ID"),
			Scope:    apistructs.OrgScope,
			ScopeID:  orgid,
			Resource: "terminal",
			Action:   "OPERATE",
		})
		if err != nil {
			logrus.Errorf("failed to check permissions for terminal: %v", err)
			return
		}
		if p.Access {
			access = true
		}
	}
	if !access && instance.ApplicationID != "" {
		appid, err := strconv.ParseUint(instance.ApplicationID, 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse applicationid for instance: %v, %v", instance.ContainerID, err)
			return
		}
		p, err := bundle.New(bundle.WithCMDB()).CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   r.Header.Get("User-ID"),
			Scope:    apistructs.AppScope,
			ScopeID:  appid,
			Resource: "terminal",
			Action:   "OPERATE",
		})
		if err != nil {
			logrus.Errorf("failed to check permissions for terminal: %v", err)
			return
		}
		if !p.Access {
			logrus.Errorf("permission denied for terminal, userid: %v, appid: %d", r.Header.Get("User-ID"), appid)
			return
		}
	}

	// 4. 判断是否为 dcos 路径
	k8snamespace, ok1 := instance.Metadata("k8snamespace")
	k8spodname, ok2 := instance.Metadata("k8spodname")
	k8scontainername, ok3 := instance.Metadata("k8scontainername")
	clustername := instance.Cluster
	if !ok1 || !ok2 || !ok3 {
		// 如果meta中没有相应 namespace, name, containername, 则认为是dcos路径, 走原来的 soldier
		SoldierTerminal(r, message, conn)
		return
	}

	K8STerminal(clustername, k8snamespace, k8spodname, k8scontainername, conn)
}

// SoldierTerminal proxy of soldier
func SoldierTerminal(r *http.Request, initmessage []byte, upperConn *websocket.Conn) {
	soldierAddr, err := url.Parse(r.URL.Query().Get("url"))
	if err != nil {
		logrus.Errorf("failed to url parse: %v, err: %v", r.URL.Query().Get("url"), err)
	}
	soldierAddr.Path = r.URL.Path
	switch soldierAddr.Scheme {
	case "https":
		soldierAddr.Scheme = "wss"
	case "http":
		soldierAddr.Scheme = "ws"
	default:
		soldierAddr.Scheme = "ws"
	}

	conn, _, err := websocket.DefaultDialer.Dial(soldierAddr.String(), nil)
	if err != nil {
		logrus.Errorf("failed to dial with %s: %v", soldierAddr, err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, initmessage); err != nil {
		logrus.Errorf("failed to write message: %v, err: %v", string(initmessage), err)
		return
	}
	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer func() {
			wait.Done()
			conn.Close()
			upperConn.Close()
		}()
		for {
			tp, m, err := upperConn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(tp, m); err != nil {
				return
			}
		}
	}()
	go func() {
		defer func() {
			wait.Done()
			conn.Close()
			upperConn.Close()
		}()
		for {
			tp, m, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := upperConn.WriteMessage(tp, m); err != nil {
				return
			}
		}
	}()
	wait.Wait()
}

func K8STerminal(clustername, namespace, podname, containername string, upperConn *websocket.Conn) {
	executorname := clusterutil.GenerateExecutorByClusterName(clustername)
	executor, err := executor.GetManager().Get(executortypes.Name(executorname))
	if err != nil {
		logrus.Errorf("failed to get executor by executorname(%s)", executorname)
		return
	}
	terminalExecutor, ok := executor.(executortypes.TerminalExecutor)
	if !ok {
		logrus.Errorf("executor(%s) not impl executortypes.TerminalExecutor", executorname)
		return
	}
	terminalExecutor.Terminal(namespace, podname, containername, upperConn)
}
