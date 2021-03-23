package spec

import (
	"time"
)

type CIV3BuildCache struct {
	ID          int64     `json:"id" xorm:"pk autoincr"`
	Name        string    `json:"name"`
	ClusterName string    `json:"clusterName"`
	LastPullAt  time.Time `json:"lastPullAt"`
	CreatedAt   time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt   time.Time `json:"updatedAt" xorm:"updated"`
	DeletedAt   time.Time `xorm:"deleted"`
}

func (*CIV3BuildCache) TableName() string {
	return "ci_v3_build_caches"
}
