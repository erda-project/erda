package promxp

import (
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var EMPTY_LABLES = make([]string, 0)

type labelPairSorter []*dto.LabelPair

func (s labelPairSorter) Len() int {
	return len(s)
}

func (s labelPairSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s labelPairSorter) Less(i, j int) bool {
	return s[i].GetName() < s[j].GetName()
}

func makeLabelPairs(constLabelPairs prometheus.Labels, variableLabels []string, labelValues []string) []*dto.LabelPair {
	totalLen := len(variableLabels) + len(constLabelPairs)
	if totalLen == 0 {
		// Super fast path.
		return nil
	}
	labelPairs := make([]*dto.LabelPair, 0, totalLen)
	for n, v := range constLabelPairs {
		labelPairs = append(labelPairs, &dto.LabelPair{
			Name:  proto.String(n),
			Value: proto.String(v),
		})
	}
	for i, n := range variableLabels {
		labelPairs = append(labelPairs, &dto.LabelPair{
			Name:  proto.String(n),
			Value: proto.String(labelValues[i]),
		})
	}
	sort.Sort(labelPairSorter(labelPairs))
	return labelPairs
}
