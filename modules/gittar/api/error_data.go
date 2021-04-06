package api

import "errors"

type Map map[string]interface{}

var (
	ERROR_CODE_INTERNAL     = "500"
	ERROR_CODE_NOT_FOUND    = "404"
	ERROR_CODE_INVALID_ARGS = "400"
)

var (
	ERROR_NOT_FILE       = errors.New("path not file")
	ERROR_PATH_NOT_FOUND = errors.New("path not exist")
	ERROR_DB             = errors.New("db error")
	ERROR_ARG_ID         = errors.New("id parse failed")
	ERROR_HOOK_NOT_FOUND = errors.New("hook not found")
	ERROR_LOCKED_DENIED  = errors.New("locked denied")
	ERROR_REPO_LOCKED    = errors.New("repo locked")
)
