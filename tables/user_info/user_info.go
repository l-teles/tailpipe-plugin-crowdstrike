package user_info

import "github.com/turbot/tailpipe-plugin-sdk/schema"

// UserInfo represents one row in an FDR UserInfo snapshot — local-account
// inventory observed on a host. The wire format does NOT include `aid` for
// every record (only `cid` is reliably present), so use UserSidReadable plus
// LastLoggedOnHost when you need to tie a row back to a specific agent.
type UserInfo struct {
	schema.CommonFields

	Cid  *string `parquet:"name=cid"`
	Time *string `parquet:"name=time"` // raw `_time` field on the wire

	AccountType      *string `parquet:"name=account_type"`
	LastLoggedOnHost *string `parquet:"name=last_logged_on_host"`
	LogonTime        *string `parquet:"name=logon_time"`
	LogonType        *string `parquet:"name=logon_type"`
	PasswordLastSet  *string `parquet:"name=password_last_set"`
	User             *string `parquet:"name=user"`
	UserIsAdmin      *string `parquet:"name=user_is_admin"`
	UserName         *string `parquet:"name=user_name"`
	UserSidReadable  *string `parquet:"name=user_sid_readable"`
	MonthsSinceReset *string `parquet:"name=months_since_reset"`

	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func (UserInfo) GetColumnDescriptions() map[string]string {
	// #nosec G101 -- map values are human-readable column descriptions, not credentials.
	return map[string]string{
		"cid":                 "Customer (tenant) identifier.",
		"time":                "Epoch seconds (string) when the record was emitted (delivered as `_time`).",
		"account_type":        "Account type (e.g. Domain, Local, AzureAD).",
		"last_logged_on_host": "Hostname where the account most recently signed in.",
		"logon_time":          "Epoch seconds (string) of the last logon.",
		"logon_type":          "Logon mechanism description (e.g. \"Cached credentials\", \"Interactive\").",
		"password_last_set":   "Epoch seconds (string) the password was last changed (\"0\" if unknown).",
		"user":                "Fully qualified principal (e.g. \"AZUREAD\\\\user@domain\").",
		"user_is_admin":       "\"1\" if the account is a local administrator, \"0\" otherwise.",
		"user_name":           "Short username form (e.g. UPN without prefix).",
		"user_sid_readable":   "Resolved SID for the account.",
		"months_since_reset":  "Months since the password was last reset (or \"N/A\").",
		"payload":             "Full record JSON, including any field not promoted to a typed column.",
	}
}
