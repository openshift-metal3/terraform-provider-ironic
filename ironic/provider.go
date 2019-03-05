package ironic

import (
	"github.com/gophercloud/gophercloud/openstack/baremetal/noauth"
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider Ironic
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_AUTH_URL", ""),
				Description: descriptions["url"],
			},
			"microversion": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("IRONIC_MICROVERSION", "1.50"),
				Description: descriptions["microversion"],
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"ironic_node_v1": resourceNodeV1(),
			//"ironic_port_v1": resourcePortV1(),
		},
		ConfigureFunc: configureProvider,
	}
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"url":          "The authentication endpoint for Ironic",
		"microversion": "The microversion to use for Ironic",
	}
}

// Creates a noauth Ironic client
// FIXME: Support regular auth Ironic
func configureProvider(schema *schema.ResourceData) (interface{}, error) {
	client, err := noauth.NewBareMetalNoAuth(noauth.EndpointOpts{
		IronicEndpoint: schema.Get("url").(string),
	})
	if err != nil {
		return nil, err
	}

	client.Microversion = schema.Get("microversion").(string)

	return client, err
}
