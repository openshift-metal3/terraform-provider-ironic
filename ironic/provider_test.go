// +build acceptance

package ironic

import (
	"net/http"
	"os"
	"testing"

	gth "github.com/gophercloud/gophercloud/testhelper"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	th "github.com/openshift-metal3/terraform-provider-ironic/testhelper"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)

	testAccProviders = map[string]terraform.ResourceProvider{
		"ironic": testAccProvider,
	}
}

func testAccPreCheckRequiredEnvVars(t *testing.T) {
	v := os.Getenv("IRONIC_ENDPOINT")
	if v == "" {
		t.Fatal("IRONIC_ENDPOINT must be set for acceptance tests")
	}
}

func testAccPreCheck(t *testing.T) {
	testAccPreCheckRequiredEnvVars(t)
}

func TestProvider(t *testing.T) {
	testAccPreCheck(t)

	p := Provider()
	raw := map[string]interface{}{
		"url":          "http://localhost:6385/v1",
		"microversion": "1.52",
	}
	err := p.Configure(terraform.NewResourceConfigRaw(raw))
	th.AssertNoError(t, err)
}

func TestProvider_clientTimeout(t *testing.T) {
	p := Provider()

	// Setup HTTP server listening for the API request
	gth.SetupHTTP()
	defer gth.TeardownHTTP()
	handleProviderTimeoutRequest(t)

	raw := map[string]interface{}{
		"url":     gth.Server.URL + "/",
		"timeout": 90,
	}
	err := p.Configure(terraform.NewResourceConfigRaw(raw))
	th.AssertNoError(t, err)

	client := p.(*schema.Provider).Meta().(*Clients)
	_, err = client.GetIronicClient()
	th.AssertError(t, err, "could not contact Ironic API")
}

func TestProvider_urlRequired(t *testing.T) {
	testAccPreCheck(t)

	p := Provider()
	raw := map[string]interface{}{}

	ironicEndpoint := os.Getenv("IRONIC_ENDPOINT")
	os.Unsetenv("IRONIC_ENDPOINT")

	err := p.Configure(terraform.NewResourceConfigRaw(raw))
	th.AssertError(t, err, "url is required")

	os.Setenv("IRONIC_ENDPOINT", ironicEndpoint)
}

func handleProviderTimeoutRequest(t *testing.T) {
	gth.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "This endpoint will never succeed.", http.StatusInternalServerError)
	})
}
