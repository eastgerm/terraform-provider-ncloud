package ncloud

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccResourceNcloudVpc_basic(t *testing.T) {
	var vpc vpc.Vpc
	rInt := rand.Intn(16)
	cidr := fmt.Sprintf("10.%d.0.0/16", rInt)
	name := fmt.Sprintf("test-vpc-basic-%s", acctest.RandString(5))
	resourceName := "ncloud_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNcloudVpcConfig(name, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv4_cidr_block", cidr),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "status", "RUN"),
				),
			},
		},
	})
}

func TestAccResourceNcloudVpc_disappears(t *testing.T) {
	var vpc vpc.Vpc
	rInt := rand.Intn(16)
	cidr := fmt.Sprintf("10.%d.0.0/16", rInt)
	name := fmt.Sprintf("test-vpc-basic-%s", acctest.RandString(5))
	resourceName := "ncloud_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNcloudVpcConfig(name, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcDisappears(&vpc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccResourceNcloudVpc_updateName(t *testing.T) {
	var vpc vpc.Vpc
	rInt := rand.Intn(16)
	cidr := fmt.Sprintf("10.%d.0.0/16", rInt)
	name := fmt.Sprintf("test-vpc-name-%s", acctest.RandString(5))
	resourceName := "ncloud_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNcloudVpcConfig(name, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
				),
			},
			{
				Config: testAccDataSourceNcloudVpcConfig("testacc-vpc-basic-update", cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
				),
				ExpectError: regexp.MustCompile("Change 'name' is not support, Please set `name` as a old value"),
			},
		},
	})
}

func testAccResourceNcloudVpcConfig(name, cidr string) string {
	return fmt.Sprintf(`
resource "ncloud_vpc" "test" {
	name               = "%s"
  ipv4_cidr_block    = "%s"
}
`, name, cidr)
}

func testAccCheckVpcExists(n string, vpc *vpc.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		client := testAccProvider.Meta().(*NcloudAPIClient)
		vpcInstance, err := getVpcInstance(client, rs.Primary.ID)
		if err != nil {
			return err
		}

		*vpc = *vpcInstance

		return nil
	}
}

func testAccCheckVpcDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*NcloudAPIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ncloud_vpc" {
			continue
		}

		instance, err := getVpcInstance(client, rs.Primary.ID)

		if err != nil {
			return err
		}

		if instance != nil {
			return errors.New("VPC still exists")
		}
	}

	return nil
}

func testAccCheckVpcDisappears(instance *vpc.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*NcloudAPIClient)
		reqParams := &vpc.DeleteVpcRequest{
			VpcNo: instance.VpcNo,
		}

		_, err := client.vpc.V2Api.DeleteVpc(reqParams)

		waitForNcloudVpcDeletion(client, *instance.VpcNo)

		return err
	}
}