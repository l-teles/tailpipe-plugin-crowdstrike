package app_info

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

// AppInfo represents one row in an FDR AppInfo snapshot — the application
// inventory observed on a host. One row per (agent, application) tuple.
type AppInfo struct {
	schema.CommonFields

	Aid        *string `parquet:"name=aid"`
	Cid        *string `parquet:"name=cid"`
	Hostname   *string `parquet:"name=hostname"`
	ExternalIP *string `parquet:"name=external_ip"`

	CompanyName           *string    `parquet:"name=company_name"`
	FileName              *string    `parquet:"name=file_name"`
	FileVersion           *string    `parquet:"name=file_version"`
	ProductName           *string    `parquet:"name=product_name"`
	ProductVersion        *string    `parquet:"name=product_version"`
	SHA256HashData        *string    `parquet:"name=sha256_hash_data"`
	DetectionCount        *int64     `parquet:"name=detection_count"`
	InstallationTimestamp *time.Time `parquet:"name=installation_timestamp"`
	SoftwareType          *string    `parquet:"name=software_type"`
	Category              *string    `parquet:"name=category"`
	Time                  *time.Time `parquet:"name=time"` // `_time` on the wire

	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func mapAppInfo(doc map[string]any) *AppInfo {
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
	r.DetectionCount = common.IntFromMap(doc, "detectionCount")
	r.InstallationTimestamp = common.EpochSecondsFromMap(doc, "installationTimestamp")
	r.SoftwareType = common.StringFromMap(doc, "SoftwareType")
	r.Category = common.StringFromMap(doc, "Category")
	r.Time = common.EpochSecondsFromMap(doc, "_time")

	return r
}

func (AppInfo) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"aid":                    "Agent (host) identifier.",
		"cid":                    "Customer (tenant) identifier.",
		"hostname":               "Hostname (lowercase as delivered).",
		"external_ip":            "Host's external IP address as observed by Falcon.",
		"company_name":           "Application vendor / company name.",
		"file_name":              "Executable filename (e.g. \"node\").",
		"file_version":           "File version of the executable.",
		"product_name":           "Product display name.",
		"product_version":        "Product version string.",
		"sha256_hash_data":       "SHA-256 of the executable.",
		"detection_count":        "Number of detections Falcon reports for the application.",
		"installation_timestamp": "When the application was installed (null if unknown).",
		"software_type":          "Broad classification (e.g. application, driver).",
		"category":               "Application category assigned by Falcon.",
		"time":                   "When this AppInfo record was emitted (delivered as `_time`).",
		"payload":                "Full record JSON, including any field not promoted to a typed column.",
		"tp_timestamp":           "Record time (`time`).",
	}
}
