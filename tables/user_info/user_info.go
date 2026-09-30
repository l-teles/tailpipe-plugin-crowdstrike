package user_info

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

// UserInfo represents one row in an FDR UserInfo snapshot — local-account
// inventory observed on a host. The wire format does NOT include `aid` for
// every record (only `cid` is reliably present), so use UserSidReadable plus
// LastLoggedOnHost when you need to tie a row back to a specific agent.
type UserInfo struct {
	schema.CommonFields

	Cid  *string    `parquet:"name=cid"`
	Time *time.Time `parquet:"name=time"` // `_time` on the wire

	AccountType      *string    `parquet:"name=account_type"`
	LastLoggedOnHost *string    `parquet:"name=last_logged_on_host"`
	LogonTime        *time.Time `parquet:"name=logon_time"`
	LogonType        *string    `parquet:"name=logon_type"`
	PasswordLastSet  *time.Time `parquet:"name=password_last_set"`
	User             *string    `parquet:"name=user"`
	UserIsAdmin      *bool      `parquet:"name=user_is_admin"`
	UserName         *string    `parquet:"name=user_name"`
	UserSidReadable  *string    `parquet:"name=user_sid_readable"`
	MonthsSinceReset *int64     `parquet:"name=months_since_reset"`

	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func mapUserInfo(doc map[string]any) *UserInfo {
	r := &UserInfo{Payload: doc}

	r.Cid = common.StringFromMap(doc, "cid")
	r.Time = common.EpochSecondsFromMap(doc, "_time")

	r.AccountType = common.StringFromMap(doc, "AccountType")
	r.LastLoggedOnHost = common.StringFromMap(doc, "LastLoggedOnHost")
	r.LogonTime = common.EpochSecondsFromMap(doc, "LogonTime")
	r.LogonType = common.StringFromMap(doc, "LogonType")
	r.PasswordLastSet = common.EpochSecondsFromMap(doc, "PasswordLastSet")
	r.User = common.StringFromMap(doc, "User")
	r.UserIsAdmin = common.BoolFromMap(doc, "UserIsAdmin")
	r.UserName = common.StringFromMap(doc, "UserName")
	r.UserSidReadable = common.StringFromMap(doc, "UserSid_readable")
	r.MonthsSinceReset = common.IntFromMap(doc, "monthsincereset")

	return r
}

func (UserInfo) GetColumnDescriptions() map[string]string {
	// #nosec G101 -- map values are human-readable column descriptions, not credentials.
	return map[string]string{
		"cid":                 "Customer (tenant) identifier.",
		"time":                "When the record was emitted (delivered as `_time`).",
		"account_type":        "Account type (e.g. Domain, Local, AzureAD).",
		"last_logged_on_host": "Hostname where the account most recently signed in.",
		"logon_time":          "When the account last logged on.",
		"logon_type":          "Logon mechanism description (e.g. \"Cached credentials\", \"Interactive\").",
		"password_last_set":   "When the password was last changed (null if unknown).",
		"user":                "Fully qualified principal (e.g. \"AZUREAD\\\\user@domain\").",
		"user_is_admin":       "True if the account is a local administrator.",
		"user_name":           "Short username form (e.g. UPN without prefix).",
		"user_sid_readable":   "Resolved SID for the account.",
		"months_since_reset":  "Months since the password was last reset (null when delivered as \"N/A\").",
		"payload":             "Full record JSON, including any field not promoted to a typed column.",
		"tp_timestamp":        "Record time: `time`, or `logon_time` when `time` is absent.",
	}
}
