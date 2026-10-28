package user_info

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

func TestMapAndEnrich_UserInfo(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	m := common.NewJSONLinesMapper("user_info_mapper", mapUserInfo)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")

	var rows []*UserInfo
	for _, l := range lines {
		r, err := m.Map(context.Background(), l)
		if err != nil {
			t.Fatalf("map: %v", err)
		}
		rows = append(rows, r)
	}

	if rows[0].UserIsAdmin == nil || *rows[0].UserIsAdmin {
		t.Errorf("row 0 UserIsAdmin: got %v, want false", rows[0].UserIsAdmin)
	}
	if rows[1].UserIsAdmin == nil || !*rows[1].UserIsAdmin {
		t.Errorf("row 1 UserIsAdmin: got %v, want true", rows[1].UserIsAdmin)
	}
	// "N/A" and "0" are placeholders, not values.
	if rows[0].MonthsSinceReset != nil {
		t.Errorf("MonthsSinceReset: got %v, want nil", *rows[0].MonthsSinceReset)
	}
	if rows[0].PasswordLastSet != nil {
		t.Errorf("PasswordLastSet: got %v, want nil", rows[0].PasswordLastSet)
	}
	if rows[0].LogonTime == nil || !rows[0].LogonTime.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("LogonTime: got %v", rows[0].LogonTime)
	}

	out, err := (UserInfoTable{}).EnrichRow(rows[0], schema.SourceEnrichment{})
	if err != nil {
		t.Fatalf("enrich: %v", err)
	}
	if !out.TpTimestamp.Equal(time.Unix(1700000060, 0)) {
		t.Errorf("TpTimestamp: got %v", out.TpTimestamp)
	}
	if len(out.TpUsernames) != 2 {
		t.Errorf("TpUsernames: got %v", out.TpUsernames)
	}
}
