package uc

import (
	"strconv"

	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/common"
)

func userMapper(user *GetUser) *common.User {
	return &common.User{
		ID:        strconv.Itoa(user.ID),
		Name:      user.Name,
		Nick:      user.Nick,
		AvatarURL: user.AvatarURL,
		Phone:     user.Phone,
		Email:     user.Email,
	}
}

func usersMapper(users []*GetUser) []*common.User {
	return lo.Map(users, func(item *GetUser, _ int) *common.User {
		return userMapper(item)
	})
}

func managedUserMapper(u *UserInPaging) *pb.ManagedUser {
	var lastLoginAt *timestamppb.Timestamp
	if u.LastLoginAt != nil && !u.LastLoginAt.IsZero() {
		lastLoginAt = timestamppb.New(u.LastLoginAt.Time)
	}

	var pwdExpireAt *timestamppb.Timestamp
	if u.PwdExpireAt != nil && !u.PwdExpireAt.IsZero() {
		pwdExpireAt = timestamppb.New(u.PwdExpireAt.Time)
	}

	return &pb.ManagedUser{
		Id:          cast.ToString(u.Id),
		Name:        u.Username,
		Nick:        u.Nickname,
		Avatar:      u.Avatar,
		Phone:       u.Mobile,
		Email:       u.Email,
		LastLoginAt: lastLoginAt,
		PwdExpireAt: pwdExpireAt,
		Source:      u.Source,
		Locked:      u.Locked,
	}
}
