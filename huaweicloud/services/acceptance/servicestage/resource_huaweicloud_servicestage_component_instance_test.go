package servicestage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk/openstack/servicestage/v2/instances"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
)

func getComponentInstanceFunc(conf *config.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.ServiceStageV2Client(acceptance.HW_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating ServiceStage v2 client: %s", err)
	}
	return instances.Get(c, state.Primary.Attributes["application_id"], state.Primary.Attributes["component_id"],
		state.Primary.ID)
}

func TestAccComponentInstance_basic(t *testing.T) {
	var (
		instance     instances.Instance
		randName     = acceptance.RandomAccResourceNameWithDash()
		resourceName = "huaweicloud_servicestage_component_instance.test"
	)

	rc := acceptance.InitResourceCheck(
		resourceName,
		&instance,
		getComponentInstanceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
			acceptance.TestAccPreCheckRepoTokenAuth(t)
			acceptance.TestAccPreCheckComponent(t)
			acceptance.TestAccPreCheckComponentDeployment(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccComponentInstance_basic(randName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "huaweicloud_servicestage_application.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "component_id", "huaweicloud_servicestage_component.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", "huaweicloud_servicestage_environment.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", randName),
					resource.TestCheckResourceAttr(resourceName, "version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "replica", "1"),
					resource.TestCheckResourceAttr(resourceName, "flavor_id", "CUSTOM-10G:250m-250m:0.5Gi-0.5Gi"),
					resource.TestCheckResourceAttr(resourceName, "description", "Created by terraform test"),
					resource.TestCheckResourceAttr(resourceName, "artifact.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "artifact.0.name", "huaweicloud_servicestage_component.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "artifact.0.type", "image"),
					resource.TestCheckResourceAttr(resourceName, "artifact.0.storage", "swr"),
					resource.TestCheckResourceAttr(resourceName, "artifact.0.url", acceptance.HW_BUILD_IMAGE_URL),
					resource.TestCheckResourceAttr(resourceName, "artifact.0.auth_type", "iam"),
					resource.TestCheckResourceAttr(resourceName, "refer_resource.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.env_variable.0.name", "TZ"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.env_variable.0.value", "Asia/Shanghai"),
					resource.TestCheckResourceAttr(resourceName, "status", "RUNNING"),
				),
			},
			{
				Config: testAccComponentInstance_update(randName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "version", "1.0.2"),
					resource.TestCheckResourceAttr(resourceName, "flavor_id", "CUSTOM-15G:500m-500m:1Gi-1Gi"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "RUNNING"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccInstanceImportStateIdFunc(),
			},
		},
	})
}

func testAccInstanceImportStateIdFunc() resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		var appId, componentId, instance_id string
		for _, rs := range s.RootModule().Resources {
			if rs.Type == "huaweicloud_servicestage_component_instance" {
				appId = rs.Primary.Attributes["application_id"]
				componentId = rs.Primary.Attributes["component_id"]
				instance_id = rs.Primary.ID
			}
		}
		if appId == "" || componentId == "" || instance_id == "" {
			return "", fmt.Errorf("resource not found: %s/%s/%s", appId, componentId, instance_id)
		}
		return fmt.Sprintf("%s/%s/%s", appId, componentId, instance_id), nil
	}
}

func testAccComponentInstance_base(rName string) string {
	return fmt.Sprintf(`
data "huaweicloud_availability_zones" "test" {}

data "huaweicloud_compute_flavors" "test" {
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  performance_type  = "normal"
  cpu_core_count    = 8
  memory_size       = 16
}

resource "huaweicloud_kps_keypair" "test" {
  name = "%[1]s"
}

resource "huaweicloud_vpc" "test" {
  name = "%[1]s"
  cidr = "192.168.0.0/16"
}

resource "huaweicloud_vpc_subnet" "test" {
  name        = "%[1]s"
  cidr        = "192.168.0.0/24"
  gateway_ip  = "192.168.0.1"
  vpc_id      = huaweicloud_vpc.test.id
  ipv6_enable = true
}

resource "huaweicloud_vpc_eip" "test" {
  publicip {
    type = "5_bgp"
  }

  bandwidth {
    share_type  = "PER"
    size        = 5
    name        = "%[1]s"
    charge_mode = "traffic"
  }
}
  
resource "huaweicloud_cce_cluster" "test" {
  name                   = "%[1]s"
  vpc_id                 = huaweicloud_vpc.test.id
  subnet_id              = huaweicloud_vpc_subnet.test.id
  flavor_id              = "cce.s2.medium"
  cluster_version        = "v1.19"
  cluster_type           = "VirtualMachine"
  container_network_type = "vpc-router"
  kube_proxy_mode        = "iptables"

  dynamic "masters" {
    for_each = slice(data.huaweicloud_availability_zones.test.names, 0, 3)

    content {
      availability_zone = masters.value
    }
  }
}

resource "huaweicloud_cce_node" "test" {
  cluster_id        = huaweicloud_cce_cluster.test.id
  name              = "%[1]s"
  flavor_id         = data.huaweicloud_compute_flavors.test.ids[0]
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  key_pair          = huaweicloud_kps_keypair.test.name
  eip_id            = huaweicloud_vpc_eip.test.id

  root_volume {
    volumetype = "SSD"
    size       = 100
  }

  data_volumes {
    volumetype = "SSD"
    size       = 100
  }
}

resource "huaweicloud_servicestage_environment" "test" {
  name   = "%[1]s"
  vpc_id = huaweicloud_vpc.test.id

  basic_resources {
    type = "cce"
    id   = huaweicloud_cce_cluster.test.id
  }

  optional_resources {
    type = "cse"
    id   = "default"
  }
}

resource "huaweicloud_servicestage_application" "test" {
  name = "%[1]s"
}

resource "huaweicloud_servicestage_repo_token_authorization" "test" {
  type  = "github"
  name  = "%[1]s"
  host  = "%[2]s"
  token = "%[3]s"
}

resource "huaweicloud_servicestage_component" "test" {
  application_id = huaweicloud_servicestage_application.test.id

  name      = "%[1]s"
  type      = "MicroService"
  runtime   = "Docker"
  framework = "Java Classis"
}
`, rName, acceptance.HW_GITHUB_REPO_HOST, acceptance.HW_GITHUB_PERSONAL_TOKEN, acceptance.HW_GITHUB_REPO_URL,
		acceptance.HW_DOMAIN_NAME)
}

func testAccComponentInstance_basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "huaweicloud_servicestage_component_instance" "test" {
  application_id = huaweicloud_servicestage_application.test.id
  component_id   = huaweicloud_servicestage_component.test.id
  environment_id = huaweicloud_servicestage_environment.test.id

  name        = "%[2]s"
  version     = "1.0.0"
  replica     = 1
  flavor_id   = "CUSTOM-10G:250m-250m:0.5Gi-0.5Gi"
  description = "Created by terraform test"

  artifact {
    name      = huaweicloud_servicestage_component.test.name
    type      = "image"
    storage   = "swr"
    url       = "%[3]s"
    auth_type = "iam"
  }

  refer_resource {
    type = "cce"
    id   = huaweicloud_cce_cluster.test.id

    parameters = {
      type      = "VirtualMachine"
      namespace = "default"
    }
  }

  refer_resource {
    type = "cse"
    id   = "default"
  }

  configuration {
    env_variable {
      name  = "TZ"
      value = "Asia/Shanghai"
    }
  }

  lifecycle {
    ignore_changes = [
      configuration[0].env_variable,
    ]
  }
}
`, testAccComponentInstance_base(rName), rName, acceptance.HW_BUILD_IMAGE_URL)
}

func testAccComponentInstance_update(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "huaweicloud_servicestage_component_instance" "test" {
  application_id = huaweicloud_servicestage_application.test.id
  component_id   = huaweicloud_servicestage_component.test.id
  environment_id = huaweicloud_servicestage_environment.test.id

  name        = "%[2]s"
  version     = "1.0.2"
  replica     = 1
  flavor_id   = "CUSTOM-15G:500m-500m:1Gi-1Gi"

  artifact {
    name      = huaweicloud_servicestage_component.test.name
    type      = "image"
    storage   = "swr"
    url       = "%[3]s"
    auth_type = "iam"
  }

  refer_resource {
    type = "cce"
    id   = huaweicloud_cce_cluster.test.id

    parameters = {
      type      = "VirtualMachine"
      namespace = "default"
    }
  }

  refer_resource {
    type = "cse"
    id   = "default"
  }

  configuration {
    env_variable {
      name  = "TZ"
      value = "Asia/Shanghai"
    }
  }

  lifecycle {
    ignore_changes = [
      configuration[0].env_variable,
    ]
  }
}
`, testAccComponentInstance_base(rName), rName, acceptance.HW_BUILD_IMAGE_URL)
}