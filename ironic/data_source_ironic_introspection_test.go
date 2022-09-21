// +build acceptance

package ironic

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	gth "github.com/gophercloud/gophercloud/testhelper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	th "github.com/openshift-metal3/terraform-provider-ironic/testhelper"
)

// TestAccIntrospection creates a node resource that is inspected, and verifies the introspection data source returns
// the information we expect.  The introspection API is mocked, as no data is available when using the 'fake' interface
// in tests.
func TestAccIntrospection(t *testing.T) {
	nodeName := th.RandomString("TerraformACC-Node-", 8)

	gth.SetupHTTP()
	defer gth.TeardownHTTP()
	handleIntrospectionRequest(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIntrospectionResource(nodeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ironic_introspection.test-data", "finished", "true"),
					resource.TestCheckResourceAttr("data.ironic_introspection.test-data", "interfaces.0.ip", "192.168.111.20"),
					resource.TestCheckResourceAttr("data.ironic_introspection.test-data", "cpu_arch", "x86_64"),
					resource.TestCheckResourceAttr("data.ironic_introspection.test-data", "cpu_count", "4"),
					resource.TestCheckResourceAttr("data.ironic_introspection.test-data", "memory_mb", "16384"),
				),
			},
		},
	})
}

// Returns a resource declaration for a particular node name, and it's related introspection data source.
func testAccIntrospectionResource(node string) string {
	return fmt.Sprintf(`
		resource "ironic_node_v1" "%s" {
			name = "%s"
			driver = "fake-hardware"

			manage = true
			inspect = true

			inspect_interface = "fake"
			boot_interface = "fake"
			deploy_interface = "fake"
			management_interface = "fake"
			power_interface = "fake"
			vendor_interface = "no-vendor"
		}

		data "ironic_introspection" "test-data" {
			uuid = ironic_node_v1.%s.id
		}
`, node, node, node)
}

const introspectionStatus = `
{
    "error": null,
    "finished": true,
    "finished_at": "2019-06-12T13:08:54.245519",
    "links": [
        {
            "href": "http://localhost:5050/v1/introspection/77f69c3c-5ab9-48f1-b044-89f5410188c1",
            "rel": "self"
        }
    ],
    "started_at": "2019-06-12T13:06:52.280981",
    "state": "finished",
    "uuid": "77f69c3c-5ab9-48f1-b044-89f5410188c1"
}
`

const introspectionData = `
{
    "cpu_arch": "x86_64",
    "macs": [
        "00:61:f4:72:48:d7"
    ],
    "root_disk": {
        "rotational": true,
        "vendor": "QEMU",
        "name": "/dev/sda",
        "wwn_vendor_extension": null,
        "hctl": "2:0:0:0",
        "wwn_with_extension": null,
        "by_path": "/dev/disk/by-path/pci-0000:00:05.0-scsi-0:0:0:0",
        "model": "QEMU HARDDISK",
        "wwn": null,
        "serial": "drive-scsi0-0-0-0",
        "size": 53687091200
    },
    "extra": {
        "network": {
            "eth1": {
                "tx-udp_tnl-csum-segmentation": "off [fixed]",
                "vlan-challenged": "off [fixed]",
                "rx-vlan-offload": "off [fixed]",
                "ipv4-network": "192.168.111.0",
                "rx-vlan-stag-filter": "off [fixed]",
                "highdma": "on [fixed]",
                "tx-nocache-copy": "off",
                "tx-gso-robust": "off [fixed]",
                "tcp-segmentation-offload/tx-tcp-mangleid-segmentation": "off",
                "rx-udp_tunnel-port-offload": "off [fixed]",
                "rx-gro-hw": "off [fixed]",
                "netns-local": "off [fixed]",
                "tx-vlan-stag-hw-insert": "off [fixed]",
                "serial": "00:61:f4:72:48:d9",
                "l2-fwd-offload": "off [fixed]",
                "large-receive-offload": "off [fixed]",
                "tx-checksumming/tx-checksum-ipv6": "off [fixed]",
                "tx-checksumming/tx-checksum-ipv4": "off [fixed]",
                "ipv4-netmask": "255.255.255.0",
                "tcp-segmentation-offload/tx-tcp-segmentation": "on",
                "tcp-segmentation-offload": "on",
                "tx-udp_tnl-segmentation": "off [fixed]",
                "tx-gre-segmentation": "off [fixed]",
                "fcoe-mtu": "off [fixed]",
                "tcp-segmentation-offload/tx-tcp-ecn-segmentation": "on",
                "tx-sctp-segmentation": "off [fixed]",
                "tx-checksumming/tx-checksum-fcoe-crc": "off [fixed]",
                "ipv4": "192.168.111.20",
                "businfo": "virtio@1",
                "rx-vlan-stag-hw-parse": "off [fixed]",
                "tx-vlan-offload": "off [fixed]",
                "tx-checksumming": "on",
                "udp-fragmentation-offload": "on",
                "tx-checksumming/tx-checksum-sctp": "off [fixed]",
                "driver": "virtio_net",
                "tx-sit-segmentation": "off [fixed]",
                "busy-poll": "off [fixed]",
                "scatter-gather/tx-scatter-gather-fraglist": "off [fixed]",
                "tx-checksumming/tx-checksum-ip-generic": "on",
                "link": "yes",
                "rx-all": "off [fixed]",
                "tx-ipip-segmentation": "off [fixed]",
                "tcp-segmentation-offload/tx-tcp6-segmentation": "on",
                "rx-checksumming": "on [fixed]",
                "loopback": "off [fixed]",
                "generic-segmentation-offload": "on",
                "tx-fcoe-segmentation": "off [fixed]",
                "tx-lockless": "off [fixed]",
                "ipv4-cidr": 24,
                "ntuple-filters": "off [fixed]",
                "rx-vlan-filter": "on [fixed]",
                "tx-gre-csum-segmentation": "off [fixed]",
                "tx-gso-partial": "off [fixed]",
                "receive-hashing": "off [fixed]",
                "scatter-gather/tx-scatter-gather": "on",
                "generic-receive-offload": "on",
                "rx-fcs": "off [fixed]",
                "scatter-gather": "on",
                "hw-tc-offload": "off [fixed]"
            },
            "eth0": {
                "tx-udp_tnl-csum-segmentation": "off [fixed]",
                "vlan-challenged": "off [fixed]",
                "rx-vlan-offload": "off [fixed]",
                "ipv4-network": "172.22.0.0",
                "rx-vlan-stag-filter": "off [fixed]",
                "highdma": "on [fixed]",
                "tx-nocache-copy": "off",
                "tx-gso-robust": "off [fixed]",
                "tcp-segmentation-offload/tx-tcp-mangleid-segmentation": "off",
                "rx-udp_tunnel-port-offload": "off [fixed]",
                "rx-gro-hw": "off [fixed]",
                "netns-local": "off [fixed]",
                "tx-vlan-stag-hw-insert": "off [fixed]",
                "serial": "00:61:f4:72:48:d7",
                "l2-fwd-offload": "off [fixed]",
                "large-receive-offload": "off [fixed]",
                "tx-checksumming/tx-checksum-ipv6": "off [fixed]",
                "tx-checksumming/tx-checksum-ipv4": "off [fixed]",
                "ipv4-netmask": "255.255.255.0",
                "tcp-segmentation-offload/tx-tcp-segmentation": "on",
                "tcp-segmentation-offload": "on",
                "tx-udp_tnl-segmentation": "off [fixed]",
                "tx-gre-segmentation": "off [fixed]",
                "fcoe-mtu": "off [fixed]",
                "tcp-segmentation-offload/tx-tcp-ecn-segmentation": "on",
                "tx-sctp-segmentation": "off [fixed]",
                "tx-checksumming/tx-checksum-fcoe-crc": "off [fixed]",
                "ipv4": "172.22.0.72",
                "businfo": "virtio@0",
                "rx-vlan-stag-hw-parse": "off [fixed]",
                "tx-vlan-offload": "off [fixed]",
                "tx-checksumming": "on",
                "udp-fragmentation-offload": "on",
                "tx-checksumming/tx-checksum-sctp": "off [fixed]",
                "driver": "virtio_net",
                "tx-sit-segmentation": "off [fixed]",
                "busy-poll": "off [fixed]",
                "scatter-gather/tx-scatter-gather-fraglist": "off [fixed]",
                "tx-checksumming/tx-checksum-ip-generic": "on",
                "link": "yes",
                "rx-all": "off [fixed]",
                "tx-ipip-segmentation": "off [fixed]",
                "tcp-segmentation-offload/tx-tcp6-segmentation": "on",
                "rx-checksumming": "on [fixed]",
                "loopback": "off [fixed]",
                "generic-segmentation-offload": "on",
                "tx-fcoe-segmentation": "off [fixed]",
                "tx-lockless": "off [fixed]",
                "ipv4-cidr": 24,
                "ntuple-filters": "off [fixed]",
                "rx-vlan-filter": "on [fixed]",
                "tx-gre-csum-segmentation": "off [fixed]",
                "tx-gso-partial": "off [fixed]",
                "receive-hashing": "off [fixed]",
                "scatter-gather/tx-scatter-gather": "on",
                "generic-receive-offload": "on",
                "rx-fcs": "off [fixed]",
                "scatter-gather": "on",
                "hw-tc-offload": "off [fixed]"
            }
        },
        "firmware": {
            "bios": {
                "date": "04/01/2014",
                "version": "1.11.0-2.el7",
                "vendor": "SeaBIOS"
            }
        },
        "system": {
            "kernel": {
                "cmdline": "ipa-inspection-callback-url=mdns ipa-inspection-collectors=default,extra-hardware,logs systemd.journald.forward_to_console=yes BOOTIF=00:61:f4:72:48:d7 ipa-debug=1 ipa-inspection-dhcp-all-interfaces=1 ipa-collect-lldp=1 initrd=ironic-python-agent.initramfs",
                "version": "3.10.0-957.12.2.el7.x86_64",
                "arch": "x86_64"
            },
            "product": {
                "version": "RHEL 7.6.0 PC (i440FX + PIIX, 1996)",
                "vendor": "Red Hat",
                "name": "KVM",
                "uuid": "9408bbc8-0510-4451-acd3-05e2fd7b4181"
            },
            "rtc": {
                "utc": "unknown"
            }
        },
        "numa": {
            "nodes": {
                "count": 1
            },
            "node_0": {
                "cpu_mask": "0xf",
                "cpu_count": 4
            }
        },
        "memory": {
            "total": {
                "size": 17179869184
            }
        },
        "disk": {
            "sda": {
                "Write Cache Enable": 1,
                "rotational": 1,
                "vendor": "QEMU",
                "rev": "2.5+",
                "scsi-id": "scsi-0QEMU_QEMU_HARDDISK_drive-scsi0-0-0-0",
                "optimal_io_size": 0,
                "SMART/vendor": "QEMU",
                "physical_block_size": 512,
                "scheduler": "deadline",
                "Read Cache Disable": 0,
                "model": "QEMU HARDDISK",
                "nr_requests": 128,
                "SMART/product": "QEMU HARDDISK",
                "size": 53
            },
            "logical": {
                "count": 1
            }
        },
        "cpu": {
            "logical": {
                "number": 4
            },
            "physical_0": {
                "product": "Intel(R) Xeon(R) CPU E5-2670 v3 @ 2.30GHz",
                "l1d cache": "32K",
                "family": 6,
                "l3 cache": "16384K",
                "current_Mhz": 2299,
                "flags": "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology eagerfpu pni pclmulqdq vmx ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm ssbd ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid xsaveopt arat spec_ctrl intel_stibp",
                "l1i cache": "32K",
                "stepping": 2,
                "cores": 1,
                "vendor": "GenuineIntel",
                "threads": 1,
                "model": 63,
                "l2 cache": "4096K"
            },
            "physical_1": {
                "product": "Intel(R) Xeon(R) CPU E5-2670 v3 @ 2.30GHz",
                "l1d cache": "32K",
                "family": 6,
                "l3 cache": "16384K",
                "current_Mhz": 2299,
                "flags": "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology eagerfpu pni pclmulqdq vmx ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm ssbd ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid xsaveopt arat spec_ctrl intel_stibp",
                "l1i cache": "32K",
                "stepping": 2,
                "cores": 1,
                "vendor": "GenuineIntel",
                "threads": 1,
                "model": 63,
                "l2 cache": "4096K"
            },
            "physical_2": {
                "product": "Intel(R) Xeon(R) CPU E5-2670 v3 @ 2.30GHz",
                "l1d cache": "32K",
                "family": 6,
                "l3 cache": "16384K",
                "current_Mhz": 2299,
                "flags": "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology eagerfpu pni pclmulqdq vmx ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm ssbd ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid xsaveopt arat spec_ctrl intel_stibp",
                "l1i cache": "32K",
                "stepping": 2,
                "cores": 1,
                "vendor": "GenuineIntel",
                "threads": 1,
                "model": 63,
                "l2 cache": "4096K"
            },
            "physical_3": {
                "product": "Intel(R) Xeon(R) CPU E5-2670 v3 @ 2.30GHz",
                "l1d cache": "32K",
                "family": 6,
                "l3 cache": "16384K",
                "current_Mhz": 2299,
                "flags": "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology eagerfpu pni pclmulqdq vmx ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm ssbd ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid xsaveopt arat spec_ctrl intel_stibp",
                "l1i cache": "32K",
                "stepping": 2,
                "cores": 1,
                "vendor": "GenuineIntel",
                "threads": 1,
                "model": 63,
                "l2 cache": "4096K"
            },
            "physical": {
                "number": 4
            }
        }
    },
    "all_interfaces": {
        "eth1": {
            "ip": "192.168.111.20",
            "mac": "00:61:f4:72:48:d9",
            "client_id": null,
            "pxe": false
        }
    },
    "cpus": 4,
    "boot_interface": "00:61:f4:72:48:d7",
    "memory_mb": 16384,
    "ipmi_address": null,
    "inventory": {
        "bmc_address": "0.0.0.0",
        "interfaces": [
            {
                "lldp": [

                ],
                "ipv6_address": "fe80::261:f4ff:fe72:48d7%%eth0",
                "vendor": "0x1af4",
                "name": "eth0",
                "has_carrier": true,
                "product": "0x0001",
                "ipv4_address": "172.22.0.72",
                "biosdevname": null,
                "client_id": null,
                "mac_address": "00:61:f4:72:48:d7"
            },
            {
                "lldp": [

                ],
                "ipv6_address": "fe80::261:f4ff:fe72:48d9%%eth1",
                "vendor": "0x1af4",
                "name": "eth1",
                "has_carrier": true,
                "product": "0x0001",
                "ipv4_address": "192.168.111.20",
                "biosdevname": null,
                "client_id": null,
                "mac_address": "00:61:f4:72:48:d9"
            }
        ],
        "disks": [
            {
                "rotational": true,
                "vendor": "QEMU",
                "name": "/dev/sda",
                "wwn_vendor_extension": null,
                "hctl": "2:0:0:0",
                "wwn_with_extension": null,
                "by_path": "/dev/disk/by-path/pci-0000:00:05.0-scsi-0:0:0:0",
                "model": "QEMU HARDDISK",
                "wwn": null,
                "serial": "drive-scsi0-0-0-0",
                "size": 53687091200
            }
        ],
        "boot": {
            "current_boot_mode": "bios",
            "pxe_interface": "00:61:f4:72:48:d7"
        },
        "system_vendor": {
            "serial_number": "",
            "product_name": "KVM",
            "manufacturer": "Red Hat"
        },
        "bmc_v6address": "::/0",
        "memory": {
            "physical_mb": 16384,
            "total": 16825597952
        },
        "cpu": {
            "count": 4,
            "frequency": "2299.998",
            "flags": [
                "fpu",
                "vme",
                "de",
                "pse",
                "tsc",
                "msr",
                "pae",
                "mce",
                "cx8",
                "apic",
                "sep",
                "mtrr",
                "pge",
                "mca",
                "cmov",
                "pat",
                "pse36",
                "clflush",
                "mmx",
                "fxsr",
                "sse",
                "sse2",
                "ss",
                "syscall",
                "nx",
                "pdpe1gb",
                "rdtscp",
                "lm",
                "constant_tsc",
                "arch_perfmon",
                "rep_good",
                "nopl",
                "xtopology",
                "eagerfpu",
                "pni",
                "pclmulqdq",
                "vmx",
                "ssse3",
                "fma",
                "cx16",
                "pcid",
                "sse4_1",
                "sse4_2",
                "x2apic",
                "movbe",
                "popcnt",
                "tsc_deadline_timer",
                "aes",
                "xsave",
                "avx",
                "f16c",
                "rdrand",
                "hypervisor",
                "lahf_lm",
                "abm",
                "ssbd",
                "ibrs",
                "ibpb",
                "stibp",
                "tpr_shadow",
                "vnmi",
                "flexpriority",
                "ept",
                "vpid",
                "fsgsbase",
                "tsc_adjust",
                "bmi1",
                "avx2",
                "smep",
                "bmi2",
                "erms",
                "invpcid",
                "xsaveopt",
                "arat",
                "spec_ctrl",
                "intel_stibp"
            ],
            "model_name": "Intel(R) Xeon(R) CPU E5-2670 v3 @ 2.30GHz",
            "architecture": "x86_64"
        }
    },
    "error": null,
    "local_gb": 49,
    "interfaces": {
        "eth0": {
            "ip": "172.22.0.72",
            "mac": "00:61:f4:72:48:d7",
            "client_id": null,
            "pxe": true
        }
    }
}
`

// When using the fake inspect interface, inspector isn't actually used, so we need
// to mock the inspector's API responses for returning data to us.
func handleIntrospectionRequest(t *testing.T) {
	gth.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		log.Printf("[DEBUG] URL is %s", r.URL.Path)

		if strings.HasSuffix(r.URL.Path, "data") {
			fmt.Fprintf(w, introspectionData)
		} else {
			fmt.Fprintf(w, introspectionStatus)
		}
	})

	os.Setenv("IRONIC_INSPECTOR_ENDPOINT", gth.Server.URL+"/v1")
}
