package aid_master

import (
	"context"
	"fmt"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

type AidMasterExtractor struct{}

func NewAidMasterExtractor() artifact_source.Extractor { return &AidMasterExtractor{} }

func (AidMasterExtractor) Identifier() string { return "aid_master_extractor" }

func (AidMasterExtractor) Extract(_ context.Context, a any) ([]any, error) {
	raw, ok := a.([]byte)
	if !ok {
		return nil, fmt.Errorf("aid_master_extractor: expected []byte, got %T", a)
	}
	return common.ExtractJSONLines(raw, "aid_master_extractor", buildAidMaster)
}

func buildAidMaster(doc map[string]any) *AidMaster {
	r := &AidMaster{Payload: doc}

	r.Aid = common.StringFromMap(doc, "aid")
	r.Aip = common.StringFromMap(doc, "aip")
	r.Cid = common.StringFromMap(doc, "cid")
	r.EventPlatform = common.StringFromMap(doc, "event_platform")

	r.AgentLoadFlags = common.StringFromMap(doc, "AgentLoadFlags")
	r.AgentLocalTime = common.StringFromMap(doc, "AgentLocalTime")
	r.AgentTimeOffset = common.StringFromMap(doc, "AgentTimeOffset")
	r.AgentVersion = common.StringFromMap(doc, "AgentVersion")
	r.BiosManufacturer = common.StringFromMap(doc, "BiosManufacturer")
	r.BiosVersion = common.StringFromMap(doc, "BiosVersion")
	r.ChassisType = common.StringFromMap(doc, "ChassisType")
	r.City = common.StringFromMap(doc, "City")
	r.ComputerName = common.StringFromMap(doc, "ComputerName")
	r.ConfigBuild = common.StringFromMap(doc, "ConfigBuild")
	r.ConfigIDBuild = common.StringFromMap(doc, "ConfigIDBuild")
	r.Continent = common.StringFromMap(doc, "Continent")
	r.Country = common.StringFromMap(doc, "Country")
	r.FalconGroupingTags = common.StringFromMap(doc, "FalconGroupingTags")
	r.FirstSeen = common.StringFromMap(doc, "FirstSeen")
	r.HostHiddenStatus = common.StringFromMap(doc, "HostHiddenStatus")
	r.MachineDomain = common.StringFromMap(doc, "MachineDomain")
	r.OU = common.StringFromMap(doc, "OU")
	r.PointerSize = common.StringFromMap(doc, "PointerSize")
	r.ProductType = common.StringFromMap(doc, "ProductType")
	r.SensorGroupingTags = common.StringFromMap(doc, "SensorGroupingTags")
	r.ServicePackMajor = common.StringFromMap(doc, "ServicePackMajor")
	r.SiteName = common.StringFromMap(doc, "SiteName")
	r.SystemManufacturer = common.StringFromMap(doc, "SystemManufacturer")
	r.SystemProductName = common.StringFromMap(doc, "SystemProductName")
	r.Time = common.StringFromMap(doc, "Time")
	r.Timezone = common.StringFromMap(doc, "Timezone")
	r.Version = common.StringFromMap(doc, "Version")

	return r
}
