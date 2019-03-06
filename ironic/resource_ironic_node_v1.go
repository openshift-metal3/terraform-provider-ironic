package ironic

import (
	"fmt"
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

				// driver_info could contain passwords
				Sensitive: true,
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
			"instance_info": {
				Type:     schema.TypeMap,
				Optional: true,
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
			"target_provision_state": {
				Type:     schema.TypeMap,
				Optional: true,

				// This did not change if the current provision state matches the target
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("provision_state").(string) == old
				},
			},
		},
	}
}

func resourceNodeV1Create(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	// Create the node
	createOpts := schemaToCreateOpts(d)
	result, err := nodes.Create(client, createOpts).Extract()
	if err != nil {
		d.SetId("")
		return err
	}
	// Setting the ID is what tells terraform we were successful
	log.Printf("[DEBUG] Node created with ID %s\n", d.Id())
	d.SetId(result.UUID)

	// Some fields can only be set in an update after create
	updateOpts := postCreateUpdateOpts(d)
	if len(updateOpts) > 0 {
		log.Printf("[DEBUG] Node updates required, issuing updates: %+v\n", updateOpts)
		_, err = nodes.Update(client, d.Id(), updateOpts).Extract()
		if err != nil {
			d.SetId("") // TODO: Should I do this if create succeeds, update fails?
			return err
		}
	}

	// Create ports
	ports := d.Get("ports").(map[string]interface{})
	if ports != nil {
		// TODO the needful
	}


	// Target provision state is special, we need to drive ironic through it's state machine
	// so, how we proceed is dependent on the current state of the node. We can rely on gophercloud
	// utils to do this (TODO)
	if d.Get("target_provision_State").(string) == "" {
		// TODO the needful
	}

	return resourceNodeV1Read(d, meta)
}

func postCreateUpdateOpts(d *schema.ResourceData) nodes.UpdateOpts {
	opts := nodes.UpdateOpts{}

	instanceInfo := d.Get("instance_info").(map[string]interface{})
	if instanceInfo != nil {
		opts = append(opts, nodes.UpdateOperation{
			Op:    nodes.AddOp,
			Path:  "/instance_info",
			Value: instanceInfo,
		},
		)
	}

	return opts
}

func resourceNodeV1Read(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	node, err := nodes.Get(client, d.Id()).Extract()
	if err != nil {
		d.SetId("")
		return err
	}

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

	return nil
}

func resourceNodeV1Update(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	d.Partial(true)

	stringFields := []string{
		"boot_interface",
		"conductor_group",
		"console_interface",
		"deploy_interface",
		"driver",
		"inspect_interface",
		"management_interface",
		"name",
		"network_interface",
		"owner",
		"power_interface",
		"raid_interface",
		"rescue_interface",
		"resource_class",
		"storage_interface",
		"vendor_interface",
	}

	for _, field := range stringFields {
		if d.HasChange(field) {
			opts := nodes.UpdateOpts{
				nodes.UpdateOperation{
					Op:    nodes.ReplaceOp,
					Path:  fmt.Sprintf("/%s", field),
					Value: d.Get(field).(string),
				},
			}

			if _, err := nodes.Update(client, d.Id(), opts).Extract(); err != nil {
				return err
			}
		}
	}

	d.Partial(false)

	return resourceNodeV1Read(d, meta)
}

func resourceNodeV1Delete(d *schema.ResourceData, meta interface{}) error {
	return nil

}

func schemaToCreateOpts(d *schema.ResourceData) *nodes.CreateOpts {
	return &nodes.CreateOpts{
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
}
