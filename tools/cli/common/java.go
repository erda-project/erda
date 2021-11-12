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

package common

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func ParseSpringBoot() (map[string]string, error) {
	p, err := parsePom()
	if err != nil {
		return nil, err
	}

	isSpringBootJar := false
	for _, plugin := range p.Build.Plugins {
		if plugin.GroupID == "org.springframework.boot" &&
			plugin.ArtifactID == "spring-boot-maven-plugin" {
			isSpringBootJar = true
			break
		}
	}
	if !isSpringBootJar {
		return nil, errors.New("Not a Spring Boot project.")
	}

	serviceTargetName := p.Build.FinalName
	if serviceTargetName == "" {
		serviceTargetName = fmt.Sprintf("%s-%s", p.Name, p.Version)
	}

	port, err := parsePort()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"ServiceName":       p.Name,
		"ServiceTargetName": serviceTargetName,
		"ServicePort":       strconv.Itoa(port),
	}, nil
}

func parsePort() (int, error) {
	port, err := parseYaml()
	if err != nil {
		return -1, err
	}

	if port == 0 {
		port, err = parseProperties()
		if err != nil {
			return -1, err
		}
	}

	if port == 0 {
		port = 8080
	}

	return port, nil
}

func parseYaml() (int, error) {
	f, err := os.Open("src/main/resources/application.yaml")
	if os.IsNotExist(err) {
		if os.IsNotExist(err) {
			return 0, nil
		}
	} else if err != nil {
		return -1, err
	}

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return -1, err
	}

	var app Application
	err = yaml.Unmarshal(bytes, &app)
	if err != nil {
		return -1, err
	}

	if app.Server != nil && app.Server.Port != 0 {
		return app.Server.Port, nil
	}

	return 0, nil
}

func parseProperties() (int, error) {
	f, err := os.Open("src/main/resources/application.properties")
	if os.IsNotExist(err) {
		f, err = os.Open("config/application.properties")
		if os.IsNotExist(err) {
			return 0, nil
		}
	}
	if err != nil {
		return -1, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.Contains(l, "=") {
			ps := strings.Split(l, "=")
			if len(ps) == 2 && strings.TrimSpace(ps[0]) == "server.port" {
				port, err := strconv.Atoi(strings.TrimSpace(ps[1]))
				if err != nil {
					return -1, err
				}
				return port, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return -1, err
	}

	return 0, nil
}

func parsePom() (*Project, error) {
	file, err := os.Open("./pom.xml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, _ := ioutil.ReadAll(file)
	var project Project
	err = xml.Unmarshal(b, &project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

type Project struct {
	XMLName xml.Name `xml:"project"`
	Version string   `xml:"version"`
	Name    string   `xml:"name"`
	Build   Build    `xml:"build"`
}

type Build struct {
	FinalName string   `xml:"finalName"`
	Plugins   []Plugin `xml:"plugins>plugin"`
}

type Plugin struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
}

type Application struct {
	Server *Server `yaml:"server"`
}

type Server struct {
	Port int `yaml:"port"`
}
