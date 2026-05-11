package main

import (
	"log/slog"
	"os"

	"github.com/l-teles/tailpipe-plugin-crowdstrike/crowdstrike"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "metadata" {
		os.Exit(plugin.PrintMetadata(crowdstrike.NewPlugin))
	}

	if err := plugin.Serve(&plugin.ServeOpts{
		PluginFunc: crowdstrike.NewPlugin,
	}); err != nil {
		slog.Error("error starting crowdstrike plugin", "error", err)
	}
}
