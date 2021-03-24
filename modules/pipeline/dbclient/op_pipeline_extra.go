package dbclient

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (client *Client) GetPipelineExtraByPipelineID(pipelineID uint64, ops ...SessionOption) (spec.PipelineExtra, bool, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var extra spec.PipelineExtra
	extra.PipelineID = pipelineID
	found, err := session.Get(&extra)
	return extra, found, err
}

func (client *Client) CreatePipelineExtra(extra *spec.PipelineExtra, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.InsertOne(extra)
	return err
}

func (client *Client) UpdatePipelineExtraByPipelineID(pipelineID uint64, extra *spec.PipelineExtra, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.Table(&spec.PipelineExtra{}).AllCols().Where("pipeline_id=?", pipelineID).Update(extra)
	return err
}

func (client *Client) UpdatePipelineExtraExtraInfoByPipelineID(pipelineID uint64, extraInfo spec.PipelineExtraInfo, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	// get extra
	extra, found, err := client.GetPipelineExtraByPipelineID(pipelineID, ops...)
	if err != nil {
		return err
	}
	if !found {
		return errors.Errorf("not found extra")
	}

	// update extra.ExtraInfo
	extra.Extra = extraInfo
	_, err = session.ID(extra.PipelineID).Cols("extra").
		Update(&spec.PipelineExtra{Extra: extraInfo}, spec.PipelineExtra{PipelineID: pipelineID})
	return err
}

func (client *Client) ListPipelineExtrasByPipelineIDs(pipelineIDs []uint64, ops ...SessionOption) (map[uint64]spec.PipelineExtra, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var extras []spec.PipelineExtra
	if err := session.In("pipeline_id", pipelineIDs).Find(&extras); err != nil {
		return nil, err
	}
	extrasMap := make(map[uint64]spec.PipelineExtra, len(extras))
	for _, extra := range extras {
		extrasMap[extra.PipelineID] = extra
	}
	return extrasMap, nil
}

func (client *Client) UpdatePipelineProgress(pipelineID uint64, progress int, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(pipelineID).Cols("progress").Update(&spec.PipelineExtra{Progress: progress})
	return err
}

func (client *Client) UpdatePipelineExtraSnapshot(pipelineID uint64, snapshot spec.Snapshot, ops ...SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(pipelineID).Cols("snapshot").Update(&spec.PipelineExtra{Snapshot: snapshot})
	return err
}
