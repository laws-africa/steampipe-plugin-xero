package main

import (
	"github.com/laws-africa/steampipe-plugin-xero/xero"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{PluginFunc: xero.Plugin})
}
