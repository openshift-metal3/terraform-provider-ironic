module github.com/openshift-metal3/terraform-provider-ironic

go 1.16

require (
	github.com/gophercloud/gophercloud v0.22.0
	github.com/gophercloud/utils v0.0.0-20210720165645-8a3ad2ad9e70
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk v1.17.2
	github.com/metal3-io/baremetal-operator v0.0.0-20220216092208-3612e86973f1
	github.com/metal3-io/baremetal-operator/apis v0.0.0
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils v0.0.0
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
)

replace (
	github.com/metal3-io/baremetal-operator/apis => github.com/metal3-io/baremetal-operator/apis v0.0.0-20220216092208-3612e86973f1
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils => github.com/metal3-io/baremetal-operator/pkg/hardwareutils v0.0.0-20220216092208-3612e86973f1
)
