package helper

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/modules/gittar/webcontext"
)

// HeaderNoCache writing function
func headerNoCache(c *webcontext.Context) {
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	c.Header("cache-Control", "no-cache, max-age=0, must-revalidate")
}

// HeaderCacheForever writing function
func headerCacheForever(c *webcontext.Context) {
	now := time.Now().Unix()
	expires := now + 31536000
	c.Header("Date", fmt.Sprintf("%d", now))
	c.Header("Expires", fmt.Sprintf("%d", expires))
	c.Header("cache-Control", "public, max-age=31536000")
}
