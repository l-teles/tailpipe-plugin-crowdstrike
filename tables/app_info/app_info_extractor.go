package app_info

import (
	"context"
	"fmt"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

type AppInfoExtractor struct{}

func NewAppInfoExtractor() artifact_source.Extractor { return &AppInfoExtractor{} }

func (AppInfoExtractor) Identifier() string { return "app_info_extractor" }

func (AppInfoExtractor) Extract(_ context.Context, a any) ([]any, error) {
	raw, ok := a.([]byte)
	if !ok {
		return nil, fmt.Errorf("app_info_extractor: expected []byte, got %T", a)
	}
	return common.ExtractJSONLines(raw, "app_info_extractor", buildAppInfo)
}

func buildAppInfo(doc map[string]any) *AppInfo {
	r := &AppInfo{Payload: doc}

	r.Aid = common.StringFromMap(doc, "aid")
	r.Cid = common.StringFromMap(doc, "cid")
	r.Hostname = common.StringFromMap(doc, "hostname")
	r.ExternalIP = common.StringFromMap(doc, "externalIP")

	r.CompanyName = common.StringFromMap(doc, "CompanyName")
	r.FileName = common.StringFromMap(doc, "FileName")
	r.FileVersion = common.StringFromMap(doc, "FileVersion")
	r.ProductName = common.StringFromMap(doc, "ProductName")
	r.ProductVersion = common.StringFromMap(doc, "ProductVersion")
	r.SHA256HashData = common.StringFromMap(doc, "SHA256HashData")
	r.DetectionCount = common.StringFromMap(doc, "detectionCount")
	r.InstallationTimestamp = common.StringFromMap(doc, "installationTimestamp")
	r.SoftwareType = common.StringFromMap(doc, "SoftwareType")
	r.Category = common.StringFromMap(doc, "Category")
	r.Time = common.StringFromMap(doc, "_time")

	return r
}
