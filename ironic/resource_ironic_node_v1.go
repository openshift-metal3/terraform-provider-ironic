package ironic

import (
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/ports"
	utils "github.com/gophercloud/utils/openstack/baremetal/v1/nodes"
	"github.com/hashicorp/terraform/helper/schema"
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
			"root_device": {
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
			"instance_uuid": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"provision_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_provision_state": {
				Type:     schema.TypeString,
				Optional: true,

				// This did not change if the current provision state matches the target
				DiffSuppressFunc: targetStateMatchesReality,
			},
			"power_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_power_state": {
				Type:     schema.TypeString,
				Optional: true,

				// If power_state is same as target_power_state, we have no changes to apply
				DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
					return new == d.Get("power_state").(string)
				},
			},
			"power_state_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			// FIXME: Suppress config drive on updates
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_data": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

// Match up Ironic verbs and nouns for target_provision_state
func targetStateMatchesReality(_, old, new string, d *schema.ResourceData) bool {
	log.Printf("[DEBUG] Current state is '%s', target is '%s'\n", d.Get("provision_state").(string), new)

	switch new {
	case "manage":
		return d.Get("provision_state").(string) == "manageable"
	case "provide":
		return d.Get("provision_state").(string) == "available"
	case "active":
		return d.Get("provision_state").(string) == "active"
	}

	return false
}

// Create a node, including driving Ironic's state machine
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

	// FIXME - This is ugly, but terraform doesn't handle nested resources well :-( We need the port created
	// in the middle of the process, after we create the node but before we start deploying it.  Maybe
	// there's a better solution.
	portSet := d.Get("ports").(*schema.Set)
	if portSet != nil {
		portList := portSet.List()
		for _, portInterface := range portList {
			port := portInterface.(map[string]interface{})

			// Terraform map can't handle bool... seriously.
			var pxeEnabled bool
			if port["pxe_enabled"] != nil {
				if port["pxe_enabled"] == "true" {
					pxeEnabled = true
				} else {
					pxeEnabled = false
				}

			}
			// FIXME all values other than address and pxe
			portCreateOpts := ports.CreateOpts{
				NodeUUID:   d.Id(),
				Address:    port["address"].(string),
				PXEEnabled: &pxeEnabled,
			}
			_, err := ports.Create(client, portCreateOpts).Extract()
			if err != nil {
				resourcePortV1Read(d, meta)
				return err
			}
		}
	}

	// target_provision_state is special, we need to drive ironic through it's state machine
	// to reach the desired state, which could take multiple long-running steps.
	if target := d.Get("target_provision_state").(string); target != "" {
		if err := changeProvisionStateToTarget(d, client); err != nil {
			return err
		}
	}

	// Change power state, if required
	if targetPowerState := d.Get("target_power_state").(string); targetPowerState != "" {
		err := changePowerState(client, d, nodes.TargetPowerState(targetPowerState))
		if err != nil {
			return fmt.Errorf("could not change power state: %s", err)
		}
	}

	return resourceNodeV1Read(d, meta)
}

// All the options that need to be updated on the node, post-create.  Not everything
// can be created through the POST to /v1/nodes.  TODO: The rest of the fields other
// than instance_info.
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

// Read the node's data from Ironic
func resourceNodeV1Read(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)

	node, err := nodes.Get(client, d.Id()).Extract()
	if err != nil {
		d.SetId("")
		return err
	}

	// TODO: Ironic's Create is different than the Node object itself, GET returns things like the
	//  RaidConfig, we need to add those and handle them in CREATE
	d.Set("boot_interface", node.BootInterface)
	d.Set("conductor_group", node.ConductorGroup)
	d.Set("console_interface", node.ConsoleInterface)
	d.Set("deploy_interface", node.DeployInterface)
	d.Set("driver", node.Driver)
	d.Set("driver_info", node.DriverInfo)
	d.Set("extra", node.Extra)
	d.Set("inspect_interface", node.InspectInterface)
	d.Set("instance_uuid", node.InstanceUUID)
	d.Set("management_interface", node.ManagementInterface)
	d.Set("name", node.Name)
	d.Set("network_interface", node.NetworkInterface)
	d.Set("owner", node.Owner)
	d.Set("power_interface", node.PowerInterface)
	d.Set("power_state", node.PowerState)
	d.Set("root_device", node.Properties["root_device"])
	delete(node.Properties, "root_device")
	d.Set("properties", node.Properties)
	d.Set("raid_interface", node.RAIDInterface)
	d.Set("rescue_interface", node.RescueInterface)
	d.Set("resource_class", node.ResourceClass)
	d.Set("storage_interface", node.StorageInterface)
	d.Set("vendor_interface", node.VendorInterface)
	d.Set("provision_state", node.ProvisionState)
	d.Set("target_provision_state", node.TargetProvisionState)

	return nil
}

// Update a node's state based on the terraform config - TODO: handle everything
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

	// Update power state if required
	if targetPowerState := d.Get("target_power_state").(string); d.HasChange("target_power_state") && targetPowerState != "" {
		if err := changePowerState(client, d, nodes.TargetPowerState(targetPowerState)); err != nil {
			return err
		}
	}

	// Update provision state if required - this could take a while
	if d.HasChange("target_provision_state") {
		if err := changeProvisionStateToTarget(d, client); err != nil {
			return err
		}
	}

	if d.HasChange("properties") || d.HasChange("root_device") {
		properties := propertiesMerge(d, "root_device")
		opts := nodes.UpdateOpts{
			nodes.UpdateOperation{
				Op:    nodes.AddOp,
				Path:  "/properties",
				Value: properties,
			},
		}
		if _, err := nodes.Update(client, d.Id(), opts).Extract(); err != nil {
			return err
		}
	}

	d.Partial(false)

	return resourceNodeV1Read(d, meta)
}

// Delete a node from Ironic
func resourceNodeV1Delete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gophercloud.ServiceClient)
	d.Set("target_provision_state", "deleted")
	return changeProvisionStateToTarget(d, client)
}

func propertiesMerge(d *schema.ResourceData, key string) map[string]interface{} {
	properties := d.Get("properties").(map[string]interface{})
	properties[key] = d.Get(key).(map[string]interface{})
	return properties
}

// Convert terraform schema to gophercloud CreateOpts
// TODO: Is there a better way to do this? Annotations?
func schemaToCreateOpts(d *schema.ResourceData) *nodes.CreateOpts {
	properties := propertiesMerge(d, "root_device")
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
		Properties:          properties,
		RAIDInterface:       d.Get("raid_interface").(string),
		RescueInterface:     d.Get("rescue_interface").(string),
		ResourceClass:       d.Get("resource_class").(string),
		StorageInterface:    d.Get("storage_interface").(string),
		VendorInterface:     d.Get("vendor_interface").(string),
	}
}

// provisionStateWorkflow is used to track state through the process of updating's it's provision state
type provisionStateWorkflow struct {
	client *gophercloud.ServiceClient
	d      *schema.ResourceData
	target string
	wait   time.Duration
}

// Drive Ironic's state machine through the process to reach our desired end state. This requires multiple
// possibly long-running steps.  If required, we'll build a config drive ISO for deployment.
// TODO: Handle clean steps and rescue password
func changeProvisionStateToTarget(d *schema.ResourceData, client *gophercloud.ServiceClient) error {
	defer resourceNodeV1Read(d, client) // Always refresh resource state before returning

	target := d.Get("target_provision_state").(string)

	// Run the provisionStateWorkflow - this could take a while
	wf := provisionStateWorkflow{
		target: target,
		client: client,
		wait:   5 * time.Second, // FIXME - Make configurable
		d:      d,
	}

	err := wf.run()
	return err
}

// Keep driving the state machine forward
func (workflow *provisionStateWorkflow) run() error {
	log.Printf("[INFO] Beginning provisioning workflow, will try to change node to state '%s'", workflow.target)

	for {
		done, err := workflow.next()
		if done || err != nil {
			return err
		}

		time.Sleep(workflow.wait)
	}

	return nil
}

func (workflow *provisionStateWorkflow) next() (done bool, err error) {
	// Refresh the node on each run
	if err := resourceNodeV1Read(workflow.d, workflow.client); err != nil {
		return true, err
	}

	log.Printf("[DEBUG] Node current state is '%s'", workflow.d.Get("provision_state").(string))

	switch target := nodes.TargetProvisionState(workflow.target); target {
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
func (workflow *provisionStateWorkflow) toManageable() (done bool, err error) {
	switch state := workflow.d.Get("provision_state").(string); state {
	case "manageable":
		// We're done!
		return true, err
	case "enroll",
		"adopt failed",
		"clean failed",
		"inspect failed",
		"available":
		return workflow.changeProvisionState(nodes.TargetManage)
	case "verifying":
		// Not done, no error - Ironic is working
		return false, nil

	default:
		// TODO: If node is not in a position to go back to manageable, should we delete it (ForceNew) and create it again?
		return true, fmt.Errorf("cannot go from state '%s' to state 'manageable'", state)
	}

	return false, nil
}

// Change a node to "available" state
func (workflow *provisionStateWorkflow) toAvailable() (done bool, err error) {
	switch state := workflow.d.Get("provision_state").(string); state {
	case "available":
		// We're done!
		return true, nil
	case "cleaning",
		"clean wait":
		// Not done, no error - Ironic is working
		log.Printf("[DEBUG] Node %s is '%s', waiting for Ironic to finish.", workflow.d.Id(), state)
		return false, nil
	case "manageable":
		// From manageable, we can go to provide
		log.Printf("[DEBUG] Node %s is '%s', going to change to 'available'", workflow.d.Id(), state)
		return workflow.changeProvisionState(nodes.TargetProvide)
	default:
		// Otherwise we have to get into manageable state first
		log.Printf("[DEBUG] Node %s is '%s', going to change to 'manageable'.", workflow.d.Id(), state)
		_, err := workflow.toManageable()
		if err != nil {
			return true, err
		}
		return false, nil
	}

	return false, nil
}

// Change a node to "active" state
func (workflow *provisionStateWorkflow) toActive() (bool, error) {

	switch state := workflow.d.Get("provision_state"); state {
	case "active":
		// We're done!
		log.Printf("[DEBUG] Node %s is 'active', we are done.", workflow.d.Id())
		return true, nil
	case "deploying",
		"wait call-back":
		// Not done, no error - Ironic is working
		log.Printf("[DEBUG] Node %s is '%s', waiting for Ironic to finish.", workflow.d.Id(), state)
		return false, nil
	case "available":
		// From available, we can go to active
		log.Printf("[DEBUG] Node %s is 'available', going to change to 'active'.", workflow.d.Id())
		workflow.wait = 30 * time.Second // Deployment takes a while
		return workflow.changeProvisionState(nodes.TargetActive)
	default:
		// Otherwise we have to get into available state first
		log.Printf("[DEBUG] Node %s is '%s', going to change to 'available'.", workflow.d.Id(), state)
		_, err := workflow.toAvailable()
		if err != nil {
			return true, err
		}
		return false, nil
	}
}

// Change a node to be "deleted," and remove the object from Ironic
func (workflow *provisionStateWorkflow) toDeleted() (bool, error) {
	switch state := workflow.d.Get("provision_state"); state {
	case "available",
		"enroll":
		// We're done deleting the node, we can now remove the object
		err := nodes.Delete(workflow.client, workflow.d.Id()).ExtractErr()
		return true, err
	case "cleaning",
		"deleting":
		// Not done, no error - Ironic is working
		log.Printf("[DEBUG] Node %s is '%s', waiting for Ironic to finish.", workflow.d.Id(), state)
		return false, nil
	case "active",
		"wait call-back",
		"deploy failed",
		"error":
		log.Printf("[DEBUG] Node %s is '%s', going to change to 'deleted'.", workflow.d.Id(), state)
		return workflow.changeProvisionState(nodes.TargetDeleted)
	default:
		return true, fmt.Errorf("cannot delete node in state '%s'", state)
	}

	return false, nil
}

// Builds the ProvisionStateOpts to send to Ironic -- including config drive.
func (workflow *provisionStateWorkflow) buildProvisionStateOpts(target nodes.TargetProvisionState) (*nodes.ProvisionStateOpts, error) {
	opts := nodes.ProvisionStateOpts{
		Target: target,
	}

	// If we're deploying, then build a config drive to send to Ironic
	if target == "active" {
		configDrive := utils.ConfigDrive{}

		if userData := utils.UserDataString(workflow.d.Get("user_data").(string)); userData != "" {
			configDrive.UserData = userData
		}

		if metaData := workflow.d.Get("meta_data"); metaData != nil {
			configDrive.MetaData = metaData.(map[string]interface{})
		}

		if networkData := workflow.d.Get("network_data"); networkData != nil {
			configDrive.NetworkData = networkData.(map[string]interface{})
		}

		configDriveData, err := configDrive.ToConfigDrive()
		if err != nil {
			return nil, err
		}
		opts.ConfigDrive = configDriveData
	}

	return &opts, nil
}

// Call Ironic's API and issue the change provision state request.
func (workflow *provisionStateWorkflow) changeProvisionState(target nodes.TargetProvisionState) (done bool, err error) {
	opts, err := workflow.buildProvisionStateOpts(target)
	if err != nil {
		log.Printf("[ERROR] Unable to construct provisioning state options: %s", err.Error())
		return true, err
	}

	return false, nodes.ChangeProvisionState(workflow.client, workflow.d.Id(), *opts).ExtractErr()
}

// Call Ironic's API and change the power state of the node
func changePowerState(client *gophercloud.ServiceClient, d *schema.ResourceData, target nodes.TargetPowerState) error {
	opts := nodes.PowerStateOpts{
		Target: target,
	}

	timeout := d.Get("power_state_timeout").(int)
	if timeout != 0 {
		opts.Timeout = timeout
	} else {
		timeout = 300 // used below for how long to wait for Ironic to finish
	}

	if err := nodes.ChangePowerState(client, d.Id(), opts).ExtractErr(); err != nil {
		return err
	}

	// Wait for target_power_state to be empty, i.e. Ironic thinks it's finished
	checkInterval := 5

	for {
		node, err := nodes.Get(client, d.Id()).Extract()
		if err != nil {
			return err
		}

		if node.TargetPowerState == "" {
			break
		}

		time.Sleep(time.Duration(checkInterval) * time.Second)
		timeout -= checkInterval
		if timeout <= 0 {
			return fmt.Errorf("timed out waiting for power state change")
		}
	}

	return nil
}
