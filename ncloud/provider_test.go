package ncloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"strings"
	"testing"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

var credsEnvVars = []string{
	"NCLOUD_ACCESS_KEY",
	"NCLOUD_SECRET_KEY",
}

var regionEnvVar = "NCLOUD_REGION"

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ncloud": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := multiEnvSearch(credsEnvVars); v == "" {
		t.Fatalf("One of %s must be set for acceptance tests", strings.Join(credsEnvVars, ", "))
	}

	region := testAccGetRegion()
	log.Printf("[INFO] Test: Using %s as test region", region)
	os.Setenv(regionEnvVar, region)

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func testAccGetRegion() string {
	v := os.Getenv(regionEnvVar)
	if v == "" {
		return "KR"
	}
	return v
}

func multiEnvSearch(ks []string) string {
	for _, k := range ks {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func getTestPrefix() string {
	rInt := acctest.RandIntRange(1, 9999)
	return fmt.Sprintf("tf%d", rInt)
}

func testApiClient(t *testing.T) (*NcloudAPIClient, error) {
	s := map[string]*schema.Schema{
		"access_key": {
			Type: schema.TypeString,
		},
		"secret_key": {
			Type: schema.TypeString,
		},
	}
	d := schema.TestResourceDataRaw(t, s, map[string]interface{}{
		"access_key": os.Getenv("NCLOUD_ACCESS_KEY"),
		"secret_key": os.Getenv("NCLOUD_SECRET_KEY"),
	})
	client, err := providerConfigure(d)
	return client.(*NcloudAPIClient), err
}
