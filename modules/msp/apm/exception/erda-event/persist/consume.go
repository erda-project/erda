package persist

import (
	"encoding/json"
	"github.com/erda-project/erda/modules/msp/apm/exception"
	"time"
)


func (p *provider) decodeEvent(key, value []byte, topic *string, timestamp time.Time) (interface{}, error) {
	data := &exception.Erda_event{}
	if err := json.Unmarshal(value, data); err != nil {
		p.stats.DecodeError(value, err)
		if p.Cfg.PrintInvalidEvent {
			p.Log.Warnf("unknown format event data: %s", string(value))
		} else {
			p.Log.Warnf("failed to decode event: %v", err)
		}
		return nil, err
	}


	if err := p.validator.Validate(data); err != nil {
		p.stats.ValidateError(data)
		if p.Cfg.PrintInvalidEvent {
			p.Log.Warnf("invalid event data: %s", string(value))
		} else {
			p.Log.Warnf("invalid event: %v", err)
		}
		return nil, err
	}
	if err := p.metadata.Process(data); err != nil {
		p.stats.MetadataError(data, err)
		p.Log.Errorf("failed to process event metadata: %v", err)
	}
	return data, nil
}

func (p *provider) handleReadError(err error) error {
	p.Log.Errorf("failed to read event from kafka: %s", err)
	return nil // return nil to continue read
}

func (p *provider) handleWriteError(list []interface{}, err error) error {
	p.Log.Errorf("failed to write event into storage: %s", err)
	return nil // return nil to continue consume
}

func (p *provider) confirmErrorHandler(err error) error {
	p.Log.Errorf("failed to confirm event from kafka: %s", err)
	return err // return error to exit
}
