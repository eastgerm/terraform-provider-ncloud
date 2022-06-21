package ncloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceNcloudNasVolume_classic_basic(t *testing.T) {
	testAccDataSourceNcloudNasVolumeBasic(t, false)
}

func TestAccDataSourceNcloudNasVolume_vpc_basic(t *testing.T) {
	testAccDataSourceNcloudNasVolumeBasic(t, true)
}

func testAccDataSourceNcloudNasVolumeBasic(t *testing.T, isVpc bool) {
	dataName := "data.ncloud_nas_volume.by_id"
	resourceName := "ncloud_nas_volume.test"
	postfix := getTestPrefix()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: getTestAccProviders(isVpc),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNcloudNasVolumeConfig(postfix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceID("data.ncloud_nas_volume.by_id"),
					testAccCheckDataSourceID("data.ncloud_nas_volume.by_filter"),
					resource.TestCheckResourceAttrPair(dataName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataName, "volume_size", resourceName, "volume_size"),
					resource.TestCheckResourceAttrPair(dataName, "volume_total_size", resourceName, "volume_total_size"),
					resource.TestCheckResourceAttrPair(dataName, "zone", resourceName, "zone"),
					resource.TestCheckResourceAttrPair(dataName, "snapshot_volume_size", resourceName, "snapshot_volume_size"),
					resource.TestCheckResourceAttrPair(dataName, "volume_allotment_protocol_type", resourceName, "volume_allotment_protocol_type"),
					resource.TestCheckResourceAttrPair(dataName, "is_event_configuration", resourceName, "is_event_configuration"),
					resource.TestCheckResourceAttrPair(dataName, "is_snapshot_configuration", resourceName, "is_snapshot_configuration"),
					resource.TestCheckResourceAttrPair(dataName, "mount_information", resourceName, "mount_information"),
				),
			},
		},
	})
}

func testAccDataSourceNcloudNasVolumeConfig(volumeNamePostfix string) string {
	return fmt.Sprintf(`
resource "ncloud_nas_volume" "test" {
	volume_name_postfix = "%s"
	volume_size = "500"
	volume_allotment_protocol_type = "NFS"
}

data "ncloud_nas_volume" "by_id" {
	id = ncloud_nas_volume.test.id
}

data "ncloud_nas_volume" "by_filter" {
	filter {
		name = "nas_volume_no"
		values = [ncloud_nas_volume.test.id]
	}
}`, volumeNamePostfix)
}
