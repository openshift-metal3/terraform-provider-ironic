package ironic

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceNodeV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceNodeV1Create,
		Read:   resourceNodeV1Read,
		Update: resourceNodeV1Update,
		Delete: resourceNodeV1Delete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"boot_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ipxe",
			},
			"conductor_group": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"console_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "no-console",
			},
			"deploy_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "direct",
			},
			"driver": {
				Type:     schema.TypeString,
				Required: true,
			},
			"driver_info": {
				Type:     schema.TypeMap,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					/* FIXME: Handle other drivers than IPMI, and we also need a way to detect the password
					changed locally.  Ironic won't let us know if the password matches what we expect
					in the read call, but we should at least issue an update if the local data has changed.
					*/
					if k == "driver_info.ipmi_password" && old == "******" {
						return true
					}

					return false
				},
			},
			"properties": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"extra": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"inspect_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "inspector",
			},
			"management_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ipmitool",
			},
			"network_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "noop",
			},
			"power_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ipmitool",
			},
			"raid_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "no-raid",
			},
			"rescue_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "no-rescue",
			},
			"resource_class": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "noop",
			},
			"vendor_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ipmitool",
			},
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ports": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     resourcePortV1(),
			},
		},
	}
}

func resourceNodeV1Create(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	createOpts := nodes.CreateOpts{
		BootInterface:       d.Get("boot_interface").(string),
		ConductorGroup:      d.Get("conductor_group").(string),
		ConsoleInterface:    d.Get("console_interface").(string),
		DeployInterface:     d.Get("deploy_interface").(string),
		Driver:              d.Get("driver").(string),
		DriverInfo:          d.Get("driver_info").(map[string]interface{}),
		Extra:               d.Get("extra").(map[string]interface{}),
		InspectInterface:    d.Get("inspect_interface").(string),
		ManagementInterface: d.Get("management_interface").(string),
		Name:                d.Get("name").(string),
		NetworkInterface:    d.Get("network_interface").(string),
		Owner:               d.Get("owner").(string),
		PowerInterface:      d.Get("power_interface").(string),
		RAIDInterface:       d.Get("raid_interface").(string),
		RescueInterface:     d.Get("rescue_interface").(string),
		ResourceClass:       d.Get("resource_class").(string),
		StorageInterface:    d.Get("storage_interface").(string),
		VendorInterface:     d.Get("vendor_interface").(string),
	}

	result, err := nodes.Create(client, createOpts).Extract()

	if err != nil {
		d.SetId("")
		return err
	}

	// Setting the ID is what tells terraform we were successful
	log.Printf("[DEBUG] Node created with ID %s", d.Id())
	d.SetId(result.UUID)

	return resourceNodeV1Read(d, meta)
}

func resourceNodeV1Read(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	node, err := nodes.Get(client, d.Id()).Extract()
	if err != nil {
		d.SetId("")
		return err
	} else {
		// TODO: Ironic's Create is different than the Node object itself, GET returns things like the
		//  RaidConfig, we need to add those and handle them in CREATE?
		d.Set("boot_interface", node.BootInterface)
		d.Set("conductor_group", node.ConductorGroup)
		d.Set("console_interface", node.ConsoleInterface)
		d.Set("deploy_interface", node.DeployInterface)
		d.Set("driver", node.Driver)
		d.Set("driver_info", node.DriverInfo)
		d.Set("extra", node.Extra)
		d.Set("inspect_interface", node.InspectInterface)
		d.Set("management_interface", node.ManagementInterface)
		d.Set("name", node.Name)
		d.Set("network_interface", node.NetworkInterface)
		d.Set("owner", node.Owner)
		d.Set("power_interface", node.PowerInterface)
		d.Set("raid_interface", node.RAIDInterface)
		d.Set("rescue_interface", node.RescueInterface)
		d.Set("resource_class", node.ResourceClass)
		d.Set("storage_interface", node.StorageInterface)
		d.Set("vendor_interface", node.VendorInterface)
	}

	return nil
}

func resourceNodeV1Update(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	d.Partial(true)

	// TODO: Refactor this to be DRY and handle everything else we need to in an update
	if d.HasChange("boot_interface") {
		opts := nodes.UpdateOpts{
			nodes.UpdateOperation{
				Op:    nodes.ReplaceOp,
				Path:  "/boot_interface",
				Value: d.Get("boot_interface").(string),
			},
		}

		if _, err := nodes.Update(client, d.Id(), opts).Extract(); err != nil {
			return err
		}
	}

	if d.HasChange("conductor_group") {
		opts := nodes.UpdateOpts{
			nodes.UpdateOperation{
				Op:    nodes.ReplaceOp,
				Path:  "/boot_interface",
				Value: d.Get("boot_interface").(string),
			},
		}

		if _, err := nodes.Update(client, d.Id(), opts).Extract(); err != nil {
			return err
		}
	}

	if d.HasChange("console_interface") {
	}
	if d.HasChange("deploy_interface") {
	}
	if d.HasChange("driver") {
	}
	if d.HasChange("driver_info") {
	}
	if d.HasChange("extra") {
	}
	if d.HasChange("inspect_interface") {
	}
	if d.HasChange("management_interface") {
	}
	if d.HasChange("name") {
	}
	if d.HasChange("network_interface") {
	}
	if d.HasChange("owner") {
	}
	if d.HasChange("power_interface") {
	}
	if d.HasChange("raid_interface") {
	}
	if d.HasChange("rescue_interface") {
	}
	if d.HasChange("resource_class") {
	}
	if d.HasChange("storage_interface") {
	}
	if d.HasChange("vendor_interface") {
	}

	d.Partial(false)

	return resourceNodeV1Read(d, meta)
}

func resourceNodeV1Delete(d *schema.ResourceData, meta interface{}) error {
	return nil

}
