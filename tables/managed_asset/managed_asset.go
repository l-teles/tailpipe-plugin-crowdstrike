package managed_asset

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/tables/common"
)

// ManagedAsset represents one row in an FDR ManagedAssets snapshot — network
// interface and gateway info per Falcon-managed agent. One row per (agent,
// interface) tuple.
type ManagedAsset struct {
	schema.CommonFields

	Aid  *string    `parquet:"name=aid"`
	Cid  *string    `parquet:"name=cid"`
	Time *time.Time `parquet:"name=time"` // `_time` on the wire

	GatewayIP            *string `parquet:"name=gateway_ip"`
	GatewayMAC           *string `parquet:"name=gateway_mac"`
	InterfaceAlias       *string `parquet:"name=interface_alias"`
	InterfaceDescription *string `parquet:"name=interface_description"`
	LocalAddressIP4      *string `parquet:"name=local_address_ip4"`
	MAC                  *string `parquet:"name=mac"`
	MACPrefix            *string `parquet:"name=mac_prefix"`

	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func mapManagedAsset(doc map[string]any) *ManagedAsset {
	r := &ManagedAsset{Payload: doc}

	r.Aid = common.StringFromMap(doc, "aid")
	r.Cid = common.StringFromMap(doc, "cid")
	r.Time = common.EpochSecondsFromMap(doc, "_time")

	r.GatewayIP = common.StringFromMap(doc, "GatewayIP")
	r.GatewayMAC = common.StringFromMap(doc, "GatewayMAC")
	r.InterfaceAlias = common.StringFromMap(doc, "InterfaceAlias")
	r.InterfaceDescription = common.StringFromMap(doc, "InterfaceDescription")
	r.LocalAddressIP4 = common.StringFromMap(doc, "LocalAddressIP4")
	r.MAC = common.StringFromMap(doc, "MAC")
	r.MACPrefix = common.StringFromMap(doc, "MACPrefix")

	return r
}

func (ManagedAsset) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"aid":                   "Agent (host) identifier.",
		"cid":                   "Customer (tenant) identifier.",
		"time":                  "When the record was emitted (delivered as `_time`).",
		"gateway_ip":            "Default gateway IP for the interface.",
		"gateway_mac":           "Default gateway MAC address.",
		"interface_alias":       "OS-level interface alias (e.g. en0, Ethernet 2).",
		"interface_description": "OS-level interface description.",
		"local_address_ip4":     "Local IPv4 address assigned to the interface.",
		"mac":                   "Interface MAC address.",
		"mac_prefix":            "First three octets of the interface MAC address (OUI).",
		"payload":               "Full record JSON, including any field not promoted to a typed column.",
		"tp_timestamp":          "Record time (`time`).",
	}
}
