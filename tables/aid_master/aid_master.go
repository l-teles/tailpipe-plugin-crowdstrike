package aid_master

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

// AidMaster represents one row in an FDR AIDMaster snapshot — a periodic
// inventory of every agent (host) seen by Falcon, with sensor / OS / hardware
// metadata. Timestamps are typed; other values keep their wire text.
type AidMaster struct {
	schema.CommonFields

	Aid           *string `parquet:"name=aid"`
	Aip           *string `parquet:"name=aip"`
	Cid           *string `parquet:"name=cid"`
	EventPlatform *string `parquet:"name=event_platform"`

	AgentLoadFlags     *string    `parquet:"name=agent_load_flags"`
	AgentLocalTime     *string    `parquet:"name=agent_local_time"`
	AgentTimeOffset    *string    `parquet:"name=agent_time_offset"`
	AgentVersion       *string    `parquet:"name=agent_version"`
	BiosManufacturer   *string    `parquet:"name=bios_manufacturer"`
	BiosVersion        *string    `parquet:"name=bios_version"`
	ChassisType        *string    `parquet:"name=chassis_type"`
	City               *string    `parquet:"name=city"`
	ComputerName       *string    `parquet:"name=computer_name"`
	ConfigBuild        *string    `parquet:"name=config_build"`
	ConfigIDBuild      *string    `parquet:"name=config_id_build"`
	Continent          *string    `parquet:"name=continent"`
	Country            *string    `parquet:"name=country"`
	FalconGroupingTags *string    `parquet:"name=falcon_grouping_tags"`
	FirstSeen          *time.Time `parquet:"name=first_seen"`
	HostHiddenStatus   *string    `parquet:"name=host_hidden_status"`
	MachineDomain      *string    `parquet:"name=machine_domain"`
	OU                 *string    `parquet:"name=ou"`
	PointerSize        *string    `parquet:"name=pointer_size"`
	ProductType        *string    `parquet:"name=product_type"`
	SensorGroupingTags *string    `parquet:"name=sensor_grouping_tags"`
	ServicePackMajor   *string    `parquet:"name=service_pack_major"`
	SiteName           *string    `parquet:"name=site_name"`
	SystemManufacturer *string    `parquet:"name=system_manufacturer"`
	SystemProductName  *string    `parquet:"name=system_product_name"`
	Time               *time.Time `parquet:"name=time"`
	Timezone           *string    `parquet:"name=timezone"`
	Version            *string    `parquet:"name=version"`

	// Whole record (including the keys promoted above) for forward-compat.
	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func mapAidMaster(doc map[string]any) *AidMaster {
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
	r.FirstSeen = common.EpochSecondsFromMap(doc, "FirstSeen")
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
	r.Time = common.EpochSecondsFromMap(doc, "Time")
	r.Timezone = common.StringFromMap(doc, "Timezone")
	r.Version = common.StringFromMap(doc, "Version")

	return r
}

func (AidMaster) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"aid":                  "Agent (host) identifier.",
		"aip":                  "Agent IP address as observed by the sensor.",
		"cid":                  "Customer (tenant) identifier.",
		"event_platform":       "Operating system family: Win, Mac, Lin, Other.",
		"agent_load_flags":     "Sensor load flags reported by the agent, as delivered.",
		"agent_local_time":     "Host's local clock when the record was produced (epoch seconds, as delivered).",
		"agent_time_offset":    "Clock offset reported for the agent, as delivered.",
		"agent_version":        "Falcon sensor version installed on the host.",
		"bios_manufacturer":    "BIOS manufacturer.",
		"bios_version":         "BIOS version.",
		"chassis_type":         "Hardware chassis type reported by the host.",
		"city":                 "City of the host's location, as reported by Falcon.",
		"computer_name":        "Hostname.",
		"config_build":         "Sensor configuration build.",
		"config_id_build":      "Sensor configuration ID build number.",
		"continent":            "Continent of the host's location, as reported by Falcon.",
		"country":              "Country of the host's location, as reported by Falcon.",
		"falcon_grouping_tags": "Falcon grouping tags assigned to the host from the console or API.",
		"first_seen":           "When the agent was first observed.",
		"host_hidden_status":   "Whether the host is hidden in the Falcon console (e.g. visible).",
		"machine_domain":       "AD / directory domain joined by the host.",
		"ou":                   "Active Directory organizational unit of the host.",
		"pointer_size":         "OS pointer size in bytes (e.g. 8 for 64-bit).",
		"product_type":         "Windows product type code (1 = workstation, 2 = domain controller, 3 = server).",
		"sensor_grouping_tags": "Grouping tags set on the sensor at install time.",
		"service_pack_major":   "Major service pack version of the operating system.",
		"site_name":            "Active Directory site of the host.",
		"system_manufacturer":  "System (hardware) manufacturer.",
		"system_product_name":  "System (hardware) model name.",
		"time":                 "When this AIDMaster record was emitted.",
		"timezone":             "Host's configured time zone.",
		"version":              "Operating system version (e.g. \"Windows 11\", \"macOS 14\").",
		"payload":              "Full record JSON, including any field not promoted to a typed column.",
		"tp_timestamp":         "Record time: `time`, or `first_seen` when `time` is absent.",
	}
}
