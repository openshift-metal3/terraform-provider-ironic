package ironic

import (
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/baremetal/noauth"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider Ironic
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("IRONIC_ENDPOINT", ""),
				Description: descriptions["url"],
			},
			"microversion": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("IRONIC_MICROVERSION", "1.52"),
				Description: descriptions["microversion"],
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"ironic_node_v1":       resourceNodeV1(),
			"ironic_port_v1":       resourcePortV1(),
			"ironic_allocation_v1": resourceAllocationV1(),
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
func configureProvider(schema *schema.ResourceData) (interface{}, error) {
	url := schema.Get("url").(string)
	if url == "" {
		return nil, fmt.Errorf("url is required for ironic provider")
	}

	client, err := noauth.NewBareMetalNoAuth(noauth.EndpointOpts{
		IronicEndpoint: url,
	})
	if err != nil {
		return nil, err
	}

	client.Microversion = schema.Get("microversion").(string)

	return client, err
}
