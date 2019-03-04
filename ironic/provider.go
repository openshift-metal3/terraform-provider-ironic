package ironic

import (
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider Ironic
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{},
	}
}
