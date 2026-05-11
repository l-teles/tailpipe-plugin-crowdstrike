package crowdstrike

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/crowdstrikeconfig"
	"github.com/l-teles/tailpipe-plugin-crowdstrike/sources/s3_bucket"
	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/aid_master"
	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/app_info"
	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/fdr_event"
	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/managed_assets"
	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/user_info"
)

type Plugin struct {
	plugin.PluginImpl
}

func init() {
	// Register tables. Type parameters are: row struct, table impl.
	table.RegisterTable[*fdr_event.FdrEvent, *fdr_event.FdrEventTable]()
	table.RegisterTable[*aid_master.AidMaster, *aid_master.AidMasterTable]()
	table.RegisterTable[*app_info.AppInfo, *app_info.AppInfoTable]()
	table.RegisterTable[*managed_assets.ManagedAsset, *managed_assets.ManagedAssetsTable]()
	table.RegisterTable[*user_info.UserInfo, *user_info.UserInfoTable]()

	// The remote S3 source is plugin-specific. The SDK ships the local-file
	// source as constants.ArtifactSourceIdentifier; tables register against it
	// in their GetSourceMetadata().
	row_source.RegisterRowSource[*s3_bucket.CrowdstrikeS3BucketSource]()
}

func NewPlugin() (_ plugin.TailpipePlugin, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	return &Plugin{PluginImpl: plugin.NewPluginImpl(crowdstrikeconfig.PluginName)}, nil
}
