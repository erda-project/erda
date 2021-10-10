package code_coverage

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/erda-project/erda/apistructs"
)

type CounterType string

const (
	LineIdx = iota
	InstructionIdx
	ComplexityIdx
	ClassIdx
	MethodIdx
	LinePercentIdx
	InstructionPercentIdx
)

var (
	LineCounter        CounterType = "LINE"
	InstructionCounter CounterType = "INSTRUCTION"
	ComplexityCounter  CounterType = "COMPLEXITY"
	ClassCounter       CounterType = "CLASS"
	MethodCounter      CounterType = "METHOD"
)

func (c CounterType) IsLineType() bool {
	return c == MethodCounter
}

func (c CounterType) IsInstructionType() bool {
	return c == InstructionCounter
}

func (c CounterType) GetValueIdx() int {
	switch c {
	case LineCounter:
		return LineIdx
	case InstructionCounter:
		return InstructionIdx
	case ComplexityCounter:
		return ComplexityIdx
	case MethodCounter:
		return MethodIdx
	case ClassCounter:
		return ClassIdx
	default:
		return LineIdx
	}
}

type Report struct {
	ProjectID   uint64    `json:"projectID"`
	ProjectName string    `json:"projectName"`
	XMLName     xml.Name  `xml:"report"`
	Packages    []Package `xml:"package"`
	Counters    []Counter `xml:"counter"`
}

type Counter struct {
	Covered int         `xml:"covered,attr"`
	Missed  int         `xml:"missed,attr"`
	Type    CounterType `xml:"type,attr"`
}

type Package struct {
	Name     string    `xml:"name,attr"`
	Classes  []Class   `xml:"class"`
	Counters []Counter `xml:"counter"`
}

type Class struct {
	Name           string    `xml:"name,attr"`
	SourceFilename string    `xml:"sourcefilename,attr"`
	Methods        []Method  `xml:"method"`
	Counters       []Counter `xml:"counter"`
}

type Method struct {
	Name     string    `xml:"name,attr"`
	Desc     string    `xml:"desc,attr"`
	Line     int       `xml:"line,attr"`
	Counters []Counter `xml:"counter"`
}

func convertXmlToReport(source []byte) (Report, error) {
	data := Report{}
	err := xml.Unmarshal(source, &data)
	if err != nil {
		return Report{}, err
	}
	return data, nil
}

func decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

func setNodeValue(root *apistructs.CodeCoverageNode, counters []Counter) {
	root.Value = make([]float64, 7)
	for _, c := range counters {
		v := float64(c.Missed + c.Covered)
		root.Value[c.Type.GetValueIdx()] = v
		if c.Type.IsLineType() && v != 0 {
			root.Value[LinePercentIdx] = decimal((float64(c.Covered) / v) * 100)
		}
		if c.Type.IsInstructionType() && v != 0 {
			root.Value[InstructionPercentIdx] = decimal((float64(c.Covered) / v) * 100)
		}
	}
}

func convertReportToTree(r Report) ([]*apistructs.CodeCoverageNode, float64) {
	var root = &apistructs.CodeCoverageNode{}
	if r.Packages == nil {
		return []*apistructs.CodeCoverageNode{}, 0
	}
	setNodeValue(root, r.Counters)
	coverage := root.Value[LinePercentIdx]
	root.Name = r.ProjectName
	for _, p := range r.Packages {
		pNode := &apistructs.CodeCoverageNode{}
		setNodeValue(pNode, p.Counters)
		pNode.Name = p.Name
		//for _, c := range p.Classes {
		//	cNode := &apistructs.CodeCoverageNode{}
		//	setNodeValue(cNode, c.Counters)
		//	cNode.Name = c.Name
		//	for _, m := range c.Methods {
		//		mNode := &apistructs.CodeCoverageNode{}
		//		mNode.Name = m.Name
		//		setNodeValue(mNode, m.Counters)
		//		cNode.Nodes = append(cNode.Nodes, mNode)
		//	}
		//	pNode.Nodes = append(pNode.Nodes, cNode)
		//}
		root.Nodes = append(root.Nodes, pNode)
	}
	return []*apistructs.CodeCoverageNode{root}, coverage
}

func getAnalyzeJson(projectID uint64, projectName string, data []byte) ([]*apistructs.CodeCoverageNode, float64) {
	report, err := convertXmlToReport(data)
	report.ProjectID = projectID
	report.ProjectName = projectName
	if err != nil {
		log.Fatal(err)
	}

	root, coverage := convertReportToTree(report)
	return root, coverage
}
