package cache

import (
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var mcpServers gcache.Cache

type McpServerInfo struct {
	Endpoint      string `json:"endpoint"`
	Version       string `json:"version,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Name          string `json:"name"`
	TransportType string `json:"transport_type"`
}

func init() {
	mcpServers = initCache()
}

func GetMcpServer(name string, tag string) (*McpServerInfo, error) {
	raw, err := mcpServers.Get(name)
	if err != nil {
		return nil, err
	}
	infos := raw.(map[string]*McpServerInfo)
	if len(infos) == 0 {
		return nil, fmt.Errorf("mcp server [%s] not found", name)
	}
	if infos[tag] == nil {
		return nil, fmt.Errorf("mcp server tag [%s] not found", tag)
	}
	return infos[tag], nil
}

func SetMcpServer(name string, tag string, info *McpServerInfo) error {
	if info == nil {
		return errors.New("info is nil")
	}
	var infos map[string]*McpServerInfo
	raw, err := mcpServers.Get(name)
	if err != nil {
		return err
	}
	if data, ok := raw.(map[string]*McpServerInfo); data == nil || !ok {
		infos = make(map[string]*McpServerInfo)
	} else {
		infos = data
	}

	infos[tag] = info

	logrus.Infof("mcp server [%s] set success, tag: [%s]", name, tag)
	return mcpServers.Set(name, infos)
}

func RemoveMcpServer(name string, tag string) error {
	raw, err := mcpServers.Get(name)
	if err != nil {
		return err
	}
	if infos := raw.(map[string]*McpServerInfo); infos != nil {
		delete(infos, tag)
		return mcpServers.Set(name, infos)
	}
	return nil
}

func ClearMcpServers() error {
	mcpServers = initCache()
	return nil
}

func initCache() gcache.Cache {
	var size = 1024
	raw := os.Getenv("MCP_SERVERS_CACHE_SIZE")
	if num, err := strconv.Atoi(raw); err == nil {
		size = num
	}

	return gcache.New(size).LoaderFunc(func(_ interface{}) (interface{}, error) {
		return make(map[string]*McpServerInfo), nil
	}).LRU().Build()
}
