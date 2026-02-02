// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iam

import (
	"testing"
	"time"
)

func Test_formatIAMTime(t *testing.T) {
	tm, err := formatIAMTime("2006-01-02T15:04:05")
	if err != nil {
		t.Fatalf("formatIAMTime: %v", err)
	}
	expected := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	if !tm.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, tm)
	}

	_, err = formatIAMTime("invalid")
	if err == nil {
		t.Fatal("expected error for invalid time string")
	}
}

func Test_userMapper(t *testing.T) {
	u := &UserDto{
		ID:       100,
		Username: "u",
		Nickname: "n",
		Mobile:   "138",
		Email:    "e@e.com",
		Avatar:   "av",
	}
	out := userMapper(u)
	if out.Id != "100" {
		t.Errorf("expected Id 100, got %s", out.Id)
	}
	if out.Name != "u" || out.Nick != "n" || out.Phone != "138" || out.Email != "e@e.com" || out.Avatar != "av" {
		t.Errorf("unexpected mapping: %+v", out)
	}
}

func Test_managedUserMapper(t *testing.T) {
	u := &UserDto{
		ID:               200,
		Username:         "mu",
		Nickname:         "mn",
		Mobile:           "139",
		Email:            "m@m.com",
		Avatar:           "mav",
		Locked:           true,
		LastLoginAt:      "2024-01-02T10:00:00",
		PasswordExpireAt: "2025-06-01T12:00:00",
	}
	out, err := managedUserMapper(u)
	if err != nil {
		t.Fatalf("managedUserMapper: %v", err)
	}
	if out.Id != "200" || out.Name != "mu" || out.Nick != "mn" || out.Phone != "139" || out.Email != "m@m.com" || out.Avatar != "mav" {
		t.Errorf("unexpected mapping: %+v", out)
	}
	if !out.Locked {
		t.Error("expected Locked true")
	}
	if out.LastLoginAt == nil {
		t.Error("expected LastLoginAt set")
	}
	if out.PwdExpireAt == nil {
		t.Error("expected PwdExpireAt set")
	}
}

func Test_managedUserMapper_emptyDates(t *testing.T) {
	u := &UserDto{ID: 1, Username: "x", Nickname: "y"}
	out, err := managedUserMapper(u)
	if err != nil {
		t.Fatalf("managedUserMapper: %v", err)
	}
	if out.LastLoginAt != nil || out.PwdExpireAt != nil {
		t.Error("expected nil timestamps when dates empty")
	}
}

func Test_managedUserMapper_invalidDate(t *testing.T) {
	u := &UserDto{ID: 1, Username: "x", LastLoginAt: "not-a-date"}
	_, err := managedUserMapper(u)
	if err == nil {
		t.Fatal("expected error for invalid LastLoginAt")
	}
}

func Test_isEmptyTrim(t *testing.T) {
	if !isEmptyTrim("") {
		t.Error("empty string should be true")
	}
	if !isEmptyTrim("   ") {
		t.Error("whitespace only should be true")
	}
	if isEmptyTrim("a") {
		t.Error("non-empty should be false")
	}
	if isEmptyTrim("  a  ") {
		t.Error("trimmed non-empty should be false")
	}
}
