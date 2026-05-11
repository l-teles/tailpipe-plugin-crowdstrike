package managed_assets

import (
	"context"
	"fmt"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

type ManagedAssetsExtractor struct{}

func NewManagedAssetsExtractor() artifact_source.Extractor { return &ManagedAssetsExtractor{} }

func (ManagedAssetsExtractor) Identifier() string { return "managed_assets_extractor" }

func (ManagedAssetsExtractor) Extract(_ context.Context, a any) ([]any, error) {
	raw, ok := a.([]byte)
	if !ok {
		return nil, fmt.Errorf("managed_assets_extractor: expected []byte, got %T", a)
	}
	return common.ExtractJSONLines(raw, "managed_assets_extractor", buildManagedAsset)
}

func buildManagedAsset(doc map[string]any) *ManagedAsset {
	r := &ManagedAsset{Payload: doc}

	r.Aid = common.StringFromMap(doc, "aid")
	r.Cid = common.StringFromMap(doc, "cid")
	r.Time = common.StringFromMap(doc, "_time")

	r.GatewayIP = common.StringFromMap(doc, "GatewayIP")
	r.GatewayMAC = common.StringFromMap(doc, "GatewayMAC")
	r.InterfaceAlias = common.StringFromMap(doc, "InterfaceAlias")
	r.InterfaceDescription = common.StringFromMap(doc, "InterfaceDescription")
	r.LocalAddressIP4 = common.StringFromMap(doc, "LocalAddressIP4")
	r.MAC = common.StringFromMap(doc, "MAC")
	r.MACPrefix = common.StringFromMap(doc, "MACPrefix")

	return r
}
