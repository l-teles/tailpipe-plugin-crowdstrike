package fdr_event

import (
	"github.com/turbot/tailpipe-plugin-sdk/schema"
)

// FdrEvent represents a single record from a Falcon Data Replicator (FDR)
// primary-events file. FDR primary streams interleave two flavours:
//
//  1. Sensor telemetry — flat records keyed by `event_simpleName` (e.g.
//     ProcessRollup2, EndOfProcess, DnsRequest), with `aid`, `aip`, `cid`,
//     `event_platform`, `name`, `ContextTimeStamp`.
//  2. External-API events — wrapped records with `EventType` =
//     "Event_ExternalApiEvent", a more specific `ExternalApiType`, and
//     PascalCase identifiers (`AgentIdString`, `CustomerIdString`,
//     `UTCTimestamp`).
//
// Hot identifiers common to investigations are typed columns. The full record
// is preserved in `payload` (JSON) so any uncommon field remains queryable via
// `payload->>'$.SomeField'`. Every value in FDR JSON is delivered as a string
// (timestamps included), so all string fields here are `*string`.
type FdrEvent struct {
	schema.CommonFields

	// Sensor identifiers (lowercase keys on the wire).
	Aid             *string `parquet:"name=aid"`
	Aip             *string `parquet:"name=aip"`
	Cid             *string `parquet:"name=cid"`
	EventPlatform   *string `parquet:"name=event_platform"`
	EventSimpleName *string `parquet:"name=event_simple_name"`
	Name            *string `parquet:"name=name"`
	ComputerName    *string `parquet:"name=computer_name"`
	EventOrigin     *string `parquet:"name=event_origin"`

	// External-API identifiers (PascalCase keys on the wire).
	EventType        *string `parquet:"name=event_type"`
	ExternalApiType  *string `parquet:"name=external_api_type"`
	AgentIdString    *string `parquet:"name=agent_id_string"`
	CustomerIdString *string `parquet:"name=customer_id_string"`

	// Raw timestamp strings (units vary by event family — see EnrichRow).
	ContextTimeStamp *string `parquet:"name=context_time_stamp"`
	Timestamp        *string `parquet:"name=timestamp_raw"`
	UTCTimestamp     *string `parquet:"name=utc_timestamp_raw"`

	// Whole record (including the keys promoted above) for ad-hoc queries.
	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func (FdrEvent) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"aid":                "Agent (host) identifier — sensor events only.",
		"aip":                "Agent IP address as observed by the sensor.",
		"cid":                "Customer identifier (CrowdStrike tenant ID).",
		"event_platform":     "Operating system family: Win, Mac, Lin, Other.",
		"event_simple_name":  "Sensor event short name (e.g. ProcessRollup2, EndOfProcess).",
		"name":               "Versioned sensor event name (e.g. EndOfProcessV15).",
		"computer_name":      "Hostname as known to the sensor.",
		"event_origin":       "Internal sensor event-origin code.",
		"event_type":         "External-API event family (e.g. Event_ExternalApiEvent).",
		"external_api_type":  "External-API event subtype (e.g. Event_ModuleSummaryInfoEvent).",
		"agent_id_string":    "Agent identifier — external-API events.",
		"customer_id_string": "Customer identifier — external-API events.",
		"context_time_stamp": "Sensor-event timestamp (epoch seconds, may include fractional milliseconds).",
		"timestamp_raw":      "Raw `timestamp` field as delivered (epoch ms or RFC3339, depending on event family).",
		"utc_timestamp_raw":  "Raw `UTCTimestamp` field as delivered — external-API events (epoch ms).",
		"payload":            "Full event JSON, including any field not promoted to a typed column.",
		"tp_timestamp":       "Best-effort event time, parsed from the most-specific timestamp present on the record.",
		"tp_index":           "Customer (tenant) ID.",
		"tp_source_ip":       "Agent IP (`aip`) where present.",
	}
}
