package ncloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccResourceNcloudBlockStorage_basic(t *testing.T) {
	var storageInstance BlockStorage
	name := fmt.Sprintf("tf-storage-basic-%s", acctest.RandString(5))
	resourceName := "ncloud_block_storage.storage"

	testCheckResourceAttr := func() func(*terraform.State) error {
		return resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^\d+$`)),
			resource.TestCheckResourceAttr(resourceName, "name", name),
			resource.TestCheckResourceAttr(resourceName, "status", "ATTAC"),
			resource.TestCheckResourceAttr(resourceName, "size", "10"),
			resource.TestCheckResourceAttr(resourceName, "type", "SVRBS"),
			resource.TestCheckResourceAttr(resourceName, "disk_type", "NET"),
			resource.TestMatchResourceAttr(resourceName, "block_storage_no", regexp.MustCompile(`^\d+$`)),
			resource.TestMatchResourceAttr(resourceName, "server_instance_no", regexp.MustCompile(`^\d+$`)),
			resource.TestCheckResourceAttr(resourceName, "description", ""),
			resource.TestCheckResourceAttr(resourceName, "device_name", "/dev/xvdb"),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBlockStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageConfig(name),
				SkipFunc: func() (bool, error) {
					config := testAccProvider.Meta().(*ProviderConfig)
					return config.SupportVPC, nil
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageExists(resourceName, &storageInstance),
					testCheckResourceAttr(),
					resource.TestCheckResourceAttr(resourceName, "product_code", "SPBSTBSTAD000002"),
					resource.TestCheckResourceAttr(resourceName, "disk_detail_type", "HDD"),
					resource.TestMatchResourceAttr(resourceName, "instance_no", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "instance_status", "ATTAC"),
				),
			},
			{
				Config: testAccBlockStorageVpcConfig(name),
				SkipFunc: func() (bool, error) {
					config := testAccProvider.Meta().(*ProviderConfig)
					return !config.SupportVPC, nil
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageExists(resourceName, &storageInstance),
					testCheckResourceAttr(),
					resource.TestCheckResourceAttr(resourceName, "product_code", "SPBSTBSTAD000006"),
					resource.TestCheckResourceAttr(resourceName, "disk_detail_type", "SSD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBlockStorageExists(n string, i *BlockStorage) resource.TestCheckFunc {
	return testAccCheckBlockStorageExistsWithProvider(n, i, func() *schema.Provider { return testAccProvider })
}

func testAccCheckBlockStorageExistsWithProvider(n string, i *BlockStorage, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		provider := providerF()
		config := provider.Meta().(*ProviderConfig)
		storage, err := getBlockStorage(config, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if storage != nil {
			*i = *storage
			return nil
		}

		return fmt.Errorf("block storage not found")
	}
}

func testAccCheckBlockStorageDestroy(s *terraform.State) error {
	return testAccCheckBlockStorageDestroyWithProvider(s, testAccProvider)
}

func testAccCheckBlockStorageDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	config := provider.Meta().(*ProviderConfig)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ncloud_block_storage" {
			continue
		}
		blockStorage, err := getBlockStorage(config, rs.Primary.ID)

		if blockStorage == nil {
			continue
		}
		if err != nil {
			return err
		}
		if blockStorage != nil && *blockStorage.Status != "ATTAC" {
			return fmt.Errorf("found attached block storage: %s", *blockStorage.BlockStorageInstanceNo)
		}
	}

	return nil
}

func testAccBlockStorageConfig(name string) string {
	return fmt.Sprintf(`
resource "ncloud_login_key" "loginkey" {
	key_name = "%[1]s-key"
}

resource "ncloud_server" "server" {
	name = "%[1]s"
	server_image_product_code = "SPSW0LINUX000032"
	server_product_code = "SPSVRSTAND000004"
	login_key_name = "${ncloud_login_key.loginkey.key_name}"
}

resource "ncloud_block_storage" "storage" {
	server_instance_no = ncloud_server.server.id
	name = "%[1]s"
	size = "10"
}
`, name)
}

func testAccBlockStorageVpcConfig(name string) string {
	return fmt.Sprintf(`
resource "ncloud_login_key" "loginkey" {
	key_name = "%[1]s-key"
}

resource "ncloud_vpc" "test" {
	name               = "%[1]s"
	ipv4_cidr_block    = "10.5.0.0/16"
}

resource "ncloud_subnet" "test" {
	vpc_no             = ncloud_vpc.test.vpc_no
	name               = "%[1]s"
	subnet             = "10.5.0.0/24"
	zone               = "KR-2"
	network_acl_no     = ncloud_vpc.test.default_network_acl_no
	subnet_type        = "PUBLIC"
	usage_type         = "GEN"
}

resource "ncloud_server" "server" {
	subnet_no = ncloud_subnet.test.id
	name = "%[1]s"
	server_image_product_code = "SW.VSVR.OS.LNX64.CNTOS.0703.B050"
	server_product_code = "SVR.VSVR.STAND.C002.M008.NET.HDD.B050.G002"
	login_key_name = ncloud_login_key.loginkey.key_name
}

resource "ncloud_block_storage" "storage" {
	server_instance_no = ncloud_server.server.id
	name = "%[1]s"
	size = "10"
}
`, name)
}
