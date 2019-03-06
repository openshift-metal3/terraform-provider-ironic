package ironic

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	utils "github.com/gophercloud/utils/openstack/baremetal/v1/nodes"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

// Schema resource definition for an Ironic node.
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
					/* FIXME: Password updates aren't considered. How can I know if the *local* data changed? */
					/* FIXME: Support drivers other than IPMI */
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
			"provision_state": {
				Type: schema.TypeString,
				Computed: true,
			},
			"target_provision_state": {
				Type:     schema.TypeString,
				Optional: true,

				// This did not change if the current provision state matches the target
				DiffSuppressFunc: targetStateMatchesReality,
			},
			"user_data": {
				Type: schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(_, _, _ string, _ *schema.ResourceData) bool {
					return true
				},
			},
			"network_data": {
				Type: schema.TypeMap,
				Optional: true,
				DiffSuppressFunc: func(_, _, _ string, _ *schema.ResourceData) bool {
					return true
				},
			},
			"metadata": {
				Type: schema.TypeMap,
				Optional: true,
				DiffSuppressFunc: func(_, _, _ string, _ *schema.ResourceData) bool {
					return true
				},
			},
		},
	}
}
func targetStateMatchesReality(_, old, new string, d *schema.ResourceData) bool {
	switch old {
	case "manage":
		return d.Get("provision_state").(string) == "manageable"
	case "provide":
		return d.Get("provision_state").(string) == "available"
	case "active":
		return d.Get("provision_state").(string) == "active"
	}

	return false
}

func resourceNodeV1Create(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	// Create the node object in Ironic
	createOpts := schemaToCreateOpts(d)
	result, err := nodes.Create(client, createOpts).Extract()
	if err != nil {
		d.SetId("")
		return err
	}

	// Setting the ID is what tells terraform we were successful in creating the node
	log.Printf("[DEBUG] Node created with ID %s\n", d.Id())
	d.SetId(result.UUID)

	// Some fields can only be set in an update after create
	updateOpts := postCreateUpdateOpts(d)
	if len(updateOpts) > 0 {
		log.Printf("[DEBUG] Node updates required, issuing updates: %+v\n", updateOpts)
		_, err = nodes.Update(client, d.Id(), updateOpts).Extract()
		if err != nil {
			resourceNodeV1Read(d, meta)
			return err
		}
	}

	// Create ports
	ports := d.Get("ports").(map[string]interface{})
	if ports != nil {
		// TODO
	}

	// Target provision state is special, we need to drive ironic through it's state machine
	// to reach the desired state, which could be in multiple steps.
	if target := d.Get("target_provision_state").(string); target != "" {
		opts := nodes.ProvisionStateOpts{
			Target: nodes.TargetProvisionState(target),
			//TODO: Clean Steps, Rescue Password
		}

		if target == "active" {
			configDrive, err := utils.ConfigDrive{
				UserData:    d.Get("user_data").(utils.UserDataString),
				MetaData:    d.Get("meta_data").(map[string]interface{}),
				NetworkData: d.Get("network_data").(map[string]interface{}),
			}.ToConfigDrive()
			if err != nil {
				return err
			}
			opts.ConfigDrive = configDrive
		}


		wf := workflow{
			opts: opts,
			client: client,
			uuid: d.Id(),
		}

		// Run the workflow - this could take a while
		if err := wf.run(); err != nil {
			resourceNodeV1Read(d, meta)
			return err
		}
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
	d.Set("provision_state", node.ProvisionState)
	d.Set("target_provision_state", node.TargetProvisionState)

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

type workflow struct {
	client *gophercloud.ServiceClient
	uuid   string
	opts   nodes.ProvisionStateOpts
	wait   time.Duration
	node   *nodes.Node
}

func (workflow *workflow) new() error {
	return workflow.refreshNode()
}

// Keep driving the state machine forward
func (workflow *workflow) run() (error) {
	log.Printf("[INFO] Beginning provisioning workflow, will try to change node to state '%s'", workflow.opts.Target)

	for {
		done, err := workflow.next()
		if done || err != nil {
			return err
		}

		time.Sleep(workflow.wait)
	}

	return nil
}

func (workflow *workflow) next() (done bool, err error) {
	// Refresh the node
	if err := workflow.refreshNode(); err != nil {
		return true, err
	}

	log.Printf("[DEBUG] Node current state is '%s'", workflow.node.ProvisionState)

	switch target := workflow.opts.Target; target {
	case nodes.TargetManage:
		return workflow.toManageable()
	case nodes.TargetProvide:
		return workflow.toAvailable()
	case nodes.TargetActive:
		return workflow.toActive()
	case nodes.TargetDeleted:
		return workflow.toDeleted()
	default:
		return true, fmt.Errorf("unknown target state '%s'", target)
	}
}

// Change a node to "manageable" stable
func (workflow *workflow) toManageable() (done bool, err error) {
	switch state := workflow.node.ProvisionState; state {
	case "enroll",
		"adopt failed",
		"clean failed",
		"insepct failed",
		"available":
		return workflow.changeProvisionState(nodes.TargetManage)
	case "verifying":
		workflow.wait = 15 * time.Second
		return false, nil // not done, no error - Ironic is working
	default:
		// TODO: If node is not in a position to go back to manageable, we could delete it (ForceNew) and create it again
		return true, fmt.Errorf("cannot go from state '%s' to state 'manageable'", state)
	}

	return false, nil
}

// Change a node to "available" state
func (workflow *workflow) toAvailable() (done bool, err error) {
	workflow.wait = 15 * time.Second

	switch state := workflow.node.ProvisionState; state {
	// Not done, no error - Ironic is working
	case "cleaning":
		log.Printf("[DEBUG] Node %s is not done, still cleaning.", workflow.node.UUID)
		return false, nil
	// From manageable, we can go to provide
	case "manageable":
		return workflow.changeProvisionState(nodes.TargetProvide)
	// Otherwise we have to get into manageable state first
	default:
		_, err := workflow.toManageable()
		if err != nil {
			return true, err
		}
		return false, nil
	}
}

// Change a node to "active" state
func (workflow *workflow) toActive() (bool, error) {
	workflow.wait = 30 * time.Second

	switch state := workflow.node.ProvisionState; state {
	// Not done, no error - Ironic is working
	case "deploying":
		return false, nil
	// From available, we can go to active
	case "available":
		return workflow.changeProvisionState(nodes.TargetActive)
	// Otherwise we have to get into available state first
	default:
		_, err := workflow.toAvailable()
		if err != nil {
			return true, err
		}
		return false, nil
	}
}

// Change a node to be "deleted"
func (workflow *workflow) toDeleted() (bool, error) {
	return false, nil
}

func (workflow *workflow) refreshNode() error {
	node, err := nodes.Get(workflow.client, workflow.uuid).Extract()
	if err != nil {
		return err
	}

	workflow.node = node
	return nil
}

func (workflow *workflow) changeProvisionState(target nodes.TargetProvisionState) (done bool, err error) {
	workflow.opts.Target = target
	return false, nodes.ChangeProvisionState(workflow.client, workflow.uuid, workflow.opts).ExtractErr()
}
