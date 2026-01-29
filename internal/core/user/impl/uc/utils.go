package uc

import (
	"strconv"

	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/user/pb"
)

func userMapper(user *GetUser) *commonpb.UserInfo {
	return &commonpb.UserInfo{
		Id:     strconv.Itoa(user.ID),
		Name:   user.Name,
		Nick:   user.Nick,
		Avatar: user.AvatarURL,
		Phone:  user.Phone,
		Email:  user.Email,
	}
}

func usersMapper(users []*GetUser) []*commonpb.UserInfo {
	return lo.Map(users, func(item *GetUser, _ int) *commonpb.UserInfo {
		return userMapper(item)
	})
}

func managedUserMapper(u *UserDto) *pb.ManagedUser {
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
