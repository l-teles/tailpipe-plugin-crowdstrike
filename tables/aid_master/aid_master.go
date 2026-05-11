package aid_master

import "github.com/turbot/tailpipe-plugin-sdk/schema"

// AidMaster represents one row in an FDR AIDMaster snapshot — a periodic
// inventory of every agent (host) seen by Falcon, with sensor / OS / hardware
// metadata. All values are strings on the wire so every column is *string.
// Numeric and date-like fields (Time, FirstSeen, AgentTimeOffset, …) hold
// their original textual representation; cast in SQL when needed.
type AidMaster struct {
	schema.CommonFields

	Aid           *string `parquet:"name=aid"`
	Aip           *string `parquet:"name=aip"`
	Cid           *string `parquet:"name=cid"`
	EventPlatform *string `parquet:"name=event_platform"`

	AgentLoadFlags     *string `parquet:"name=agent_load_flags"`
	AgentLocalTime     *string `parquet:"name=agent_local_time"`
	AgentTimeOffset    *string `parquet:"name=agent_time_offset"`
	AgentVersion       *string `parquet:"name=agent_version"`
	BiosManufacturer   *string `parquet:"name=bios_manufacturer"`
	BiosVersion        *string `parquet:"name=bios_version"`
	ChassisType        *string `parquet:"name=chassis_type"`
	City               *string `parquet:"name=city"`
	ComputerName       *string `parquet:"name=computer_name"`
	ConfigBuild        *string `parquet:"name=config_build"`
	ConfigIDBuild      *string `parquet:"name=config_id_build"`
	Continent          *string `parquet:"name=continent"`
	Country            *string `parquet:"name=country"`
	FalconGroupingTags *string `parquet:"name=falcon_grouping_tags"`
	FirstSeen          *string `parquet:"name=first_seen"`
	HostHiddenStatus   *string `parquet:"name=host_hidden_status"`
	MachineDomain      *string `parquet:"name=machine_domain"`
	OU                 *string `parquet:"name=ou"`
	PointerSize        *string `parquet:"name=pointer_size"`
	ProductType        *string `parquet:"name=product_type"`
	SensorGroupingTags *string `parquet:"name=sensor_grouping_tags"`
	ServicePackMajor   *string `parquet:"name=service_pack_major"`
	SiteName           *string `parquet:"name=site_name"`
	SystemManufacturer *string `parquet:"name=system_manufacturer"`
	SystemProductName  *string `parquet:"name=system_product_name"`
	Time               *string `parquet:"name=time"`
	Timezone           *string `parquet:"name=timezone"`
	Version            *string `parquet:"name=version"`

	// Whole record (including the keys promoted above) for forward-compat.
	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func (AidMaster) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"aid":            "Agent (host) identifier.",
		"aip":            "Agent IP address as observed by the sensor.",
		"cid":            "Customer (tenant) identifier.",
		"event_platform": "Operating system family: Win, Mac, Lin, Other.",
		"agent_version":  "Falcon sensor version installed on the host.",
		"computer_name":  "Hostname.",
		"first_seen":     "Epoch seconds (string) when the agent was first observed.",
		"time":           "Epoch seconds (string) when this AIDMaster record was emitted.",
		"version":        "Operating system version (e.g. \"Windows 11\", \"macOS 14\").",
		"machine_domain": "AD / directory domain joined by the host.",
		"payload":        "Full record JSON, including any field not promoted to a typed column.",
	}
}
