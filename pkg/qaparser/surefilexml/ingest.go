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

package surefilexml

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/qaparser"
)

// IngestDir will search the given directory for XML files and return a slice
// of all contained JUnit test suite definitions.
func IngestDir(directory string) ([]*apistructs.TestSuite, error) {
	var filenames []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Add all regular files that end with ".xml"
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".xml") {
			filenames = append(filenames, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return IngestFiles(filenames)
}

// IngestFiles will parse the given XML files and return a slice of all
// contained JUnit test suite definitions.
func IngestFiles(filenames []string) ([]*apistructs.TestSuite, error) {
	var all []*apistructs.TestSuite

	for _, filename := range filenames {
		suites, err := IngestFile(filename)
		if err != nil {
			return nil, err
		}
		all = append(all, suites...)
	}

	return all, nil
}

// IngestFile will parse the given XML file and return a slice of all contained
// JUnit test suite definitions.
func IngestFile(filename string) ([]*apistructs.TestSuite, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Ingest(data)
}

// Ingest will parse the given XML data and return a slice of all contained
// JUnit test suite definitions.
func Ingest(data []byte) ([]*apistructs.TestSuite, error) {
	var (
		suiteChan = make(chan apistructs.TestSuite)
		suites    []*apistructs.TestSuite
	)

	nodes, err := qaparser.NodeParse(data)
	if err != nil {
		return nil, err
	}

	go func() {
		findSuites(nodes, suiteChan)
		close(suiteChan)
	}()

	for temp := range suiteChan {
		suite := temp
		if suite.Totals.Tests == 0 || suite.Tests == nil {
			continue
		}
		suites = append(suites, &suite)
	}

	return suites, nil
}

// findparser.Suites performs a depth-first search through the XML document, and
// attempts to ingest any "testsuite" tags that are encountered.
func findSuites(nodes []qaparser.XmlNode, suites chan apistructs.TestSuite) {
	for _, node := range nodes {
		switch node.XMLName.Local {
		case "testsuite":
			suites <- ingestSuite(node)
		default:
			findSuites(node.Nodes, suites)
		}
	}
}

func ingestSuite(root qaparser.XmlNode) apistructs.TestSuite {
	suite := apistructs.TestSuite{
		Name:    root.Attr("name"),
		Package: root.Attr("package"),
	}

	for _, node := range root.Nodes {
		switch node.XMLName.Local {
		case "testcase":
			testcase := ingestTestcase(node)
			suite.Tests = append(suite.Tests, testcase)
		case "properties":
			props := ingestProperties(node)
			suite.Properties = props
		case "system-out":
			suite.SystemOut = string(node.Content)
		case "system-err":
			suite.SystemErr = string(node.Content)
		}
	}

	su := &qaparser.Suite{&suite}

	su.Aggregate()
	return suite
}

func ingestProperties(root qaparser.XmlNode) map[string]string {
	props := make(map[string]string, len(root.Nodes))

	for _, node := range root.Nodes {
		switch node.XMLName.Local {
		case "property":
			name := node.Attr("name")
			value := node.Attr("value")
			props[name] = value
		}
	}

	return props
}

func ingestTestcase(root qaparser.XmlNode) *apistructs.Test {
	test := apistructs.Test{
		Name:      root.Attr("name"),
		Classname: root.Attr("classname"),
		Duration:  duration(root.Attr("time")),
		Status:    apistructs.TestStatusPassed,
	}

	for _, node := range root.Nodes {
		switch node.XMLName.Local {
		case "skipped":
			test.Status = apistructs.TestStatusSkipped
		case "failure":
			test.Error = ingestError(node)
			test.Status = apistructs.TestStatusFailed
		case "error":
			test.Error = ingestError(node)
			test.Status = apistructs.TestStatusError
		case "system-out":
			test.SystemOut = string(node.Content)
		case "system-err":
			test.SystemErr = string(node.Content)
		}
	}

	return &test
}

func ingestError(root qaparser.XmlNode) apistructs.TestError {
	return apistructs.TestError{
		Body:    string(root.Content),
		Type:    root.Attr("type"),
		Message: root.Attr("message"),
	}
}

func duration(t string) time.Duration {
	// Check if there was a valid decimal value
	if s, err := strconv.ParseFloat(t, 64); err == nil {
		return time.Duration(s*1000000) * time.Microsecond
	}

	// Check if there was a valid duration string
	if d, err := time.ParseDuration(t); err == nil {
		return d
	}

	return 0
}
