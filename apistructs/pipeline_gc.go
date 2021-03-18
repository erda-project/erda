package apistructs

import (
	"time"
)

type PipelineGCInfo struct {
	CreatedAt time.Time `json:"createdAt,omitempty"`
	GCAt      time.Time `json:"gcAt,omitempty"`
	TTL       uint64    `json:"ttl,omitempty"`
	LeaseID   string    `json:"leaseID,omitempty"`
	Data      []byte    `json:"data,omitempty"`
}

func MakePipelineGCInfo(ttl uint64, leaseID string, data []byte) PipelineGCInfo {
	now := time.Now()
	return PipelineGCInfo{
		CreatedAt: now,
		GCAt:      now.Add(time.Second * time.Duration(ttl)),
		TTL:       ttl,
		LeaseID:   leaseID,
		Data:      data,
	}
}

type PipelineGCDBOption struct {
	NeedArchive bool `json:"needArchive"`
}
