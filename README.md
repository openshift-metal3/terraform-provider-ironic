[![Build Status](https://travis-ci.org/metalkube/terraform-provider-ironic.svg?branch=master)](https://travis-ci.org/metalkube/terraform-provider-ironic) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Terraform provider for Ironic

This is a terraform provider that lets you provision baremetal servers managed by Ironic.

# Usage

Example:

```terraform
provider "ironic" {
  "url" = "http://localhost:6385/v1"
  "microversion" = "1.50"
}

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

# License

Apache 2.0, See LICENSE file
