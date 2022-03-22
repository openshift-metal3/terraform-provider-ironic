// +build acceptance

package ironic

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccIronicNode(t *testing.T) {
	var node nodes.Node

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccNodeDestroy,
		Steps: []resource.TestStep{

			// Create a node and check that it exists
			{
				Config: testAccNodeResource(""),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "enroll",
					),
				),
			},

			// Ensure node is manageable
			{
				Config: testAccNodeResource("manage = true"),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "manageable"),
				),
			},

			// Inspect the node
			{
				Config: testAccNodeResource("inspect = true"),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "manageable"),
				),
			},

			// Clean the node
			{
				Config: testAccNodeResource("clean = true"),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"provision_state", "manageable"),
				),
			},

			// Change the node's power state to 'power on', with a timeout
			{
				Config: testAccNodeResource(`
					target_power_state = "power on"
					power_state_timeout = 10
				`),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"power_state", "power on"),
				),
			},

			// Change the node's power state to 'power off'.
			{
				Config: testAccNodeResource("target_power_state = \"power off\""),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"power_state", "power off"),
				),
			},

			// Change the node's power state to 'rebooting', it probably
			// doesn't make a whole lot of sense for a terraform user to
			// declare a node's state as forever rebooting, as it'd reboot
			// every time, but we should check anyway that if they do say
			// rebooting, power_state goes to power_on and terraform exits
			// successfully.
			{
				Config: testAccNodeResource("target_power_state = \"rebooting\""),
				Check: resource.ComposeTestCheckFunc(
					CheckNodeExists("ironic_node_v1.node-0", &node),
					resource.TestCheckResourceAttr("ironic_node_v1.node-0",
						"power_state", "power on"),
				),
			},
		},
	})
}

func CheckNodeExists(name string, node *nodes.Node) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client, err := testAccProvider.Meta().(*Clients).GetIronicClient()
		if err != nil {
			return err
		}

		rs, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no node ID is set")
		}

		result, err := nodes.Get(client, rs.Primary.ID).Extract()
		if err != nil {
			return fmt.Errorf("node (%s) not found: %s", rs.Primary.ID, err)
		}

		*node = *result

		return nil
	}
}

func testAccNodeDestroy(state *terraform.State) error {
	client, err := testAccProvider.Meta().(*Clients).GetIronicClient()
	if err != nil {
		return err
	}

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

func testAccNodeResource(extraValue string) string {
	return fmt.Sprintf(`resource "ironic_node_v1" "node-0" {
			name = "node-0"
			driver = "fake-hardware"

			boot_interface = "pxe"
			deploy_interface = "fake"
			inspect_interface = "fake"
			management_interface = "fake"
			power_interface = "fake"
			resource_class = "baremetal"
			vendor_interface = "no-vendor"

			driver_info = {
				ipmi_port      = "6230"
				ipmi_username  = "admin"
				ipmi_password  = "admin"
			}

			%s
		}`, extraValue)
}

func TestBuildManualCleaningSteps(t *testing.T) {
	cases := []struct {
		Scenario      string
		RAIDInterface string
		RAIDConfig    string
		BIOSSettings  string
		Expected      []nodes.CleanStep
		ExpectedError bool
	}{
		{
			Scenario:      "no raid and no bios",
			RAIDInterface: "no-raid",
			RAIDConfig:    "",
			BIOSSettings:  "",
			Expected:      nil,
		},
		{
			Scenario:      "just raid",
			RAIDInterface: "irmc",
			RAIDConfig:    "{\"hardwareRAIDVolumes\":[{\"level\":\"0\",\"name\":\"raid0\"}],\"softwareRAIDVolumes\":null}",
			BIOSSettings:  "",
			Expected: []nodes.CleanStep{
				{
					Interface: "raid",
					Step:      "delete_configuration",
				},
				{
					Interface: "raid",
					Step:      "create_configuration",
				},
			},
		},
		{
			Scenario:      "just bios",
			RAIDInterface: "irmc",
			RAIDConfig:    "",
			BIOSSettings:  "[{\"name\":\"cpu_vt_enabled\",\"value\":\"False\"},{\"name\":\"hyper_threading_enabled\",\"value\":\"True\"}]",
			Expected: []nodes.CleanStep{
				{
					Interface: "bios",
					Step:      "apply_configuration",
					Args: map[string]interface{}{
						"settings": []map[string]string{
							{
								"name":  "cpu_vt_enabled",
								"value": "False",
							},
							{
								"name":  "hyper_threading_enabled",
								"value": "True",
							},
						},
					},
				},
			},
		},
		{
			Scenario:      "raid and bios",
			RAIDInterface: "irmc",
			RAIDConfig:    "{\"hardwareRAIDVolumes\":[{\"level\":\"0\",\"name\":\"raid0\"}],\"softwareRAIDVolumes\":null}",
			BIOSSettings:  "[{\"name\":\"cpu_vt_enabled\",\"value\":\"False\"},{\"name\":\"hyper_threading_enabled\",\"value\":\"True\"}]",
			Expected: []nodes.CleanStep{
				{
					Interface: "raid",
					Step:      "delete_configuration",
				},
				{
					Interface: "raid",
					Step:      "create_configuration",
				},
				{
					Interface: "bios",
					Step:      "apply_configuration",
					Args: map[string]interface{}{
						"settings": []map[string]string{
							{
								"name":  "cpu_vt_enabled",
								"value": "False",
							},
							{
								"name":  "hyper_threading_enabled",
								"value": "True",
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Scenario, func(t *testing.T) {
			step, err := buildManualCleaningSteps(c.RAIDInterface, c.RAIDConfig, c.BIOSSettings)
			if !reflect.DeepEqual(c.Expected, step) {
				t.Errorf("expected: %v, got: %v", c.Expected, step)
			}
			if (err != nil) != c.ExpectedError {
				t.Errorf("got unexpected error: %v", err)
			}
		})
	}
}
