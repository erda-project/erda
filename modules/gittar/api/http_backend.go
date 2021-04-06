package api

import (
	"path"
	"strings"

	"github.com/erda-project/erda/modules/gittar/helper"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

// GetRepoHead function
func GetRepoHead(context *webcontext.Context) {
	helper.SendTextFile("HEAD", context)
}

// GetRepoInfoRefs function
func GetRepoInfoRefs(c *webcontext.Context) {
	serviceName := strings.Trim(c.Query("service"), "git-")
	if serviceName != "" {
		helper.RunAdvertisement(serviceName, c)
	} else {
		helper.SendInfoPacks(c)
	}
}

// GetRepoObjects function
func GetRepoObjects(c *webcontext.Context) {
	prefix := c.Param("prefix")
	suffix := c.Param("suffix")
	switch prefix {
	case "pack":
		pack := suffix
		isIdx := strings.HasSuffix(pack, "idx")
		helper.SendPackIdxFile(suffix, isIdx, c)
	case "info":
		if suffix == "packs" {
			helper.SendInfoPacks(c)
		} else {
			file := suffix
			helper.SendTextFile(path.Join("objects", "info", file), c)
		}
	default:
		helper.SendLooseObject(prefix, suffix, c)
	}
}

// ServiceRepoRPC function
func ServiceRepoRPC(c *webcontext.Context) {
	service := c.Param("service")
	if service == "receive-pack" {
		// 检查仓库是否锁定
		isLocked, err := c.Service.GetRepoLocked(c.Repository.ProjectId, c.Repository.ApplicationId)
		if err != nil {
			c.Abort(err)
			return
		}
		if isLocked {
			c.Abort(ERROR_REPO_LOCKED)
			return
		}
	}
	helper.RunProcess(service, c)
}
