// ----------------------------------------------------------------------------
//
//     ***     AUTO GENERATED CODE    ***    AUTO GENERATED CODE     ***
//
// ----------------------------------------------------------------------------
//
//     This file is automatically generated by Magic Modules and manual
//     changes will be clobbered when the file is regenerated.
//
//     Please read more about how to change this file at
//     https://www.github.com/huaweicloud/magic-modules
//
// ----------------------------------------------------------------------------

package huaweicloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/huaweicloud/golangsdk"
)

func TestAccDliQueueV1_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDliQueueV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDliQueueV1_basic(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDliQueueV1Exists(),
				),
			},
		},
	})
}

func testAccDliQueueV1_basic(val string) string {
	return fmt.Sprintf(`
resource "huaweicloud_dli_queue_v1" "queue" {
  name = "terraform_dli_queue_v1_test%s"
  cu_count = 4
}
	`, val)
}

func testAccCheckDliQueueV1Destroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	client, err := config.DliV1Client(HW_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating sdk client, err=%s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "huaweicloud_dli_queue_v1" {
			continue
		}

		_, err = fetchDliQueueV1ByListOnTest(rs, client)
		if err != nil {
			if strings.Index(err.Error(), "Error finding the resource by list api") != -1 {
				return nil
			}
			return err
		}
		return fmt.Errorf("huaweicloud_dli_queue_v1 still exists")
	}

	return nil
}

func testAccCheckDliQueueV1Exists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		client, err := config.DliV1Client(HW_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating sdk client, err=%s", err)
		}

		rs, ok := s.RootModule().Resources["huaweicloud_dli_queue_v1.queue"]
		if !ok {
			return fmt.Errorf("Error checking huaweicloud_dli_queue_v1.queue exist, err=not found this resource")
		}

		_, err = fetchDliQueueV1ByListOnTest(rs, client)
		if err != nil {
			if strings.Index(err.Error(), "Error finding the resource by list api") != -1 {
				return fmt.Errorf("huaweicloud_dli_queue_v1 is not exist")
			}
			return fmt.Errorf("Error checking huaweicloud_dli_queue_v1.queue exist, err=%s", err)
		}
		return nil
	}
}

func fetchDliQueueV1ByListOnTest(rs *terraform.ResourceState,
	client *golangsdk.ServiceClient) (interface{}, error) {
	link := client.ServiceURL("queues")

	return findDliQueueV1ByList(client, link, rs.Primary.ID)
}
