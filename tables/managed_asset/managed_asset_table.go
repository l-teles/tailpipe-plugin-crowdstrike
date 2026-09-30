package managed_asset

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

const ManagedAssetTableIdentifier = "crowdstrike_managed_asset"

type ManagedAssetTable struct{}

func (ManagedAssetTable) Identifier() string { return ManagedAssetTableIdentifier }

func (ManagedAssetTable) GetDescription() string {
	return "CrowdStrike FDR ManagedAssets snapshots: network interface and gateway info per Falcon-managed agent."
}

func (ManagedAssetTable) GetSourceMetadata() ([]*table.SourceMetadata[*ManagedAsset], error) {
	return common.SourceMetadata(common.NewJSONLinesMapper("managed_asset_mapper", mapManagedAsset)), nil
}

func (ManagedAssetTable) EnrichRow(row *ManagedAsset, sourceEnrichmentFields schema.SourceEnrichment) (*ManagedAsset, error) {
	ts, err := common.FirstTime(row.Time)
	if err != nil {
		return nil, err
	}

	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now().UTC()
	row.TpTimestamp = ts
	row.TpDate = ts.Truncate(24 * time.Hour)

	if row.LocalAddressIP4 != nil {
		row.TpSourceIP = row.LocalAddressIP4
		row.TpIps = append(row.TpIps, *row.LocalAddressIP4)
	}
	if row.GatewayIP != nil {
		row.TpIps = append(row.TpIps, *row.GatewayIP)
	}
	if row.Aid != nil {
		row.TpAkas = append(row.TpAkas, "crowdstrike:aid:"+*row.Aid)
	}

	return row, nil
}
