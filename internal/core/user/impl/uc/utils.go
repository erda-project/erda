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
	return &pb.ManagedUser{
		Id:          cast.ToString(u.Id),
		Name:        u.Username,
		Nick:        u.Nickname,
		Avatar:      u.Avatar,
		Phone:       u.Mobile,
		Email:       u.Email,
		LastLoginAt: timestamppb.New(u.LastLoginAt),
		PwdExpireAt: timestamppb.New(u.PwdExpireAt),
		Source:      u.Source,
		Locked:      u.Locked,
	}
}
