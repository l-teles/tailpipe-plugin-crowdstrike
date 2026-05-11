package fdr_event

import (
	"context"
	"fmt"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

// FdrEventExtractor splits a gunzipped FDR primary-events file into one
// FdrEvent per JSON line. The SDK runs this after the gzip loader has fully
// decompressed the artifact, so the input is the entire uncompressed file
// as a single []byte.
type FdrEventExtractor struct{}

func NewFdrEventExtractor() artifact_source.Extractor { return &FdrEventExtractor{} }

func (FdrEventExtractor) Identifier() string { return "fdr_event_extractor" }

func (FdrEventExtractor) Extract(_ context.Context, a any) ([]any, error) {
	raw, ok := a.([]byte)
	if !ok {
		return nil, fmt.Errorf("fdr_event_extractor: expected []byte, got %T", a)
	}
	return common.ExtractJSONLines(raw, "fdr_event_extractor", buildFdrEvent)
}

func buildFdrEvent(doc map[string]any) *FdrEvent {
	evt := &FdrEvent{Payload: doc}

	// Sensor (lowercase) identifiers.
	evt.Aid = common.StringFromMap(doc, "aid")
	evt.Aip = common.StringFromMap(doc, "aip")
	evt.Cid = common.StringFromMap(doc, "cid")
	evt.EventPlatform = common.StringFromMap(doc, "event_platform")
	evt.EventSimpleName = common.StringFromMap(doc, "event_simpleName")
	evt.Name = common.StringFromMap(doc, "name")
	evt.ComputerName = common.StringFromMap(doc, "ComputerName")
	evt.EventOrigin = common.StringFromMap(doc, "EventOrigin")

	// External-API (PascalCase) identifiers.
	evt.EventType = common.StringFromMap(doc, "EventType")
	evt.ExternalApiType = common.StringFromMap(doc, "ExternalApiType")
	evt.AgentIdString = common.StringFromMap(doc, "AgentIdString")
	evt.CustomerIdString = common.StringFromMap(doc, "CustomerIdString")

	// Raw timestamps (units differ across event families — see EnrichRow).
	evt.ContextTimeStamp = common.StringFromMap(doc, "ContextTimeStamp")
	evt.Timestamp = common.StringFromMap(doc, "timestamp")
	evt.UTCTimestamp = common.StringFromMap(doc, "UTCTimestamp")

	// Cross-fill cid from CustomerIdString so it's always populated.
	if evt.Cid == nil && evt.CustomerIdString != nil {
		evt.Cid = evt.CustomerIdString
	}
	return evt
}
