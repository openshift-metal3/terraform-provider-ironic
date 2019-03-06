[![Build Status](https://travis-ci.org/metalkube/terraform-provider-ironic.svg?branch=master)](https://travis-ci.org/metalkube/terraform-provider-ironic) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Terraform provider for Ironic

This is a terraform provider that lets you provision baremetal servers managed by Ironic.

# Usage

Example:

```
provider "ironic" {
    "url" = "http://localhost:6385/v1"
    "microversion" = "1.50"
}

resource "ironic_node_v1" "openshift-master-0" {
    name = "openshift-master-0"
    driver = "ipmi"

    target_provision_state = "active"

    config_drive {
        user_data {}
        meta_data {}
        network_data {}
    }
         
    instance_info {
        "image_source" = "http://172.22.0.1/images/rhcos.img"
        "image_checksum" = "http://172.22.0.1/images/rhcos.img.md5sum"     
        "root_gb" = 25
    }
    
    properties {
        "name" = "/dev/vda" 
    }
  
    driver_info {
		"ipmi_username" = "admin"
		"ipmi_port" = "6230"
		"deploy_kernel" = "http://172.22.0.1/images/tinyipa-stable-rocky.vmlinuz"
		"ipmi_address" = "192.168.122.1"
		"deploy_ramdisk" = "http://172.22.0.1/images/tinyipa-stable-rocky.gz"
		"ipmi_password" =  "admin"
    }
}
```

# License

Apache 2.0, See LICENSE file
