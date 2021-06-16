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

package nacos

import (
	"net"
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/msp/registercenter/pb"
)

// ServiceSearchResult .
type ServiceSearchResult struct {
	ServiceName   string                        `json:"serviceName"`
	GroupName     string                        `json:"groupName"`
	ClusterMap    map[string]*ServiceHosts      `json:"clusterMap"`
	OwnerNameList []string                      `json:"ownerNameList"`
	IPList        []string                      `json:"ipList"`
	OwnerMap      map[string]*pb.InterfaceOwner `json:"ownerMap"`
}

// ToHTTPService .
func (s *ServiceSearchResult) ToHTTPService() *pb.HTTPService {
	var service pb.HTTPService
	service.ServiceName = s.ServiceName
	if len(s.ClusterMap) <= 0 || s.ClusterMap["DEFAULT"] == nil || len(s.ClusterMap["DEFAULT"].Hosts) <= 0 {
		return &service
	}
	def := s.ClusterMap["DEFAULT"]
	service.ServiceDomain = def.Hosts[0].getDomain()
	for _, item := range def.Hosts {
		service.HttpServiceDto = append(service.HttpServiceDto, item.ToHTTPServiceItem())
	}
	return &service
}

func (s *ServiceSearchResult) getSide() string {
	idx := strings.Index(s.ServiceName, ":")
	if idx < 0 {
		return ""
	}
	return s.ServiceName[0:idx]
}

func (s *ServiceSearchResult) getInterfaceName() string {
	idx := strings.Index(s.ServiceName, ":")
	if idx < 0 {
		return s.ServiceName
	}
	return s.ServiceName[idx+1:]
}

// ToInterface .
func (s *ServiceSearchResult) ToInterface() *pb.Interface {
	var i pb.Interface
	i.Interfacename = s.getInterfaceName()
	switch s.getSide() {
	case "consumers:":
		i.Consumerlist = append(i.Consumerlist, s.getIPs()...)
		for k, v := range s.getOwnerMap() {
			i.Consumermap[k] = v
		}
	case "providers:":
		i.Providerlist = append(i.Providerlist, s.getIPs()...)
		for k, v := range s.getOwnerMap() {
			i.Providermap[k] = v
		}
	}
	return &i
}

func (s *ServiceSearchResult) getIPs() []string {
	if s.IPList != nil {
		return s.IPList
	}
	if len(s.ClusterMap) <= 0 || s.ClusterMap["DEFAULT"] == nil || len(s.ClusterMap["DEFAULT"].Hosts) <= 0 {
		s.IPList = make([]string, 0)
		return nil
	}
	for _, item := range s.ClusterMap["DEFAULT"].Hosts {
		if item == nil {
			continue
		}
		s.IPList = append(s.IPList, item.IP)
	}
	return s.IPList
}

func (s *ServiceSearchResult) getOwnerMap() map[string]*pb.InterfaceOwner {
	if s.OwnerMap != nil {
		return s.OwnerMap
	}
	s.OwnerMap = make(map[string]*pb.InterfaceOwner)
	if len(s.ClusterMap) <= 0 || s.ClusterMap["DEFAULT"] == nil || len(s.ClusterMap["DEFAULT"].Hosts) <= 0 {
		return s.OwnerMap
	}
	for _, item := range s.ClusterMap["DEFAULT"].Hosts {
		if item == nil {
			continue
		}
		owner := getOwner(item.IP, item.MetaData["owner"])
		if owner == nil {
			continue
		}
		s.OwnerMap[item.IP] = owner
	}
	return s.OwnerMap
}

func getOwner(ip, owner string) *pb.InterfaceOwner {
	ip, owner = strings.TrimSpace(ip), strings.TrimSpace(owner)
	if len(ip) <= 0 || len(owner) <= 0 {
		return nil
	}
	parts := strings.Split(owner, "_")
	if len(parts) < 3 {
		return nil
	}
	o := pb.InterfaceOwner{
		Ip:        ip,
		Owner:     owner,
		ProjectId: parts[0],
		Env:       parts[1],
		HostIp:    parts[2],
	}
	if len(parts) >= 5 {
		o.ApplicationId = parts[3]
		o.Feature = parts[4]
	}
	if len(parts) >= 7 {
		o.ServiceName = parts[6]
	}
	return &o
}

func (s *ServiceSearchResult) getHostByIP(ip string) *ServiceHost {
	return nil
}

// ServiceHosts
type ServiceHosts struct {
	Hosts []*ServiceHost `json:"hosts"`
}

// ServiceHost .
type ServiceHost struct {
	Valid    bool              `json:"valid"`
	Port     int64             `json:"port"`
	IP       string            `json:"ip"`
	Weight   int64             `json:"weight"`
	Enabled  bool              `json:"enabled"`
	MetaData map[string]string `json:"metadata"`
}

func (s *ServiceHost) getDomain() string {
	h := s.MetaData["SELF_HOST"]
	if len(h) <= 0 {
		return ""
	}
	return net.JoinHostPort(h, strconv.FormatInt(s.Port, 10))
}

func (s *ServiceHost) ToHTTPServiceItem() *pb.HTTPServiceItem {
	var data pb.HTTPServiceItem
	data.Address = net.JoinHostPort(s.IP, strconv.FormatInt(s.Port, 10))
	data.Online = s.Enabled
	return &data
}
