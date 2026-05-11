package managed_assets

import "github.com/turbot/tailpipe-plugin-sdk/schema"

// ManagedAsset represents one row in an FDR ManagedAssets snapshot — network
// interface and gateway info per Falcon-managed agent. One row per (agent,
// interface) tuple.
type ManagedAsset struct {
	schema.CommonFields

	Aid  *string `parquet:"name=aid"`
	Cid  *string `parquet:"name=cid"`
	Time *string `parquet:"name=time"` // raw `_time` field on the wire

	GatewayIP            *string `parquet:"name=gateway_ip"`
	GatewayMAC           *string `parquet:"name=gateway_mac"`
	InterfaceAlias       *string `parquet:"name=interface_alias"`
	InterfaceDescription *string `parquet:"name=interface_description"`
	LocalAddressIP4      *string `parquet:"name=local_address_ip4"`
	MAC                  *string `parquet:"name=mac"`
	MACPrefix            *string `parquet:"name=mac_prefix"`

	Payload map[string]any `parquet:"name=payload, type=JSON"`
}

func (ManagedAsset) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"aid":                   "Agent (host) identifier.",
		"cid":                   "Customer (tenant) identifier.",
		"time":                  "Epoch seconds (string) when the record was emitted (delivered as `_time`).",
		"gateway_ip":            "Default gateway IP for the interface.",
		"gateway_mac":           "Default gateway MAC address.",
		"interface_alias":       "OS-level interface alias (e.g. en0, Ethernet 2).",
		"interface_description": "OS-level interface description.",
		"local_address_ip4":     "Local IPv4 address assigned to the interface.",
		"mac":                   "Interface MAC address.",
		"mac_prefix":            "First three octets of the interface MAC address (OUI).",
		"payload":               "Full record JSON, including any field not promoted to a typed column.",
	}
}
