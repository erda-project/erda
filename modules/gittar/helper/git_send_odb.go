package helper

import (
	"path"

	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

// SendTextFile file
func SendTextFile(file string, c *webcontext.Context) {
	headerNoCache(c)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.File(path.Join(c.Repository.DiskPath(), file))
}

// SendInfoPacks info packs
func SendInfoPacks(c *webcontext.Context) {
	headerNoCache(c)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.File(c.MustGet("repository").(*gitmodule.Repository).InfoPacksPath())
}

// SendLooseObject loose object
func SendLooseObject(prefix string, suffix string, c *webcontext.Context) {
	headerCacheForever(c)
	c.Header("Content-Type", "application/x-git-loose-object")
	c.File(c.MustGet("repository").(*gitmodule.Repository).LooseObjectPath(prefix, suffix))
}

// SendPackIdxFile file
func SendPackIdxFile(pack string, isIdx bool, c *webcontext.Context) {
	headerCacheForever(c)
	if isIdx == false {
		c.Header("Content-Type", "application/x-git-packed-objects")
	} else {
		c.Header("Content-Type", "application/x-git-packed-objects-toc")
	}
	c.File(c.MustGet("repository").(*gitmodule.Repository).PackIdxPath(pack))
}
