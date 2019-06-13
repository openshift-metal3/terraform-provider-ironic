// +build acceptance

package ironic

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	th "github.com/openshift-metal3/terraform-provider-ironic/testhelper"
)

// Creates a node, and an allocation that should use it
func TestAccIronicDeployment(t *testing.T) {
	var node nodes.Node

	nodeName := th.RandomString("TerraformACC-Node-", 8)
	allocationName := th.RandomString("TerraformACC-Allocation-", 8)
	resourceClass := th.RandomString("baremetal-", 8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			// Create a test deployment
			{
				Config: testAccDeploymentResource(nodeName, resourceClass, allocationName),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1."+nodeName, &node),
					resource.TestCheckResourceAttr("ironic_deployment."+nodeName, "provision_state", "active"),
				),
			},
		},
	})
}

func testAccDeploymentDestroy(state *terraform.State) error {
	client := testAccProvider.Meta().(Clients).Ironic

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

func testAccDeploymentResource(node, resourceClass, allocation string) string {
	return fmt.Sprintf(`
		resource "ironic_node_v1" "%s" {
			name = "%s"
			driver = "fake-hardware"
			available = true
			target_power_state = "power off"

			boot_interface = "fake"
			deploy_interface = "fake"
			management_interface = "fake"
			power_interface = "fake"
			resource_class = "%s"
			vendor_interface = "no-vendor"
		}

		resource "ironic_allocation_v1" "%s" {
			name = "%s"
			resource_class = "%s"
			candidate_nodes = [
				"${ironic_node_v1.%s.id}"
			]
		}

		resource "ironic_deployment" "%s" {
			name = "%s"
			node_uuid = "${ironic_allocation_v1.%s.node_uuid}"

			instance_info = {
				image_source   = "http://172.22.0.1/images/redhat-coreos-maipo-latest.qcow2"
				image_checksum = "26c53f3beca4e0b02e09d335257826fd"
				root_gb = "25"
			}

			user_data = "asdf"
		}

`, node, node, resourceClass, allocation, allocation, resourceClass, node, node, node, allocation)
}
