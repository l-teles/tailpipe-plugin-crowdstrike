package fdr_event

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
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
// `payload->>'$.SomeField'`.
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

	// Timestamps; the wire units differ by event family (see mapFdrEvent).
	ContextTimeStamp *time.Time `parquet:"name=context_time_stamp"`
	Timestamp        *time.Time `parquet:"name=timestamp"`
	UTCTimestamp     *time.Time `parquet:"name=utc_timestamp"`

	// Whole record (including the keys promoted above) for ad-hoc queries.
	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func mapFdrEvent(doc map[string]any) *FdrEvent {
	evt := &FdrEvent{Payload: doc}

	evt.Aid = common.StringFromMap(doc, "aid")
	evt.Aip = common.StringFromMap(doc, "aip")
	evt.Cid = common.StringFromMap(doc, "cid")
	evt.EventPlatform = common.StringFromMap(doc, "event_platform")
	evt.EventSimpleName = common.StringFromMap(doc, "event_simpleName")
	evt.Name = common.StringFromMap(doc, "name")
	evt.ComputerName = common.StringFromMap(doc, "ComputerName")
	evt.EventOrigin = common.StringFromMap(doc, "EventOrigin")

	evt.EventType = common.StringFromMap(doc, "EventType")
	evt.ExternalApiType = common.StringFromMap(doc, "ExternalApiType")
	evt.AgentIdString = common.StringFromMap(doc, "AgentIdString")
	evt.CustomerIdString = common.StringFromMap(doc, "CustomerIdString")

	// ContextTimeStamp: epoch seconds (sensor). UTCTimestamp: epoch ms
	// (external-API). timestamp: epoch ms (sensor) or RFC3339 (external-API).
	evt.ContextTimeStamp = common.EpochSecondsFromMap(doc, "ContextTimeStamp")
	evt.UTCTimestamp = common.EpochMillisFromMap(doc, "UTCTimestamp")
	evt.Timestamp = common.RFC3339OrEpochMillisFromMap(doc, "timestamp")

	// Cross-fill cid from CustomerIdString so it's always populated.
	if evt.Cid == nil && evt.CustomerIdString != nil {
		evt.Cid = evt.CustomerIdString
	}
	return evt
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
		"context_time_stamp": "Sensor event time (`ContextTimeStamp`, delivered as epoch seconds).",
		"timestamp":          "Record `timestamp` (delivered as epoch ms for sensor events, RFC3339 for external-API events).",
		"utc_timestamp":      "External-API event time (`UTCTimestamp`, delivered as epoch ms).",
		"payload":            "Full event JSON, including any field not promoted to a typed column.",
		"tp_timestamp":       "Event time: the first of context_time_stamp, utc_timestamp and timestamp that is present.",
		"tp_source_ip":       "Agent IP (`aip`) where present.",
	}
}
