package ironic

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccIronicNode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResource,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNodeExists("ironic_node_v1.node-0"),
				),
			},
		},
	})
}

func testAccCheckNodeExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*gophercloud.ServiceClient)

		rs, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no node ID is set")
		}

		_, err := nodes.Get(client, rs.Primary.ID).Extract()
		if err != nil {
			return fmt.Errorf("node (%s) not found: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccNodeDestroy(state *terraform.State) error {
	client := testAccProvider.Meta().(*gophercloud.ServiceClient)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "ironic_node_v1" {
			continue
		}

		_, err := nodes.Get(client, rs.Primary.ID).Extract()
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return fmt.Errorf("unexpected error: %s, expected 404", err)
		}
	}

	return nil
}

const (
	testAccNodeResource = `
		resource "ironic_node_v1" "node-0" {
			name = "node-0"
			driver = "fake-hardware"
			boot_interface = "pxe"
			driver_info = {
				ipmi_port      = "6230"
				ipmi_username  = "admin"
				deploy_kernel  = "http://172.22.0.1/images/tinyipa-stable-rocky.vmlinuz"
				ipmi_address   = "192.168.122.1"
				deploy_ramdisk = "http://172.22.0.1/images/tinyipa-stable-rocky.gz"
				ipmi_password  = "admin"
			}
		}`
)
