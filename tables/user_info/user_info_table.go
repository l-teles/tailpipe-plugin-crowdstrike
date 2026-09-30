package user_info

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

const UserInfoTableIdentifier = "crowdstrike_user_info"

type UserInfoTable struct{}

func (UserInfoTable) Identifier() string { return UserInfoTableIdentifier }

func (UserInfoTable) GetDescription() string {
	return "CrowdStrike FDR UserInfo snapshots: local-account inventory observed on each Falcon-managed host."
}

func (UserInfoTable) GetSourceMetadata() ([]*table.SourceMetadata[*UserInfo], error) {
	return common.SourceMetadata(common.NewJSONLinesMapper("user_info_mapper", mapUserInfo)), nil
}

func (UserInfoTable) EnrichRow(row *UserInfo, sourceEnrichmentFields schema.SourceEnrichment) (*UserInfo, error) {
	ts, err := common.FirstTime(row.Time, row.LogonTime)
	if err != nil {
		return nil, err
	}

	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = ts
	row.TpDate = ts.Truncate(24 * time.Hour)

	if row.UserName != nil {
		row.TpUsernames = append(row.TpUsernames, *row.UserName)
	}
	if row.User != nil && (row.UserName == nil || *row.User != *row.UserName) {
		row.TpUsernames = append(row.TpUsernames, *row.User)
	}
	if row.UserSidReadable != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:sid:"+*row.UserSidReadable)
	}

	return row, nil
}
