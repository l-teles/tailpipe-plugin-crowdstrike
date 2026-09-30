// Package common holds helpers shared by the CrowdStrike table packages.
package common

import "github.com/turbot/pipe-fittings/v2/utils"

// DefaultBatchLayout is the shared FDR file layout (relative to the
// per-tenant prefix). It matches both FDR layout variants observed in the
// wild:
//
//  1. Hive-style ("classic" FDR):
//     batch=<uuid>/year=YYYY/month=MM/day=DD/hour=HH/platform=<plat>/part-*.txt.gz
//  2. Flat ("newer" FDR / recently provisioned tenants):
//     <uuid>/part-*.gz
//
// The `batch=` prefix, date partitioning, platform partitioning, and the
// `.txt.` extension prefix are all wrapped in optional groups so the SDK's
// ExpandPatternIntoOptionalAlternatives expansion covers every combination.
//
// Users supply a `prefix` like "<tenant-id>/data/" or
// "<tenant-id>/fdrv2/aidmaster/"; the SDK prepends the prefix to this layout
// during artifact discovery.
var DefaultBatchLayout = utils.ToStringPointer(
	"(batch=)?%{DATA:batch}/" +
		"(year=%{YEAR:year}/month=%{MONTHNUM:month}/day=%{MONTHDAY:day}/hour=%{HOUR:hour}/)?" +
		"(platform=%{DATA:platform}/)?" +
		"%{DATA}.gz")
