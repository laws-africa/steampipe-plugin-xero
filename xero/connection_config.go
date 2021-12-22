package xero

import (
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/schema"
)

type xeroConfig struct {
	ClientId     *string `cty:"client_id"`
	ClientSecret *string `cty:"client_secret"`
	TenantName   *string `cty:"tenant_name"`
	OauthCode    *string `cty:"oauth_code"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"client_id": {
		Type: schema.TypeString,
	},
	"client_secret": {
		Type: schema.TypeString,
	},
	"tenant_name": {
		Type: schema.TypeString,
	},
	"oauth_code": {
		Type: schema.TypeString,
	},
}

func ConfigInstance() interface{} {
	return &xeroConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) xeroConfig {
	if connection == nil || connection.Config == nil {
		return xeroConfig{}
	}
	config, _ := connection.Config.(xeroConfig)
	return config
}
