package aptible

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccResourceEndpoint_app(t *testing.T) {
	appHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointApp(appHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", appHandle),
					resource.TestCheckResourceAttr("aptible_app.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),

					resource.TestCheckResourceAttr("aptible_endpoint.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "app"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "HTTPS"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "true"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "alb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "hostname"),
				),
			},
		},
	})
}

func TestAccResourceEndpoint_db(t *testing.T) {
	dbHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointDatabase(dbHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_db.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_db.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttrSet("aptible_db.test", "db_id"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "connection_url"),

					resource.TestCheckResourceAttr("aptible_endpoint.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "database"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "TCP"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "false"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "elb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "hostname"),
				),
			},
		},
	})
}

func TestAccResourceEndpoint_updateIPWhitelist(t *testing.T) {
	appHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointApp(appHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", appHandle),
					resource.TestCheckResourceAttr("aptible_app.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),

					resource.TestCheckResourceAttr("aptible_endpoint.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "app"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "HTTPS"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "true"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "alb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "hostname"),
				),
			},
			{
				Config: testAccAptibleEndpointUpdateIPWhitelist(appHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_endpoint.test", "ip_filtering.0", "1.1.1.1/32"),
				),
			},
		},
	})
}

func testAccCheckEndpointDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	// Allow time for deprovision operation to complete.
	// TODO: Replace this by waiting on the actual operation
	time.Sleep(30 * time.Second)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_endpoint" {
			continue
		}

		res_id, err := strconv.Atoi(rs.Primary.Attributes["resource_id"])
		if err != nil {
			return err
		}

		res_typ := rs.Primary.Attributes["resource_type"]
		if err != nil {
			return err
		}

		if res_typ == "app" {
			deleted, err := client.GetApp(int64(res_id))
			log.Println("Deleted? ", deleted)
			if !deleted {
				return fmt.Errorf("App %v not removed", res_id)
			}

			if err != nil {
				return err
			}

		} else {
			_, deleted, err := client.GetDatabase(int64(res_id))
			log.Println("Deleted? ", deleted)
			if !deleted {
				return fmt.Errorf("Database %v not removed", res_id)
			}

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func testAccAptibleEndpointApp(appHandle string) string {
	output := fmt.Sprintf(`
resource "aptible_app" "test" {
	env_id = %d
	handle = "%v"
	config = {
		"APTIBLE_DOCKER_IMAGE" = "nginx"
	}
}

resource "aptible_endpoint" "test" {
	env_id = %d
	resource_id = aptible_app.test.app_id
	resource_type = "app"
	endpoint_type = "HTTPS"
	internal = true
	platform = "alb"
}`, TestEnvironmentId, appHandle, TestEnvironmentId)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointDatabase(dbHandle string) string {
	output := fmt.Sprintf(`
resource "aptible_db" "test" {
	env_id = %d
	handle = "%v"
	db_type = "postgresql"
	container_size = 1024
	disk_size = 10
}

resource "aptible_endpoint" "test" {
	env_id = %d
	resource_id = aptible_db.test.db_id
	resource_type = "database"
	endpoint_type = "TCP"
	internal = false
	platform = "elb"
}`, TestEnvironmentId, dbHandle, TestEnvironmentId)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointUpdateIPWhitelist(appHandle string) string {
	output := fmt.Sprintf(`
resource "aptible_app" "test" {
	env_id = %d
	handle = "%v"
	config = {
		"APTIBLE_DOCKER_IMAGE" = "nginx"
	}
}

resource "aptible_endpoint" "test" {
	env_id = %d
	resource_id = aptible_app.test.app_id
	resource_type = "app"
	endpoint_type = "HTTPS"
	internal = true
	platform = "alb"
	ip_filtering = [
		"1.1.1.1/32",
	]
}`, TestEnvironmentId, appHandle, TestEnvironmentId)
	log.Println("HCL generated: ", output)
	return output
}
