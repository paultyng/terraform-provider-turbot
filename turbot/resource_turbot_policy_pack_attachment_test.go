package turbot

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/turbot/terraform-provider-turbot/apiClient"
	"github.com/turbot/terraform-provider-turbot/errors"
	"testing"
)

// test suites
func TestAccPolicyPackAttachment_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyPackAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyPackAttachmentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyPackAttachmentExists("turbot_policy_pack_attachment.test"),
				),
			},
		},
	})
}

// configs
func testAccPolicyPackAttachmentConfig() string {
	return `
resource "turbot_folder" "test" {
  parent = "tmod:@turbot/turbot#/"
  title = "provider_test"
  description = "test folder"
}

resource "turbot_policy_pack" "test" {
  filter      = "resourceType:181381985925765 $.turbot.tags.a:b"
  description = "Policy Pack Testing"
  title       = "policy_pack"
}

resource "turbot_policy_pack_attachment" "test" {
  resource    = turbot_folder.test.id
  policy_pack = turbot_policy_pack.test.id
}
`
}

// helper functions
func testAccCheckPolicyPackAttachmentExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		client := testAccProvider.Meta().(*apiClient.Client)
		policyPackId, resource := parsePolicyPackId(rs.Primary.ID)
		_, err := client.ReadPolicyPack(policyPackId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}
func testAccCheckPolicyPackAttachmentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*apiClient.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "policyPack" {
			continue
		}
		_, err := client.ReadPolicyPack(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("alert still exists")
		}
		if !errors.NotFoundError(err) {
			return fmt.Errorf("expected 'not found' error, got %s", err)
		}
	}
	return nil
}
