package dao

import (
	"github.com/pkg/errors"
)

// dao层错误码统一定义
var (
	ErrNotFoundOrg         = errors.New("org not found")
	ErrNotFoundProject     = errors.New("project not found")
	ErrNotFoundApplication = errors.New("application not found")
	ErrNotFoundMember      = errors.New("member not found")
	ErrNotFoundTicket      = errors.New("ticket not found")
	ErrNotFoundPublisher   = errors.New("publisher not found")
	ErrNotFoundCertificate = errors.New("certificate not found")
	ErrNotFoundApprove     = errors.New("approve not found")
	ErrNotFoundUsecase     = errors.New("usecase not found")
)
