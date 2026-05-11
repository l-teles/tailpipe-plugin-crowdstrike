package app_info

import "github.com/turbot/tailpipe-plugin-sdk/schema"

// AppInfo represents one row in an FDR AppInfo snapshot — the application
// inventory observed on a host. One row per (agent, application) tuple.
type AppInfo struct {
	schema.CommonFields

	Aid        *string `parquet:"name=aid"`
	Cid        *string `parquet:"name=cid"`
	Hostname   *string `parquet:"name=hostname"`
	ExternalIP *string `parquet:"name=external_ip"`

	CompanyName           *string `parquet:"name=company_name"`
	FileName              *string `parquet:"name=file_name"`
	FileVersion           *string `parquet:"name=file_version"`
	ProductName           *string `parquet:"name=product_name"`
	ProductVersion        *string `parquet:"name=product_version"`
	SHA256HashData        *string `parquet:"name=sha256_hash_data"`
	DetectionCount        *string `parquet:"name=detection_count"`
	InstallationTimestamp *string `parquet:"name=installation_timestamp"`
	SoftwareType          *string `parquet:"name=software_type"`
	Category              *string `parquet:"name=category"`
	Time                  *string `parquet:"name=time"` // raw `_time` field on the wire

	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func (AppInfo) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"aid":                    "Agent (host) identifier.",
		"cid":                    "Customer (tenant) identifier.",
		"hostname":               "Hostname (lowercase as delivered).",
		"external_ip":            "Host's external IP address as observed by Falcon.",
		"company_name":           "Application vendor / company name.",
		"file_name":              "Executable filename (e.g. \"node\").",
		"product_name":           "Product display name.",
		"product_version":        "Product version string.",
		"sha256_hash_data":       "SHA-256 of the executable.",
		"installation_timestamp": "Epoch seconds when the application was installed (\"0\" if unknown).",
		"software_type":          "Broad classification (e.g. application, driver).",
		"time":                   "Epoch seconds (string) when this AppInfo record was emitted (delivered as `_time`).",
		"payload":                "Full record JSON, including any field not promoted to a typed column.",
	}
}
