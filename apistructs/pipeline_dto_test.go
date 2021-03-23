package apistructs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EnablePipelineVolume(t *testing.T) {
	var tables = []struct {
		StorageConfig            StorageConfig
		EnablePipelineVolume     bool
		EnableCloseNetworkVolume bool
		EnableShareVolume        bool
	}{
		{
			StorageConfig: StorageConfig{
				EnableNFS:   false,
				EnableLocal: true,
			},
			EnablePipelineVolume:     false,
			EnableCloseNetworkVolume: false,
			EnableShareVolume:        true,
		},
		{
			StorageConfig: StorageConfig{
				EnableNFS:   true,
				EnableLocal: false,
			},
			EnablePipelineVolume:     true,
			EnableCloseNetworkVolume: true,
			EnableShareVolume:        false,
		},
		{
			StorageConfig: StorageConfig{
				EnableNFS:   true,
				EnableLocal: false,
			},
			EnablePipelineVolume:     true,
			EnableCloseNetworkVolume: true,
			EnableShareVolume:        false,
		},
	}
	for _, data := range tables {
		if data.EnablePipelineVolume {
			assert.True(t, data.StorageConfig.EnablePipelineVolume(), "not true")
		} else {
			assert.True(t, !data.StorageConfig.EnablePipelineVolume(), "not true")
		}
		if data.EnableCloseNetworkVolume {
			assert.True(t, data.StorageConfig.EnableNFSVolume(), "not true")
		} else {
			assert.True(t, !data.StorageConfig.EnableNFSVolume(), "not true")
		}
		if data.EnableShareVolume {
			assert.True(t, data.StorageConfig.EnableShareVolume(), "not true")
		} else {
			assert.True(t, !data.StorageConfig.EnableShareVolume(), "not true")
		}
	}
}
