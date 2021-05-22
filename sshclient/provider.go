package sshclient

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"sshclient_run":     resourceRun(),
			"sshclient_scp_put": resourceScpPut(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"sshclient_host":    dataSourceHost(),
			"sshclient_keyscan": dataSourceKeyscan(),
		},
	}
}
