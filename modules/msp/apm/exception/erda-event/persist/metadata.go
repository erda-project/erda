package persist

import (
	"github.com/erda-project/erda/modules/msp/apm/exception"
)

type MetadataProcessor interface {
	Process(data *exception.Erda_event) error
}

func newMetadataProcessor(cfg *config) MetadataProcessor {
	return NopMetadataProcessor
}

type nopMetadataProcessor struct{}

func (*nopMetadataProcessor) Process(data *exception.Erda_event) error { return nil }

// NopMetadataProcessor .
var NopMetadataProcessor MetadataProcessor = &nopMetadataProcessor{}