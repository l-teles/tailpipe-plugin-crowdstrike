package user_info

import (
	"context"
	"fmt"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

type UserInfoExtractor struct{}

func NewUserInfoExtractor() artifact_source.Extractor { return &UserInfoExtractor{} }

func (UserInfoExtractor) Identifier() string { return "user_info_extractor" }

func (UserInfoExtractor) Extract(_ context.Context, a any) ([]any, error) {
	raw, ok := a.([]byte)
	if !ok {
		return nil, fmt.Errorf("user_info_extractor: expected []byte, got %T", a)
	}
	return common.ExtractJSONLines(raw, "user_info_extractor", buildUserInfo)
}

func buildUserInfo(doc map[string]any) *UserInfo {
	r := &UserInfo{Payload: doc}

	r.Cid = common.StringFromMap(doc, "cid")
	r.Time = common.StringFromMap(doc, "_time")

	r.AccountType = common.StringFromMap(doc, "AccountType")
	r.LastLoggedOnHost = common.StringFromMap(doc, "LastLoggedOnHost")
	r.LogonTime = common.StringFromMap(doc, "LogonTime")
	r.LogonType = common.StringFromMap(doc, "LogonType")
	r.PasswordLastSet = common.StringFromMap(doc, "PasswordLastSet")
	r.User = common.StringFromMap(doc, "User")
	r.UserIsAdmin = common.StringFromMap(doc, "UserIsAdmin")
	r.UserName = common.StringFromMap(doc, "UserName")
	r.UserSidReadable = common.StringFromMap(doc, "UserSid_readable")
	r.MonthsSinceReset = common.StringFromMap(doc, "monthsincereset")

	return r
}
