package turbot

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/turbot/terraform-provider-turbot/apiClient"
	"github.com/turbot/terraform-provider-turbot/errors"
)

// test suites
func TestAccControlMute_Basic(t *testing.T) {
	resourceName := "turbot_control_mute.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckControlMuteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccControlMuteConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlMuteExists("turbot_control_mute.test"),
					resource.TestCheckResourceAttr("turbot_control_mute.test", "control_id", "330102006163524"),
					resource.TestCheckResourceAttr("turbot_control_mute.test", "note", "Muting the control for testing"),
					resource.TestCheckResourceAttr("turbot_control_mute.test", "to_timestamp", "2024-12-18T12:54:07.000Z"),
				),
			},
			{
				Config: testAccControlMuteUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlMuteExists("turbot_control_mute.test"),
					resource.TestCheckResourceAttr("turbot_control_mute.test", "control_id", "330102006163524"),
					resource.TestCheckResourceAttr("turbot_control_mute.test", "note", "Muting the control"),
					resource.TestCheckResourceAttr("turbot_control_mute.test", "to_timestamp", "2024-12-18T12:54:07.000Z"),
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

// configs
func testAccControlMuteConfig() string {
	return `
resource "turbot_control_mute" "test" {
  control_id   = "330102006163524"
  note         = "Muting the control for testing"
  to_timestamp = "2024-12-18T12:54:07.000Z"
  until_states = ["error"]
}
`
}

func testAccControlMuteUpdateConfig() string {
	return `
resource "turbot_control_mute" "test" {
  control_id   = "330102006163524"
  note         = "Muting the control"
  to_timestamp = "2024-12-18T12:54:07.000Z"
  until_states = ["alarm"]
}
`
}

// helper functions
func testAccCheckControlMuteExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}
		client := testAccProvider.Meta().(*apiClient.Client)
		_, err := client.ReadControl(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckControlMuteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*apiClient.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "turbot_control_mute" {
			_, err := client.ReadControl(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("alert still exists")
			}
			if !errors.NotFoundError(err) {
				return fmt.Errorf("expected 'not found' error, got %s", err)
			}
		}
	}

	return nil
}
