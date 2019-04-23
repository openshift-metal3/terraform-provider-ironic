[![Build Status](https://travis-ci.org/openshift-metal3/terraform-provider-ironic.svg?branch=master)](https://travis-ci.org/openshift-metal3/terraform-provider-ironic) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Terraform provider for Ironic

This is a terraform provider that lets you provision baremetal servers managed by Ironic.

## Provider

Currently the provider only supports standalone noauth Ironic.  At a
minimum, the Ironic endpoint URL must be specified. The user may also
optionally specify an API microversion.

```terraform
provider "ironic" {
  url = "http://localhost:6385/v1"
  microversion = "1.52"
}
```

## Resources

This provider currently implements a number of native Ironic resources,
described below.

### Nodes

A node describes a hardware resource.

```terraform
resource "ironic_node_v1" "openshift-master-0" {
  name = "openshift-master-0"
  target_provision_state = "active"
  user_data = "${file("master.ign")}"

  ports = [
    {
      "address" = "00:bb:4a:d0:5e:38"
      "pxe_enabled" = "true"
    }
  ]

  properties {
    "local_gb" = "50"
    "cpu_arch" =  "x86_64"
  }

  instance_info = {
    "image_source" = "http://172.22.0.1/images/redhat-coreos-maipo-latest.qcow2"
    "image_checksum" = "26c53f3beca4e0b02e09d335257826fd"
    "root_gb" = "25"
    "root_device" = "/dev/vda"
  }

  driver = "ipmi"
  driver_info {
			"ipmi_port"=      "6230"
			"ipmi_username"=  "admin"
			"ipmi_password"=  "password"
			"ipmi_address"=   "192.168.111.1"
			"deploy_kernel"=  "http://172.22.0.1/images/ironic-python-agent.kernel"
			"deploy_ramdisk"= "http://172.22.0.1/images/ironic-python-agent.initramfs"
  }
}
```

## Ports

Ports may be specified as part of the node resource, or as a separate `ironic_port_v1`
declaration.

```terraform
resource "ironic_port_v1" "openshift-master-0-port-0" {
  node_uuid   = "${ironic_node_v1.openshift-master-0.id}"
  pxe_enabled = true
  address     = "00:bb:4a:d0:5e:38"
}
```

## Allocation

The Allocation resource represents a request to find and allocate a Node
for deployment. The microversion must be 1.52 or later.

```terraform
resource "ironic_allocation_v1" "openshift-master-allocation" {
  name = "master-${count.index}"
  count = 3

  resource_class = "baremetal"

  candidates = [
    "${ironic_node_v1.openshift-master-0.id}",
    "${ironic_node_v1.openshift-master-1.id}",
    "${ironic_node_v1.openshift-master-2.id}",
  ]

  traits = [
    "CUSTOM_FOO",
  ]
}
```

# License

Apache 2.0, See LICENSE file
