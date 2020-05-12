package aptible

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/aptible/go-deploy/aptible"
)

func TestAccResourceApp_basic(t *testing.T) {
	rHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleAppBasic(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_app.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
				),
			},
		},
	})
}

func TestAccResourceApp_deploy(t *testing.T) {
	rHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleAppDeploy(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_app.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "nginx"),
					resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
				),
			},
		},
	})
}

func TestAccResourceApp_updateConfig(t *testing.T) {
	rHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleAppDeploy(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_app.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "nginx"),
					resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
				),
			},
			{
				Config: testAccAptibleAppUpdateConfig(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "httpd:alpine"),
					resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "nothing"),
				),
			},
		},
	})
}

func testAccCheckAppDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_app" {
			continue
		}

		app_id, err := strconv.Atoi(rs.Primary.Attributes["app_id"])
		if err != nil {
			return err
		}

		deleted, err := client.GetApp(int64(app_id))
		log.Println("Deleted? ", deleted)
		if !deleted {
			return fmt.Errorf("App %v not removed", app_id)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccAptibleAppBasic(handle string) string {
	return fmt.Sprintf(`
resource "aptible_app" "test" {
    env_id = %d
    handle = "%v"
}
`, TestEnvironmentId, handle)
}

func testAccAptibleAppDeploy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_app" "test" {
		env_id = %d
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
		}
	}
	`, TestEnvironmentId, handle)
}

func testAccAptibleAppUpdateConfig(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_app" "test" {
		env_id = %d
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "httpd:alpine"
			"WHATEVER" = "nothing"
		}
	}
	`, TestEnvironmentId, handle)
}
